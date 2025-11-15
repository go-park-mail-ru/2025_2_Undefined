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
	AvatarID    *uuid.UUID
}

type UserInfo struct {
	UserID       uuid.UUID
	ChatID       uuid.UUID
	UserName     string
	UserAvatarID *uuid.UUID
	Role         string
}
