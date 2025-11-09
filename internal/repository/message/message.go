package messages

import (
	"context"
	"database/sql"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/google/uuid"
)

const (
	insertMessageQuery = `INSERT INTO message (chat_id, user_id, text, created_at, message_type) VALUES
						($1, $2, $3, $4, $5::message_type_enum)
						RETURNING id`
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) InsertMessage(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error) {
	const op = "MessageRepository.InsertMessage"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", msg.ChatID.String()).WithField("user_id", msg.UserID.String())
	logger.Debug("Starting database operation: insert message")

	var id uuid.UUID
	err := r.db.QueryRow(insertMessageQuery, msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
		Scan(&id)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert message query")
		return uuid.Nil, err
	}

	logger.WithField("message_id", id.String()).Info("Database operation completed successfully: message inserted")
	return id, nil
}
