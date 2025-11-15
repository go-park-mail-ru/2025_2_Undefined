package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	"github.com/google/uuid"
)

type SessionUsecase interface {
	GetSession(ctx context.Context, sessionID uuid.UUID) (*dto.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.Session, error)
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

	logger := domains.GetLogger(r.Context()).WithField("operation", op)

	sessionCookie, err := r.Cookie(s.sessionConfig.Signature)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrJWTIsRequired)
		logger.WithError(wrappedErr).Error("session cookie not found")
		return uuid.Nil, errors.New("session required")
	}

	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("invalid session ID"))
		logger.WithError(wrappedErr).Error("invalid session ID format")
		return uuid.Nil, errors.New("invalid session ID")
	}

	sessionInfo, err := s.uc.GetSession(r.Context(), sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidToken)
		logger.WithError(wrappedErr).Error("failed to get session info")
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
		return nil, wrappedErr
	}

	sessions, err := s.uc.GetSessionsByUserID(context.Background(), userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		return nil, wrappedErr
	}

	return sessions, nil
}

func (s *SessionUtils) GetSessionFromCookie(r *http.Request) (uuid.UUID, error) {
	const op = "SessionUtils.GetSessionFromCookie"

	logger := domains.GetLogger(r.Context()).WithField("operation", op)

	sessionCookie, err := r.Cookie(s.sessionConfig.Signature)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrJWTIsRequired)
		logger.WithError(wrappedErr).Error("session cookie not found")
		return uuid.Nil, errors.New("session required")
	}

	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("invalid session ID"))
		logger.WithError(wrappedErr).Error("invalid session ID format")
		return uuid.Nil, errors.New("invalid session ID")
	}

	return sessionID, nil
}
