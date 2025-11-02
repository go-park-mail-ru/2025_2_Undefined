package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/csrf"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

func CSRFMiddleware(sessionConf *config.SessionConfig, csrfSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "CSRFMiddleware"

			// Пропускаем GET, HEAD, OPTIONS запросы
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			csrfToken := r.Header.Get("X-CSRF-Token")
			if csrfToken == "" {
				utils.SendError(r.Context(), op, w, http.StatusForbidden, "CSRF token required")
				return
			}

			cookie, err := r.Cookie(sessionConf.Signature)
			if err != nil {
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Session required for CSRF validation")
				return
			}

			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Invalid session ID")
				return
			}

			// Валидируем CSRF токен
			err = csrf.ValidateCSRFToken(csrfToken, sessionID.String(), csrfSecret)
			if err != nil {
				utils.SendError(r.Context(), op, w, http.StatusForbidden, "Invalid CSRF token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
