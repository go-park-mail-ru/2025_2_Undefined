package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	cookieUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
)

func AuthGRPCMiddleware(sessionConf *config.SessionConfig, authClient gen.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "AuthGRPCMiddleware"

			cookie, err := r.Cookie(sessionConf.Signature)
			if err != nil {
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Session required")
				return
			}

			// Валидируем сессию через gRPC
			res, err := authClient.ValidateSession(r.Context(), &gen.ValidateSessionReq{
				SessionId: cookie.Value,
			})
			if err != nil || !res.Valid {
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Invalid or expired session")
				return
			}

			// Добавляем user_id в контекст
			ctx := context.WithValue(r.Context(), domains.UserIDKey{}, res.UserId)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
