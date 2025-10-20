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
	GetSessionsByUserID(userID uuid.UUID) ([]*session.Session, error)
}

type SessionUtils struct {
	sessionRepo SessionRepository
}

func NewSessionUtils(sessionRepo SessionRepository) *SessionUtils {
	return &SessionUtils{
		sessionRepo: sessionRepo,
	}
}

// GetUserIDFromSession извлекает ID пользователя из сессии в cookie
func (s *SessionUtils) GetUserIDFromSession(r *http.Request) (uuid.UUID, error) {
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
	sessionInfo, err := s.sessionRepo.GetSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidToken)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, errors.New("invalid session")
	}

	return sessionInfo.UserID, nil
}

// GetSessionsByUserID получает все сессии пользователя
func (s *SessionUtils) GetSessionsByUserID(userID uuid.UUID) ([]*session.Session, error) {
	return s.sessionRepo.GetSessionsByUserID(userID)
}
