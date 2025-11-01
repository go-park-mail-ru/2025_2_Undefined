package session

import (
	"net/http"

	"github.com/google/uuid"
)

type SessionUsecase interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
}
