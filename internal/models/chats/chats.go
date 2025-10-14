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
	ID          uuid.UUID
	Type        string
	Name        string
	Description string
}

type Message struct {
	ID         uuid.UUID
	ChatID     uuid.UUID
	UserID     uuid.UUID
	UserName   string
	UserAvatar *string
	Text       string
	CreatedAt  time.Time
	Type       string
}

type UserInfo struct {
	UserID     uuid.UUID
	ChatID     uuid.UUID
	UserName   string
	UserAvatar *string
	Role       string
}
