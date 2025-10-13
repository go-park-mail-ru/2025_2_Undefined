package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChatTypeChannel = "channel"
	ChatTypeDialog  = "dialog"
	ChatTypeGroup   = "group"
)

const (
	RoleAdmin  = "admin"
	RoleMember = "writer"
	RoleViewer = "viewer"
)

const (
	MessageTypeUser   = "user"
	MessageTypeSystem = "system"
)

type Chat struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Message struct {
	ID        uuid.UUID `json:"-"`
	ChatID    uuid.UUID `json:"-"`
	UserID    uuid.UUID `json:"sender"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	Type      string
}

type UserInfo struct {
	UserID uuid.UUID `json:"user_id"`
	ChatID uuid.UUID `json:"-"`
	Role   string    `json:"role"`
}
