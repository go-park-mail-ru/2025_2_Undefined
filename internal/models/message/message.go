package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	MessageTypeUser   = "user"
	MessageTypeSystem = "system"
)

type Message struct {
	ID           uuid.UUID
	ChatID       uuid.UUID
	UserID       *uuid.UUID
	UserName     string
	UserAvatarID *uuid.UUID
	Text         string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	Type         string
}

type CreateMessage struct {
	ChatID    uuid.UUID
	UserID    *uuid.UUID
	Text      string
	CreatedAt time.Time
	Type      string
}
