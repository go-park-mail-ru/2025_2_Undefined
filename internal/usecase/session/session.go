package usecase

import (
	"errors"
	"fmt"
	"log"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	"github.com/google/uuid"
)

type SessionRepository interface {
	AddSession(userID uuid.UUID, device string) (uuid.UUID, error)
	DeleteSession(sessionID uuid.UUID) error
	GetSession(sessionID uuid.UUID) (*models.Session, error)
	GetSessionsByUserID(userID uuid.UUID) ([]*models.Session, error)
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

func (uc *SessionUsecase) GetSession(sessionID uuid.UUID) (*dto.Session, error) {
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

	sessDTO := &dto.Session{
		ID: sess.ID,
		UserID: sess.UserID,
		Device: sess.Device,
		Created_at: sess.Created_at,
		Last_seen: sess.Last_seen,
	}

	return sessDTO, nil
}

func (uc *SessionUsecase) GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error) {
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

	sessionsDTO := make([]*dto.Session, 0, len(sessions))
	for _, sess := range sessions {
		sessDTO := &dto.Session{
			ID:         sess.ID,
			UserID:     sess.UserID,
			Device:     sess.Device,
			Created_at: sess.Created_at,
			Last_seen:  sess.Last_seen,
		}
		sessionsDTO = append(sessionsDTO, sessDTO)
	}

	return sessionsDTO, nil
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
