package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Device     string
	Created_at time.Time
	Last_seen  time.Time
}
