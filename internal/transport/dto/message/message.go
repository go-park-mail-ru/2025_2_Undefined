package dto

import (
	"time"

	"github.com/google/uuid"
)

type MessageDTO struct {
	SenderID        *uuid.UUID `json:"sender_id" swaggertype:"string" format:"uuid"`
	SenderName      string     `json:"sender_name" swaggertype:"string"`
	SenderAvatarURL string     `json:"sender_avatar_url,omitempty" swaggertype:"string"`
	Text            string     `json:"text"`
	CreatedAt       time.Time  `json:"created_at" swaggertype:"string" format:"date-time"`
	ChatId          uuid.UUID  `json:"chat_id" swaggertype:"string" format:"uuid"`
	Type            string     `json:"type" swaggertype:"string"` // Тип сообщения - системное или пользовательское
}

type CreateMessageDTO struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	ChatId    uuid.UUID `json:"chat_id" swaggertype:"string" format:"uuid"`
}

const (
	WebSocketMessageTypeNewChatMessage    = "new message of chat"
	WebSocketMessageTypeEditChatMessage   = "edit chat message"
	WebSocketMessageTypeDeleteChatMessage = "delete chat message"
	WebSocketMessageTypeCreatedNewChat    = "new chat created"
)

type WebSocketMessageDTO struct {
	Type   string    `json:"type"`
	ChatID uuid.UUID `json:"chat_id"`
	Value  any       `json:"value"`
}
