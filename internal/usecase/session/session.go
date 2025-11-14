package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	"github.com/google/uuid"
)

type SessionRepository interface {
	AddSession(userID uuid.UUID, device string) (uuid.UUID, error)
	DeleteSession(sessionID uuid.UUID) error
	DeleteAllSessionWithoutCurrent(userID uuid.UUID, currentSessionID uuid.UUID) error
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

	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if sessionID == uuid.Nil {
		err := errors.New("session required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("session ID is required")
		return nil, wrappedErr
	}

	sess, err := uc.sessionrepo.GetSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to get session")
		return nil, wrappedErr
	}

	sessDTO := &dto.Session{
		ID:         sess.ID,
		UserID:     sess.UserID,
		Device:     sess.Device,
		Created_at: sess.Created_at,
		Last_seen:  sess.Last_seen,
	}

	return sessDTO, nil
}

func (uc *SessionUsecase) GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error) {
	const op = "SessionUseCase.GetSessionsByUserID"

	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if userID == uuid.Nil {
		err := errors.New("user ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("user ID is required")
		return nil, wrappedErr
	}

	sessions, err := uc.sessionrepo.GetSessionsByUserID(userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to get sessions by user ID")
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

	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if sessionID == uuid.Nil {
		err := errors.New("session ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("session ID is required")
		return wrappedErr
	}

	err := uc.sessionrepo.UpdateSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to update session")
		return wrappedErr
	}

	return nil
}

func (uc *SessionUsecase) DeleteSession(userID uuid.UUID, sessionID uuid.UUID) error {
	const op = "SessionUsecase.DeleteSession"

	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if userID == uuid.Nil {
		err := errors.New("user ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("user ID is required")
		return err
	}

	if sessionID == uuid.Nil {
		err := errors.New("session ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("session ID is required")
		return err
	}

	session, err := uc.sessionrepo.GetSession(sessionID)
	if err != nil {
		if errors.Is(err, errs.ErrSessionNotFound) {
			wrappedErr := fmt.Errorf("%s: session not found: %w", op, err)
			logger.WithError(wrappedErr).Error("session not found")
			return errs.ErrSessionNotFound
		}
		wrappedErr := fmt.Errorf("%s: failed to get session: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to get session")
		return wrappedErr
	}

	if session.UserID != userID {
		err := errors.New("session does not belong to user")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("session does not belong to user")
		return err
	}

	err = uc.sessionrepo.DeleteSession(sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to delete session")
		return err
	}

	return nil
}

func (uc *SessionUsecase) DeleteAllSessionWithoutCurrent(userID uuid.UUID, currentSessionID uuid.UUID) error {
	const op = "SessionUsecase.DeleteAllSessionWithoutCurrent"

	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if currentSessionID == uuid.Nil {
		err := errors.New("session ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("session ID is required")
		return err
	}

	if userID == uuid.Nil {
		err := errors.New("user ID is required")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("user ID is required")
		return err
	}

	err := uc.sessionrepo.DeleteAllSessionWithoutCurrent(userID, currentSessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to delete all sessions without current")
		return wrappedErr
	}

	return nil
}
