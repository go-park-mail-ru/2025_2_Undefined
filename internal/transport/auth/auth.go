package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/csrf"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
)

type SessionUtilsI interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, req *AuthDTO.RegisterRequest, device string) (uuid.UUID, *dto.ValidationErrorsDTO)
	Login(ctx context.Context, req *AuthDTO.LoginRequest, device string) (uuid.UUID, error)
	Logout(ctx context.Context, SessionID uuid.UUID) error
}

type AuthHandler struct {
	uc            AuthUsecase
	sessionConfig *config.SessionConfig
	csrfConfig    *config.CSRFConfig
	sessionUtils  SessionUtilsI
}

func New(uc AuthUsecase, sessionConfig *config.SessionConfig, csrfConfig *config.CSRFConfig, sessionUtils SessionUtilsI) *AuthHandler {
	return &AuthHandler{
		uc:            uc,
		sessionConfig: sessionConfig,
		csrfConfig:    csrfConfig,
		sessionUtils:  sessionUtils,
	}
}

// getDeviceFromUserAgent извлекает информацию об устройстве из User-Agent заголовка
func getDeviceFromUserAgent(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return "Unknown Device"
	}

	ua := user_agent.New(userAgent)
	name, version := ua.Browser()
	os := ua.OS()

	if name == "" && version == "" {
		return "Unknown Device"
	}
	if os == "" {
		return fmt.Sprintf("%s %s", name, version)
	}

	return fmt.Sprintf("%s %s on %s", name, version, os)
}

// Register регистрирует нового пользователя
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя в системе и создает сессию
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body  dto.RegisterRequest  true  "Данные для регистрации"
// @Success      201   {object}  dto.AuthResponse  "Пользователь успешно зарегистрирован"
// @Failure      400   {object}  dto.ValidationErrorsDTO  "Ошибки валидации"
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Register"
	var req AuthDTO.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// Валидируем все поля и собираем все ошибки
	validationErrors := validation.ValidateRegisterRequest(&req)
	if len(validationErrors) > 0 {
		validationDTO := validation.ConvertToValidationErrorsDTO(validationErrors)
		utils.SendValidationErrors(r.Context(), op, w, http.StatusBadRequest, validationDTO)
		return
	}

	// Получаем информацию об устройстве из User-Agent
	device := getDeviceFromUserAgent(r)

	sessionID, validationErr := h.uc.Register(r.Context(), &req, device)
	if validationErr != nil {
		utils.SendValidationErrors(r.Context(), op, w, http.StatusBadRequest, *validationErr)
		return
	}

	if sessionID == uuid.Nil {
		utils.SendValidationErrors(r.Context(), op, w, http.StatusBadRequest, dto.ValidationErrorsDTO{
			Message: "session ID is missing",
		})
		return
	}

	// Генерируем CSRF токен
	csrfToken := csrf.GenerateCSRFToken(sessionID.String(), h.csrfConfig.Secret)

	cookie.Set(w, sessionID.String(), h.sessionConfig.Signature)

	response := AuthDTO.AuthResponse{
		CSRFToken: csrfToken,
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusCreated, response)
}

// Login аутентифицирует пользователя
// @Summary      Аутентификация пользователя
// @Description  Аутентифицирует пользователя по номеру телефона и паролю, создает сессию
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body  dto.LoginRequest  true  "Креденшиалы для входа"
// @Success      200  {object}  dto.AuthResponse  "Вход выполнен успешно"
// @Failure      400  {object}  dto.ValidationErrorsDTO  "Ошибки валидации"
// @Failure      401  {object}  dto.ErrorDTO  "Неверные креденшиалы"
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Login"
	var req AuthDTO.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// Валидируем все поля и собираем все ошибки
	validationErrors := validation.ValidateLoginRequest(&req)
	if len(validationErrors) > 0 {
		validationDTO := validation.ConvertToValidationErrorsDTO(validationErrors)
		utils.SendValidationErrors(r.Context(), op, w, http.StatusBadRequest, validationDTO)
		return
	}

	// Получаем информацию об устройстве из User-Agent
	device := getDeviceFromUserAgent(r)

	sessionID, err := h.uc.Login(r.Context(), &req, device)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrInvalidCredentials.Error())
		return
	}

	// Генерируем CSRF токен
	csrfToken := csrf.GenerateCSRFToken(sessionID.String(), h.csrfConfig.Secret)

	cookie.Set(w, sessionID.String(), h.sessionConfig.Signature)

	response := AuthDTO.AuthResponse{
		CSRFToken: csrfToken,
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, response)
}

// Logout завершает сессию пользователя
// @Summary      Выход из системы
// @Description  Аннулирует текущую сессию и удаляет cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  "Logout successful"
// @Failure      401  {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Router       /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Logout"
	_, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Получаем ID сессии из cookie для удаления
	sessionCookie, err := r.Cookie(h.sessionConfig.Signature)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrJWTIsRequired.Error())
		return
	}

	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "invalid session ID")
		return
	}

	err = h.uc.Logout(r.Context(), sessionID)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}
	cookie.Unset(w, h.sessionConfig.Signature)
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}
