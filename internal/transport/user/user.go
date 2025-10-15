package transport

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	sessionUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/session"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type UserUsecase interface {
	GetUserById(id uuid.UUID) (*UserModels.User, error)
}

type UserHandler struct {
	uc          UserUsecase
	sessionRepo sessionUtils.SessionRepository
}

func New(uc UserUsecase, sessionRepo sessionUtils.SessionRepository) *UserHandler {
	return &UserHandler{
		uc:          uc,
		sessionRepo: sessionRepo,
	}
}

// GetCurrentUser получает информацию о текущем пользователе
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает полные данные о текущем авторизованном пользователе
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  models.User   "Информация о пользователе"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Router       /me [get]
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetCurrentUser"

	userID, err := sessionUtils.GetUserIDFromSession(r, h.sessionRepo)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := h.uc.GetUserById(userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
		log.Printf("Error: %v", wrappedErr)
		cookie.Unset(w, "session_token")
		utils.SendError(w, http.StatusUnauthorized, errs.ErrUserNotFound.Error())
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, user)
}
