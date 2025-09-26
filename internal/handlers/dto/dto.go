package dto

import (
	"time"

	"github.com/google/uuid"
)

type ErrorDTO struct {
	Message string `json:"message"`
}

type MessageDTO struct {
	Sender    uuid.UUID `json:"sender"`
	Text      string    `json:"string"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatViewInformationDTO struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	LastMessage MessageDTO `json:"last_message"`
}

type ChatDetailedInformationDTO struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	IsAdmin   bool              `json:"is_admin"`
	CanChat   bool              `json:"can_chat"`
	IsMember  bool              `json:"is_member"`
	IsPrivate bool              `json:"is_private"`
	Messages  []MessageDTO      `json:"messages"`
	Members   []UserInfoChatDTO `json:"members"`
}

type ChatCreateInformationDTO struct {
	Name    string            `json:"name"`
	Type    int               `json:"type"`
	Members []UserInfoChatDTO `json:"members"`
}

type UserInfoChatDTO struct {
	UserId uuid.UUID `json:"user_id"`
	Role   int       `json:"role"`
}

type IdDTO struct {
	ID uuid.UUID `json:"id"`
}
