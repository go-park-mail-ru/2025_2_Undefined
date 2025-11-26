package dto

import (
	"time"

	"github.com/google/uuid"
)

type MessageDTO struct {
	ID         uuid.UUID  `json:"id" swaggertype:"string" format:"uuid"`
	SenderID   *uuid.UUID `json:"sender_id" swaggertype:"string" format:"uuid"`
	SenderName *string    `json:"sender_name" swaggertype:"string"`
	Text       string     `json:"text"`
	CreatedAt  time.Time  `json:"created_at" swaggertype:"string" format:"date-time"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty" swaggertype:"string" format:"date-time"`
	ChatID     uuid.UUID  `json:"chat_id" swaggertype:"string" format:"uuid"`
	Type       string     `json:"type" swaggertype:"string"` // Тип сообщения - системное или пользовательское
}

type CreateMessageDTO struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	ChatId    uuid.UUID `json:"chat_id" swaggertype:"string" format:"uuid"`
}

type EditMessageDTO struct {
	ID        uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	Text      string    `json:"text"`
	UpdatedAt time.Time `json:"updated_at" swaggertype:"string" format:"date-time"`
}

type DeleteMessageDTO struct {
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
}

type UserJoinedDTO struct {
	UserID uuid.UUID `json:"user_id" swaggertype:"string" format:"uuid"`
	ChatID uuid.UUID `json:"chat_id" swaggertype:"string" format:"uuid"`
}

const (
	WebSocketMessageTypeNewChatMessage    = "new_message"
	WebSocketMessageTypeEditChatMessage   = "edit_message"
	WebSocketMessageTypeDeleteChatMessage = "delete_message"
	WebSocketMessageTypeCreatedNewChat    = "chat_created"
)

type WebSocketMessageDTO struct {
	Type   string    `json:"type"`
	ChatID uuid.UUID `json:"chat_id"`
	Value  any       `json:"value"`
}
