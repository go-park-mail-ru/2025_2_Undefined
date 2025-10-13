package dto

import (
	"time"

	"github.com/google/uuid"
)

type MessageDTO struct {
	Sender    uuid.UUID `json:"sender" swaggertype:"string" format:"uuid"`
	Text      string    `json:"string"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatViewInformationDTO struct {
	ID          uuid.UUID  `json:"id" swaggertype:"string" format:"uuid"`
	Name        string     `json:"name"`
	LastMessage MessageDTO `json:"last_message"`
}

type ChatDetailedInformationDTO struct {
	ID        uuid.UUID         `json:"id" swaggertype:"string" format:"uuid"`
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
	Type    string            `json:"type"` // Тип чата - канал, диалог или группа
	Members []UserInfoChatDTO `json:"members"`
}

type UserInfoChatDTO struct {
	UserId uuid.UUID `json:"user_id" swaggertype:"string" format:"uuid"`
	Role   string    `json:"role"` // Роль пользователя в чате - админ(писать и добавлять участников), участник(писать), зритель (только просмотр)
}

type IdDTO struct {
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
}
