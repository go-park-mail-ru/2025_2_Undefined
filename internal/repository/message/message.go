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

	getLastMessagesOfChatsQuery = `
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT DISTINCT ON (msg.chat_id)
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			la.attachment_id,
			msg.text, msg.created_at, msg.message_type::text
		FROM message msg
		JOIN chat_member cm ON cm.chat_id = msg.chat_id
		JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN latest_avatars la ON la.user_id = msg.user_id
		WHERE cm.user_id = $1
		ORDER BY msg.chat_id, msg.created_at DESC`

	getMessagesOfChatQuery = `
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT 
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			la.attachment_id,
			msg.text, msg.created_at, msg.message_type::text
		FROM message msg
		JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN latest_avatars la ON la.user_id = msg.user_id
		WHERE chat_id = $1
		ORDER BY msg.created_at DESC
		LIMIT $3 OFFSET $2`
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
	const query = "INSERT message"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("chat_id", msg.ChatID.String()).
		WithField("user_id", msg.UserID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var id uuid.UUID
	err := r.db.QueryRow(insertMessageQuery, msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
		Scan(&id)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return uuid.Nil, err
	}

	return id, nil
}

func (r *MessageRepository) GetLastMessagesOfChats(ctx context.Context, userId uuid.UUID) ([]modelsMessage.Message, error) {
	const op = "MessageRepository.GetLastMessagesOfChats"
	const query = "SELECT last messages"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	rows, err := r.db.Query(getLastMessagesOfChatsQuery, userId)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}
	defer rows.Close()

	result := make([]modelsMessage.Message, 0)
	for rows.Next() {
		var message modelsMessage.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.UserName,
			&message.UserAvatarID, &message.Text, &message.CreatedAt,
			&message.Type); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}

		result = append(result, message)
	}

	return result, nil
}

func (r *MessageRepository) GetMessagesOfChat(ctx context.Context, chatId uuid.UUID, offset, limit int) ([]modelsMessage.Message, error) {
	const op = "MessageRepository.GetMessagesOfChat"
	const query = "SELECT chat messages"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("chat_id", chatId.String()).
		WithField("offset", offset).
		WithField("limit", limit)

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	rows, err := r.db.Query(getMessagesOfChatQuery, chatId, offset, limit)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}
	defer rows.Close()

	result := make([]modelsMessage.Message, 0)
	for rows.Next() {
		var message modelsMessage.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.UserName,
			&message.UserAvatarID, &message.Text, &message.CreatedAt,
			&message.Type); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}

		message.ChatID = chatId
		result = append(result, message)
	}

	return result, nil
}
