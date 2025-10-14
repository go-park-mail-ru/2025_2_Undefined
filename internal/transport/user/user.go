package transport

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type UserUsecase interface {
	GetUserById(id uuid.UUID) (*UserModels.User, error)
}

type UserHandler struct {
	uc UserUsecase
}

func New(uc UserUsecase) *UserHandler {
	return &UserHandler{
		uc: uc,
	}
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
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
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
