package dto

import (
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
)

type ChatViewInformationDTO struct {
	ID          uuid.UUID             `json:"id" swaggertype:"string" format:"uuid"`
	Name        string                `json:"name"`
	LastMessage dtoMessage.MessageDTO `json:"last_message"`
	Type        string                `json:"type"`
}

type ChatDetailedInformationDTO struct {
	ID        uuid.UUID               `json:"id" swaggertype:"string" format:"uuid"`
	Name      string                  `json:"name"`
	IsAdmin   bool                    `json:"is_admin"`
	CanChat   bool                    `json:"can_chat"`
	IsMember  bool                    `json:"is_member"`
	IsPrivate bool                    `json:"is_private"`
	Type      string                  `json:"type"`
	Messages  []dtoMessage.MessageDTO `json:"messages"`
	Members   []UserInfoChatDTO       `json:"members"`
}

type ChatCreateInformationDTO struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"` // Тип чата - канал, диалог или группа
	Members []UserInfoChatDTO `json:"members"`
}

type UserInfoChatDTO struct {
	UserId     uuid.UUID `json:"user_id" swaggertype:"string" format:"uuid"`
	UserName   string    `json:"user_name"`
	UserAvatar *string   `json:"user_avatar"`
	Role       string    `json:"role"` // Роль пользователя в чате - админ(писать и добавлять участников), участник(писать), зритель (только просмотр)
}

type IdDTO struct {
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
}
