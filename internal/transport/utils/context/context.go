package context

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
)

// GetUserIDFromContext извлекает ID пользователя из контекста запроса
func GetUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	const op = "context.GetUserIDFromContext"
	logger := domains.GetLogger(r.Context()).WithField("operation", op)

	userIDVal := r.Context().Value(domains.UserIDKey{})
	if userIDVal == nil {
		err := errors.New("user_id not found in context")
		logger.WithError(err).Error("missing user_id in request context")
		return uuid.Nil, err
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		err := errors.New("user_id has invalid type in context")
		logger.WithError(err).Error("user_id type assertion failed")
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.WithError(err).Error("failed to parse user_id from context")
		return uuid.Nil, errors.New("invalid user_id format")
	}

	return userID, nil
}

// GetSessionIDFromCookie извлекает session_id из cookie
func GetSessionIDFromCookie(r *http.Request, cookieName string) (uuid.UUID, error) {
	const op = "context.GetSessionIDFromCookie"
	logger := domains.GetLogger(r.Context()).WithField("operation", op)

	sessionCookie, err := r.Cookie(cookieName)
	if err != nil {
		logger.WithError(err).Error("session cookie not found")
		return uuid.Nil, errors.New("session not found")
	}

	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		logger.WithError(err).Error("invalid session ID format")
		return uuid.Nil, errors.New("invalid session ID")
	}

	return sessionID, nil
}
