package models

import (
	"time"

	modelsAttachment "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/attachment"
	"github.com/google/uuid"
)

const (
	MessageTypeUser   = "user"
	MessageTypeSystem = "system"
)

type Message struct {
	ID         uuid.UUID
	ChatID     uuid.UUID
	UserID     *uuid.UUID
	UserName   *string
	Text       string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	Type       string
	Attachment *modelsAttachment.Attachment
}

type CreateMessage struct {
	ChatID     uuid.UUID
	UserID     *uuid.UUID
	Text       string
	CreatedAt  time.Time
	Type       string
	Attachment *modelsAttachment.CreateAttachment
}
