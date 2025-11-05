package session

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	"github.com/google/uuid"
)

type SessionUsecase interface {
	GetSession(sessionID uuid.UUID) (*dto.Session, error)
	GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error)
}

type SessionUtils struct {
	uc            SessionUsecase
	sessionConfig *config.SessionConfig
}

func NewSessionUtils(uc SessionUsecase, sessionConfig *config.SessionConfig) *SessionUtils {
	return &SessionUtils{
		uc:            uc,
		sessionConfig: sessionConfig,
	}
}

// GetUserIDFromSession извлекает ID пользователя из сессии в cookie
func (s *SessionUtils) GetUserIDFromSession(r *http.Request) (uuid.UUID, error) {
	const op = "SessionUtils.GetUserIDFromSession"

	// Получаем сессию из куки
	sessionCookie, err := r.Cookie(s.sessionConfig.Signature)
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
	sessionInfo, err := s.uc.GetSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidToken)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, errors.New("invalid session")
	}

	return sessionInfo.UserID, nil
}

// GetSessionsByUserID получает все сессии пользователя
func (s *SessionUtils) GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error) {
	const op = "SessionUtils.GetSessionsByUserID"

	if userID == uuid.Nil {
		err := errors.New("user ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	sessions, err := s.uc.GetSessionsByUserID(userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	return sessions, nil
}

func (s *SessionUtils) GetSessionFromCookie(r *http.Request) (uuid.UUID, error) {
	const op = "SessionUtils.GetSessionFromCookie"

	sessionCookie, err := r.Cookie(s.sessionConfig.Signature)
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

	return sessionID, nil
}