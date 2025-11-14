package middleware

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	cookieUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	SessionUC "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/session"
	"github.com/google/uuid"
)

// AuthMiddleware создает middleware для проверки аутентификации через сессии
func AuthMiddleware(sessionConf *config.SessionConfig, sessionUC *SessionUC.SessionUsecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "AuthMiddleware"
			cookie, err := r.Cookie(sessionConf.Signature)
			if err != nil {
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Session required")
				return
			}

			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Invalid session ID")
				return
			}

			sess, err := sessionUC.GetSession(r.Context(), sessionID)
			if err != nil {
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}

			// Проверяем истекла ли сессия
			if time.Since(sess.Last_seen) > sessionConf.LifeSpan {
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "session expired")
				return
			}

			err = sessionUC.UpdateSession(r.Context(), sessionID)
			if err != nil {
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
