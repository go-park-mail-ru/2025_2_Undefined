package usecase

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	"github.com/google/uuid"
)

type SessionRepository interface {
	AddSession(userID uuid.UUID, device string) (uuid.UUID, error)
	DeleteSession(sessionID uuid.UUID) error
	GetSession(sessionID uuid.UUID) (*session.Session, error)
	GetSessionsByUserID(userID uuid.UUID) ([]*session.Session, error)
	UpdateSession(sessionID uuid.UUID) error
}

type SessionUsecase struct {
	sessionrepo SessionRepository
}

func New(sessionrepo SessionRepository) *SessionUsecase {
	return &SessionUsecase{
		sessionrepo: sessionrepo,
	}
}

func (uc *SessionUsecase) GetSession(sessionID uuid.UUID) (*session.Session, error) {
	const op = "SessionUseCase.GetSession"
	if sessionID == uuid.Nil {
		err := errors.New("session required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	sess, err := uc.sessionrepo.GetSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	return sess, nil
}

func (uc *SessionUsecase) GetSessionsByUserID(userID uuid.UUID) ([]*session.Session, error) {
	const op = "SessionUseCase.GetSessionsByUserID"

	if userID == uuid.Nil {
		err := errors.New("user ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	sessions, err := uc.sessionrepo.GetSessionsByUserID(userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	return sessions, nil
}

func (uc *SessionUsecase) UpdateSession(sessionID uuid.UUID) error {
	const op = "SessionUsecase.UpdateSession"

	if sessionID == uuid.Nil {
		err := errors.New("session ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	err := uc.sessionrepo.UpdateSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}
