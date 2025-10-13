package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
)

type AuthUsecase interface {
	Register(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO)
	Login(req *AuthModels.LoginRequest) (string, error)
	Logout(tokenString string) error
	GetUserById(id uuid.UUID) (*UserModels.User, error)
}

type AuthHandler struct {
	uc AuthUsecase
}

func New(uc AuthUsecase) *AuthHandler {
	return &AuthHandler{
		uc: uc,
	}
}

// Register регистрирует нового пользователя
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя в системе и возвращает JWT токен в cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body  models.RegisterRequest  true  "Данные для регистрации"
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

	token, err := h.uc.Register(&req)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("registration error"))
		log.Printf("Error: %v", wrappedErr)
		utils.SendValidationErrors(w, http.StatusBadRequest, *err)
		return
	}

	if token == "" {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("token is missing"))
		log.Printf("Error: %v", wrappedErr)
		utils.SendValidationErrors(w, http.StatusBadRequest, dto.ValidationErrorsDTO{
			Message: "token is missing",
		})
		return
	}

	cookie.Set(w, token, domains.TokenCookieName)
	w.WriteHeader(http.StatusCreated)
}

// Login аутентифицирует пользователя
// @Summary      Аутентификация пользователя
// @Description  Аутентифицирует пользователя по номеру телефона и паролю, возвращает JWT токен в cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body  models.LoginRequest  true  "Креденшиалы для входа"
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

	token, err := h.uc.Login(&req)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrInvalidCredentials.Error())
		return
	}

	cookie.Set(w, token, domains.TokenCookieName)
	w.WriteHeader(http.StatusOK)
}

// Logout завершает сессию пользователя
// @Summary      Выход из системы
// @Description  Аннулирует текущий JWT токен и удаляет cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  "Logout successful"
// @Failure      401  {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Router       /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Logout"
	jwtCookie, err := r.Cookie(domains.TokenCookieName)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrJWTIsRequired)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrJWTIsRequired.Error())
		return
	}
	err = h.uc.Logout(jwtCookie.Value)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	cookie.Unset(w, domains.TokenCookieName)
	utils.SendJSONResponse(w, http.StatusOK, nil)
}

// GetCurrentUser получает информацию о текущем пользователе
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает полные данные о текущем авторизованном пользователе
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  models.User   "Информация о пользователе"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Router       /me [get]
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.GetCurrentUser"
	jwtCookie, err := r.Cookie(domains.TokenCookieName)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrJWTIsRequired)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrJWTIsRequired.Error())
		return
	}
	jwttoken := jwt.NewTokenator()
	claims, err := jwttoken.ParseJWT(jwtCookie.Value)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidToken)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		err = errors.New("Invalid user ID format")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user, err := h.uc.GetUserById(userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
		log.Printf("Error: %v", wrappedErr)
		cookie.Unset(w, domains.TokenCookieName)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrUserNotFound.Error())
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, user)
}
