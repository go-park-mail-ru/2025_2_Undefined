package dto

import "github.com/google/uuid"

type AttachmentDTO struct {
	ID       *uuid.UUID `json:"id,omitempty" swaggertype:"string" format:"uuid"`
	Type     *string    `json:"type,omitempty" swaggertype:"string"` // sticker, voice, video_note
	FileURL  string     `json:"file_url,omitempty" swaggertype:"string"`
	Duration *int       `json:"duration,omitempty"` // Длительность в секундах для voice/video_note
}

type CreateAttachmentDTO struct {
	Type         string `json:"type"`                    // sticker, voice, video_note
	AttachmentID string `json:"attachment_id,omitempty"` // либо айди вложение либо айди стикера (может не быть в uuid)
	Duration     *int   `json:"duration,omitempty"`      // Для voice/video_note
}
