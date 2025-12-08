package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	AttachmentTypeImage     = "image"
	AttachmentTypeDocument  = "document"
	AttachmentTypeAudio     = "audio"
	AttachmentTypeVideo     = "video"
	AttachmentTypeSticker   = "sticker"
	AttachmentTypeVoice     = "voice"
	AttachmentTypeVideoNote = "video_note"
)

type Attachment struct {
	ID                 uuid.UUID
	Type               *string
	FileName           string
	FileSize           int64
	ContentDisposition string
	Duration           *int // Длительность в секундах для voice/video_note/audio
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateAttachment struct {
	ID                 uuid.UUID
	Type               *string
	FileName           string
	FileSize           int64
	ContentDisposition string
	Duration           *int
}

type MessageAttachment struct {
	MessageID    uuid.UUID
	AttachmentID uuid.UUID
	UserID       uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
