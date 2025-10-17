package models

import (
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

type Chat struct {
	ID          uuid.UUID
	Type        string
	Name        string
	Description string
}

type UserInfo struct {
	UserID     uuid.UUID
	ChatID     uuid.UUID
	UserName   string
	UserAvatar *string
	Role       string
}
