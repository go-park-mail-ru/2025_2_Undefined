package message

import (
	"time"

	"github.com/google/uuid"
)

type MessageDTO struct {
	SenderName      string    `json:"sender_name" swaggertype:"string"`
	SenderAvatarURL string    `json:"sender_avatar_url,omitempty" swaggertype:"string"`
	Text            string    `json:"text"`
	CreatedAt       time.Time `json:"created_at"`
	ChatId          uuid.UUID `json:"chat_id"`
	Type            string    `json:"type"` // Тип сообщения - системное или пользовательское
}

type CreateMessageDTO struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	ChatId    uuid.UUID `json:"chat_id" swaggertype:"string" format:"uuid"`
}
