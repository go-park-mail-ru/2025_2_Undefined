package messages

import (
	"database/sql"
	"fmt"
	"log"

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

func (r *MessageRepository) InsertMessage(msg modelsMessage.CreateMessage) (uuid.UUID, error) {
	const op = "MessageRepository.InsertMessage"
	var id uuid.UUID
	err := r.db.QueryRow(insertMessageQuery, msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
		Scan(&id)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to insert chat: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, wrappedErr
	}

	return id, nil
}
