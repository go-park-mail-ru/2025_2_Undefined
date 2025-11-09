package messages

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertMessageQuery = `
		INSERT INTO message (chat_id, user_id, text, created_at, message_type) 
		VALUES ($1, $2, $3, $4, $5::message_type_enum)
		RETURNING id`

	getMessageByIDQuery = `
		SELECT m.id, m.chat_id, m.user_id, u.name, 
			   m.text, m.created_at, m.message_type::text
		FROM message m
		JOIN "user" u ON u.id = m.user_id
		WHERE m.id = $1`

	updateMessageQuery = `
		UPDATE message 
		SET text = $2
		WHERE id = $1 AND user_id = $3
		RETURNING updated_at`

	deleteMessageQuery = `
		DELETE FROM message 
		WHERE id = $1 AND user_id = $2`
)

type MessageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{
		pool: pool,
	}
}

func (r *MessageRepository) InsertMessage(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error) {
	const op = "MessageRepository.InsertMessage"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", msg.ChatID.String()).WithField("user_id", msg.UserID.String())
	logger.Debug("Starting database operation: insert message")

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, insertMessageQuery, msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
		Scan(&id)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert message query")
		return uuid.Nil, fmt.Errorf("failed to insert message: %w", err)
	}

	logger.WithField("message_id", id.String()).Info("Database operation completed successfully: message inserted")
	return id, nil
}

func (r *MessageRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*modelsMessage.Message, error) {
	const op = "MessageRepository.GetMessageByID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("message_id", messageID.String())
	logger.Debug("Starting database operation: get message by ID")

	var message modelsMessage.Message
	err := r.pool.QueryRow(ctx, getMessageByIDQuery, messageID).
		Scan(&message.ID, &message.ChatID, &message.UserID, &message.UserName,
			&message.Text, &message.CreatedAt, &message.Type)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get message by ID query")
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	logger.Info("Database operation completed successfully: message retrieved")
	return &message, nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, messageID, userID uuid.UUID, newText string) error {
	const op = "MessageRepository.UpdateMessage"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("message_id", messageID.String()).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: update message")

	var updatedAt interface{}
	err := r.pool.QueryRow(ctx, updateMessageQuery, messageID, newText, userID).Scan(&updatedAt)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update message")
		return fmt.Errorf("failed to update message: %w", err)
	}

	logger.Info("Database operation completed successfully: message updated")
	return nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	const op = "MessageRepository.DeleteMessage"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("message_id", messageID.String()).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: delete message")

	cmdTag, err := r.pool.Exec(ctx, deleteMessageQuery, messageID, userID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: delete message")
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Debug("Database operation completed: message not found or access denied")
		return fmt.Errorf("message not found or access denied")
	}

	logger.Info("Database operation completed successfully: message deleted")
	return nil
}
