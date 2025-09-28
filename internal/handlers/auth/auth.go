package handlers

import (
	"encoding/json"
	"net/http"

	_ "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/dto"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/utils/response"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/utils/validation"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	service "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/auth"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register регистрирует нового пользователя
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя в системе и возвращает JWT токен в cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body  models.RegisterRequest  true  "Данные для регистрации"
// @Success      201   "Пользователь успешно зарегистрирован"
// @Failure      400   {object}  dto.ErrorDTO  "Некорректный запрос"
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthModels.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	if req.PhoneNumber == "" || req.Email == "" || req.Username == "" || req.Password == "" || req.Name == "" {
		utils.SendError(w, http.StatusBadRequest, "All fields are required")
		return
	}
	normalizedPhone, isValid := validation.ValidateAndNormalizePhone(req.PhoneNumber)
	if !isValid {
		utils.SendError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}
	req.PhoneNumber = normalizedPhone

	if !validation.ValidateEmail(req.Email) {
		utils.SendError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	if !validation.ValidatePassword(req.Password) {
		utils.SendError(w, http.StatusBadRequest, "Only Latin letters, numbers and special characters are allowed in the password")
		return
	}

	if !validation.ValidateUsername(req.Username) {
		utils.SendError(w, http.StatusBadRequest, "Invalid username format")
		return
	}

	if !validation.ValidateName(req.Name) {
		utils.SendError(w, http.StatusBadRequest, "Invalid name format")
		return
	}

	token, err := h.authService.Register(&req)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if token == "" {
		utils.SendError(w, http.StatusBadRequest, "token is empty")
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
// @Failure      400  {object}  dto.ErrorDTO  "Некорректный запрос"
// @Failure      401  {object}  dto.ErrorDTO  "Неверные креденшиалы"
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthModels.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	if req.PhoneNumber == "" || req.Password == "" {
		utils.SendError(w, http.StatusBadRequest, errs.ErrRequiredFieldsMissing.Error())
		return
	}

	normalizedPhone, isValid := validation.ValidateAndNormalizePhone(req.PhoneNumber)
	if !isValid {
		utils.SendError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}
	req.PhoneNumber = normalizedPhone

	if !validation.ValidatePassword(req.Password) {
		utils.SendError(w, http.StatusBadRequest, "Only Latin letters, numbers and special characters are allowed in the password")
		return
	}

	token, err := h.authService.Login(&req)
	if err != nil {
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
	jwtCookie, err := r.Cookie(domains.TokenCookieName)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "JWT token required")
		return
	}
	err = h.authService.Logout(jwtCookie.Value)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	cookie.Unset(w, domains.TokenCookieName)
	utils.SendJSONResponse(w, http.StatusUnauthorized, nil)
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
	jwtCookie, err := r.Cookie(domains.TokenCookieName)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "JWT token required")
		return
	}
	jwttoken := jwt.NewTokenator()
	claims, err := jwttoken.ParseJWT(jwtCookie.Value)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}
	user, err := h.authService.GetUserById(userID)
	if err != nil {
		cookie.Unset(w, domains.TokenCookieName)
		utils.SendError(w, http.StatusUnauthorized, errs.ErrUserNotFound.Error())
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, user)
}
