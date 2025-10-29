package transport

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	UserDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type SessionUtilsI interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
	GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error)
}

type UserUsecase interface {
	GetUserById(ctx context.Context, id uuid.UUID) (*UserDto.User, error)
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
		cookie.Unset(w, "session_token")
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
