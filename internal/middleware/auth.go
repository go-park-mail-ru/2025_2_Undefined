package middleware

import (
	"context"
	"net/http"
	"strings"

	service "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/auth"
)

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "Invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			user, err := authService.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusUnauthorized)
				return
			}

			if user == nil {
				http.Error(w, `{"error": "User account not found"}`, http.StatusUnauthorized)
				return
			}

			// Добавляем пользователя в контекст
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
