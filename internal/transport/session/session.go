package session

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	"github.com/google/uuid"
)

type SessionRepository interface {
	GetSession(sessionID uuid.UUID) (*session.Session, error)
}

// GetUserIDFromSession извлекает ID пользователя из сессии в cookie
func GetUserIDFromSession(r *http.Request, sessionRepo SessionRepository) (uuid.UUID, error) {
	const op = "session.GetUserIDFromSession"
	
	// Получаем сессию из куки
	sessionCookie, err := r.Cookie(domains.SessionName)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrJWTIsRequired)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, errors.New("session required")
	}

	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("invalid session ID"))
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, errors.New("invalid session ID")
	}

	// Получаем информацию о сессии
	sessionInfo, err := sessionRepo.GetSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidToken)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, errors.New("invalid session")
	}

	return sessionInfo.UserID, nil
}
