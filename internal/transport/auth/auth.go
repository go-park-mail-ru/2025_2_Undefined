package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	sessionUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/session"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
)

type AuthUsecase interface {
	Register(req *AuthModels.RegisterRequest, device string) (uuid.UUID, *dto.ValidationErrorsDTO)
	Login(req *AuthModels.LoginRequest, device string) (uuid.UUID, error)
	Logout(SessionID uuid.UUID) error
}

type AuthHandler struct {
	uc          AuthUsecase
	sessionRepo sessionUtils.SessionRepository
}

func New(uc AuthUsecase, sessionRepo sessionUtils.SessionRepository) *AuthHandler {
	return &AuthHandler{
		uc:          uc,
		sessionRepo: sessionRepo,
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

	device := fmt.Sprintf("%s %s on %s", name, version, os)
	if device == "  on " {
		return "Unknown Device"
	}

	return device
}

// Register регистрирует нового пользователя
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя в системе и создает сессию
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body  dto.RegisterRequest  true  "Данные для регистрации"
// @Success      201   "Пользователь успешно зарегистрирован"
// @Failure      400   {object}  dto.ValidationErrorsDTO  "Ошибки валидации"
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Register"
	var req AuthModels.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrBadRequest)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// Валидируем все поля и собираем все ошибки
	validationErrors := validation.ValidateRegisterRequest(&req)
	if len(validationErrors) > 0 {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("validationErrors"))
		log.Printf("Error: %v", wrappedErr)
		validationDTO := validation.ConvertToValidationErrorsDTO(validationErrors)
		utils.SendValidationErrors(w, http.StatusBadRequest, validationDTO)
		return
	}

	// Получаем информацию об устройстве из User-Agent
	device := getDeviceFromUserAgent(r)

	sessionID, validationErr := h.uc.Register(&req, device)
	if validationErr != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("registration error"))
		log.Printf("Error: %v", wrappedErr)
		utils.SendValidationErrors(w, http.StatusBadRequest, *validationErr)
		return
	}

	if sessionID == uuid.Nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("session ID is missing"))
		log.Printf("Error: %v", wrappedErr)
		utils.SendValidationErrors(w, http.StatusBadRequest, dto.ValidationErrorsDTO{
			Message: "session ID is missing",
		})
		return
	}

	cookie.Set(w, sessionID.String(), domains.SessionName)
	w.WriteHeader(http.StatusCreated)
}

// Login аутентифицирует пользователя
// @Summary      Аутентификация пользователя
// @Description  Аутентифицирует пользователя по номеру телефона и паролю, создает сессию
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body  dto.LoginRequest  true  "Креденшиалы для входа"
// @Success      200  "Вход выполнен успешно"
// @Failure      400  {object}  dto.ValidationErrorsDTO  "Ошибки валидации"
// @Failure      401  {object}  dto.ErrorDTO  "Неверные креденшиалы"
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Login"
	var req AuthModels.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrBadRequest)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// Валидируем все поля и собираем все ошибки
	validationErrors := validation.ValidateLoginRequest(&req)
	if len(validationErrors) > 0 {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("validationErrors"))
		log.Printf("Error: %v", wrappedErr)
		validationDTO := validation.ConvertToValidationErrorsDTO(validationErrors)
		utils.SendValidationErrors(w, http.StatusBadRequest, validationDTO)
		return
	}

	// Получаем информацию об устройстве из User-Agent
	device := getDeviceFromUserAgent(r)

	sessionID, err := h.uc.Login(&req, device)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrInvalidCredentials.Error())
		return
	}

	cookie.Set(w, sessionID.String(), domains.SessionName)
	w.WriteHeader(http.StatusOK)
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
	_, err := sessionUtils.GetUserIDFromSession(r, h.sessionRepo)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("invalid session"))
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Получаем ID сессии из cookie для удаления
	sessionCookie, err := r.Cookie(domains.SessionName)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrJWTIsRequired)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrJWTIsRequired.Error())
		return
	}

	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("invalid session ID"))
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, "invalid session ID")
		return
	}

	err = h.uc.Logout(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	cookie.Unset(w, domains.SessionName)
	utils.SendJSONResponse(w, http.StatusOK, nil)
}
