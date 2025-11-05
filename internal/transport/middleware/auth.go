package middleware

import (
	"log" // Добавлен импорт log
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
			log.Printf("AuthMiddleware: Processing request to %s %s", r.Method, r.URL.Path) // Логирование начала обработки

			// Получаем сессию из куки
			cookie, err := r.Cookie(sessionConf.Signature)
			if err != nil {
				log.Printf("AuthMiddleware: No session cookie '%s' found for %s %s", sessionConf.Signature, r.Method, r.URL.Path) // Логирование отсутствия куки
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Session required")
				return
			}
			log.Printf("AuthMiddleware: Found session cookie for %s %s, value: %s", r.Method, r.URL.Path, cookie.Value) // Логирование значения куки

			// Парсим UUID сессии
			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				log.Printf("AuthMiddleware: Invalid session ID format in cookie for %s %s, error: %v", r.Method, r.URL.Path, err) // Логирование ошибки парсинга
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "Invalid session ID")
				return
			}

			sess, err := sessionUC.GetSession(sessionID)
			if err != nil {
				log.Printf("AuthMiddleware: Session %s not found or invalid for %s %s, error: %v", sessionID, r.Method, r.URL.Path, err) // Логирование ошибки получения сессии
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}
			log.Printf("AuthMiddleware: Session %s found for %s %s, user ID: %s", sessionID, r.Method, r.URL.Path, sess.UserID) // Логирование успешного получения сессии

			// Проверяем истекла ли сессия
			if time.Since(sess.Last_seen) > sessionConf.LifeSpan {
				log.Printf("AuthMiddleware: Session %s for user %s expired for %s %s", sessionID, sess.UserID, r.Method, r.URL.Path) // Логирование истечения срока
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "session expired")
				return
			}
			log.Printf("AuthMiddleware: Session %s for user %s is valid for %s %s", sessionID, sess.UserID, r.Method, r.URL.Path) // Логирование валидности сессии

			// Обновляем сессию
			err = sessionUC.UpdateSession(sessionID)
			if err != nil {
				log.Printf("AuthMiddleware: Failed to update session %s for user %s for %s %s, error: %v", sessionID, sess.UserID, r.Method, r.URL.Path, err) // Логирование ошибки обновления
				cookieUtils.Unset(w, sessionConf.Signature)
				utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrInvalidToken.Error())
				return
			}
			log.Printf("AuthMiddleware: Session %s for user %s updated successfully for %s %s", sessionID, sess.UserID, r.Method, r.URL.Path) // Логирование успешного обновления

			// Передаем запрос дальше
			log.Printf("AuthMiddleware: User %s authorized for %s %s, passing to next handler", sess.UserID, r.Method, r.URL.Path) // Логирование передачи дальше
			next.ServeHTTP(w, r)
		})
	}
}