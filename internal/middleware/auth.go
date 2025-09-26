package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	cookieUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/utils/response"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	BlackToken "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/token"
)

// Ключи для хранения в контексте
type contextKey string

// AuthMiddleware создает middleware для проверки аутентификации
func AuthMiddleware(tokenator *jwt.Tokenator, blacktokenRepo BlackToken.TokenRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен из куки
			cookie, err := r.Cookie(domains.TokenCookieName)
			if err != nil {
				cookieUtils.Unset(w, domains.TokenCookieName)
				utils.SendError(w, http.StatusUnauthorized, "JWT token required")
				return
			}

			// Проверяем, не в черном списке ли токен
			if blacktokenRepo.IsInBlacklist(cookie.Value) {
				cookieUtils.Unset(w, domains.TokenCookieName)
				utils.SendError(w, http.StatusUnauthorized, "Token is blacklisted")
				return
			}

			// Парсим токен
			_, err = tokenator.ParseJWT(cookie.Value)
			if err != nil {
				cookieUtils.Unset(w, domains.TokenCookieName)
				utils.SendError(w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}

			// Передаем запрос дальше
			next.ServeHTTP(w, r)
		})
	}
}
