package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	SessionRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/session"
	cookieUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

// AuthMiddleware создает middleware для проверки аутентификации через сессии
func AuthMiddleware(sessionRepo *SessionRepo.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "AuthMiddleware"
			// Получаем сессию из куки
			cookie, err := r.Cookie(domains.SessionName)
			if err != nil {
				wrappedErr := fmt.Errorf("%s: %w", op, errors.New("Session required"))
				log.Printf("Error: %v", wrappedErr)
				cookieUtils.Unset(w, domains.SessionName)
				utils.SendError(w, http.StatusUnauthorized, "Session required")
				return
			}

			// Парсим UUID сессии
			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				wrappedErr := fmt.Errorf("%s: %w", op, errors.New("Invalid session ID"))
				log.Printf("Error: %v", wrappedErr)
				cookieUtils.Unset(w, domains.SessionName)
				utils.SendError(w, http.StatusUnauthorized, "Invalid session ID")
				return
			}

			// Проверяем существование сессии и обновляем время последней активности
			err = sessionRepo.UpdateSession(sessionID)
			if err != nil {
				wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidToken)
				log.Printf("Error: %v", wrappedErr)
				cookieUtils.Unset(w, domains.SessionName)
				utils.SendError(w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}

			// Передаем запрос дальше
			next.ServeHTTP(w, r)
		})
	}
}
