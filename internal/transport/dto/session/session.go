package dto

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"-"`
	Device     string    `json:"device"`
	Created_at time.Time `json:"created_at"`
	Last_seen  time.Time `json:"last_seen"`
}

type DeleteSession struct {
	ID uuid.UUID `json:"id"`
}
