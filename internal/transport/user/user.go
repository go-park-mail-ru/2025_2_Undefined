package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	UserDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type SessionUtilsI interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
	GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error)
}

type UserUsecase interface {
	GetUserById(ctx context.Context, id uuid.UUID) (*UserDto.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*UserDto.User, error)
	GetUserByUsername(ctx context.Context, username string) (*UserDto.User, error)
}

type UserHandler struct {
	uc           UserUsecase
	sessionUtils SessionUtilsI
}

func New(uc UserUsecase, sessionUtils SessionUtilsI) *UserHandler {
	return &UserHandler{
		uc:           uc,
		sessionUtils: sessionUtils,
	}
}

// GetCurrentUser получает информацию о текущем пользователе
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает полные данные о текущем авторизованном пользователе
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  dto.User   "Информация о пользователе"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Router       /me [get]
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetCurrentUser"

	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := h.uc.GetUserById(r.Context(), userID)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrUserNotFound.Error())
		return
	}
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// GetSessionsByUser получает все сессии текущего пользователя
// @Summary      Получить список сессий пользователя
// @Description  Возвращает все активные сессии текущего авторизованного пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   dto.Session  "Список сессий пользователя"
// @Failure      401  {object}  dto.ErrorDTO     "Неавторизованный доступ"
// @Failure      500  {object}  dto.ErrorDTO     "Внутренняя ошибка сервера"
// @Router       /sessions [get]
func (h *UserHandler) GetSessionsByUser(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetSessionsByUser"

	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	sessions, err := h.sessionUtils.GetSessionsByUserID(userID)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, sessions)
}

// GetUserByPhone получает информацию о пользователе по номеру телефона
// @Summary      Получить информацию о пользователе по номеру телефона
// @Description  Возвращает полные данные о пользователе по указанному номеру телефона
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request body dto.GetUserByPhone true "Запрос с номером телефона"
// @Success      200    {object}  dto.User   "Информация о пользователе"
// @Failure      400    {object}  dto.ErrorDTO   "Неверный формат номера телефона"
// @Failure      404    {object}  dto.ErrorDTO   "Пользователь не найден"
// @Failure      500    {object}  dto.ErrorDTO   "Внутренняя ошибка сервера"
// @Router       /user/by-phone [post]
func (h *UserHandler) GetUserByPhone(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetUserByPhone"

	var req UserDto.GetUserByPhone
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PhoneNumber == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Phone number is required")
		return
	}

	user, err := h.uc.GetUserByPhone(r.Context(), req.PhoneNumber)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			utils.SendError(r.Context(), op, w, http.StatusNotFound, err.Error())
			return
		}
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// GetUserByUsername получает информацию о пользователе по имени пользователя
// @Summary      Получить информацию о пользователе по имени пользователя
// @Description  Возвращает полные данные о пользователе по указанному имени пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request body dto.GetUserByUsername true "Запрос с именем пользователя"
// @Success      200       {object}  dto.User   "Информация о пользователе"
// @Failure      400       {object}  dto.ErrorDTO   "Неверный формат имени пользователя"
// @Failure      404       {object}  dto.ErrorDTO   "Пользователь не найден"
// @Failure      500       {object}  dto.ErrorDTO   "Внутренняя ошибка сервера"
// @Router       /user/by-username [post]
func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetUserByUsername"

	var req UserDto.GetUserByUsername
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Username is required")
		return
	}

	user, err := h.uc.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			utils.SendError(r.Context(), op, w, http.StatusNotFound, err.Error())
			return
		}
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}
