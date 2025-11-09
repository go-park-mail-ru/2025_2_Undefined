package session

import (
	"net/http"

	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	"github.com/google/uuid"
)

type SessionUsecase interface {
	GetSession(sessionID uuid.UUID) (*dto.Session, error)
	GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error)
	UpdateSession(sessionID uuid.UUID) error
	DeleteSession(userID uuid.UUID, sessionID uuid.UUID) error
	DeleteAllSessionWithoutCurrent(userID uuid.UUID, currentSessionID uuid.UUID) error
}

type SessionUtils interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
	GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error)
	GetSessionFromCookie(r *http.Request) (uuid.UUID, error)
}
