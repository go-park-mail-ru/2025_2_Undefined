package models

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	UserID        uuid.UUID
	ContactUserID uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
