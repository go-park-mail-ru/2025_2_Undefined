package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChatChannel int = iota
	ChatDialog
	ChatGroup
)

const (
	RoleAdmin int = iota
	RoleMember
	RoleViewer
)

const (
	UserMessage int = iota
	SystemMessage
)

type Chat struct {
	ID          uuid.UUID `json:"id"`
	Type        int       `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Message struct {
	ID        uuid.UUID `json:"-"`
	ChatID    uuid.UUID `json:"-"`
	UserID    uuid.UUID `json:"sender"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	Type      int
}

type UserInfo struct {
	UserID uuid.UUID `json:"user_id"`
	ChatID uuid.UUID `json:"-"`
	Role   int       `json:"role"`
}
