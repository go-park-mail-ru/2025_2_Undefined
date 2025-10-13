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
	Type    int               `json:"type" enums:"0,1,2"` // Тип чата - 0 канал, 1 диалог, 2 группа
	Members []UserInfoChatDTO `json:"members"`
}

type UserInfoChatDTO struct {
	UserId uuid.UUID `json:"user_id" swaggertype:"string" format:"uuid"`
	Role   int       `json:"role" enums:"0,1,2"` // Роль пользователя в чате - 0 админ(писать и добавлять участников), 1 участник(писать), 2 зритель (только просмотр)
}

type IdDTO struct {
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
}
