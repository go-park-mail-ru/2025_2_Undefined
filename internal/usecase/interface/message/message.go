package messages

import (
	"context"

	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/google/uuid"
)

type MessageRepository interface {
	InsertMessage(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error)
}
