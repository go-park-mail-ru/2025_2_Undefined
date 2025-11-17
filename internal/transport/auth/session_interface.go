package auth

import (
	"context"

	sessionDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	"github.com/google/uuid"
)

//go:generate mockgen -source=session_interface.go -destination=../../usecase/mocks/mock_session_usecase_mock.go -package=mocks ISessionUsecase
type ISessionUsecase interface {
	GetSession(ctx context.Context, sessionID uuid.UUID) (*sessionDTO.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*sessionDTO.Session, error)
	UpdateSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	DeleteAllSessionWithoutCurrent(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error
}
