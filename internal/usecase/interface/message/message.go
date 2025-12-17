package messages

import (
	"context"

	modelsAttachment "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/attachment"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/google/uuid"
)

type MessageRepository interface {
	InsertMessage(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error)
	InsertMessageWithAttachment(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error)
	GetLastMessagesOfChats(ctx context.Context, userID uuid.UUID) ([]modelsMessage.Message, error)
	GetMessagesOfChat(ctx context.Context, chatID uuid.UUID, offset, limit int) ([]modelsMessage.Message, error)
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (modelsMessage.Message, error)
	GetMessageAttachments(ctx context.Context, messageID uuid.UUID) (*modelsAttachment.Attachment, error)
	UpdateMessage(ctx context.Context, messageID uuid.UUID, newText string) error
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error
	SearchMessagesInChat(ctx context.Context, userID uuid.UUID, chatID uuid.UUID, text string) ([]modelsMessage.Message, error)
	GetLastMessagesOfChatsByIDs(ctx context.Context, chatIDs []uuid.UUID) (map[uuid.UUID]modelsMessage.Message, error)
	InsertAttachment(ctx context.Context, attachment modelsAttachment.CreateAttachment, userID uuid.UUID) error
	GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*modelsAttachment.Attachment, error)
	CheckAttachmentOwnership(ctx context.Context, attachmentID, userID uuid.UUID) (bool, error)
	LinkAttachmentToMessage(ctx context.Context, messageID, attachmentID, userID uuid.UUID) error
	UpdateAttachmentType(ctx context.Context, attachmentID uuid.UUID, attachmentType string) error
}
