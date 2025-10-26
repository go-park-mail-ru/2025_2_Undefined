package middleware

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	cookieUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	SessionUC "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/session"
	"github.com/google/uuid"
)

// AuthMiddleware создает middleware для проверки аутентификации через сессии
func AuthMiddleware(conf *config.Config, sessionUC *SessionUC.SessionUsecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "AuthMiddleware"
			// Получаем сессию из куки
			cookie, err := r.Cookie(conf.SessionConfig.Signature)
			if err != nil {
				cookieUtils.Unset(w, domains.SessionName)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Session required")
				return
			}

			// Парсим UUID сессии
			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				cookieUtils.Unset(w, domains.SessionName)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Invalid session ID")
				return
			}
			sess, err := sessionUC.GetSession(sessionID)
			if err != nil {
				cookieUtils.Unset(w, conf.SessionConfig.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}

			// Проверяем истекла ли сессия
			if time.Since(sess.Last_seen) > conf.SessionConfig.LifeSpan {
				cookieUtils.Unset(w, conf.SessionConfig.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "session expired")
				return
			}
			// Обновляем сессию
			err = sessionUC.UpdateSession(sessionID)
			if err != nil {
				cookieUtils.Unset(w, domains.SessionName)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}

			// Передаем запрос дальше
			next.ServeHTTP(w, r)
		})
	}
}
