package messages

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/pgxinterface"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertMessageQuery = `INSERT INTO message (chat_id, user_id, text, created_at, message_type) VALUES
						($1, $2, $3, $4, $5::message_type_enum)
						RETURNING id`

	getLastMessagesOfChatsQuery = `
		SELECT DISTINCT ON (msg.chat_id)
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text
		FROM message msg
		JOIN chat_member cm ON cm.chat_id = msg.chat_id
		LEFT JOIN "user" usr ON usr.id = msg.user_id
		WHERE cm.user_id = $1
		ORDER BY msg.chat_id, msg.created_at DESC`

	getLastMessagesOfChatsByIDsQuery = `
		SELECT DISTINCT ON (msg.chat_id)
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text
		FROM message msg
		LEFT JOIN "user" usr ON usr.id = msg.user_id
		WHERE msg.chat_id = ANY($1)
		ORDER BY msg.chat_id, msg.created_at DESC`

	getMessagesOfChatQuery = `
		SELECT 
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text
		FROM message msg
		LEFT JOIN "user" usr ON usr.id = msg.user_id
		WHERE chat_id = $1
		ORDER BY msg.created_at DESC
		LIMIT $3 OFFSET $2`

	searchMessagesInChatQuery = `
		SELECT 
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text
		FROM message msg
		JOIN chat_member cm ON cm.chat_id = msg.chat_id
		JOIN "user" usr ON usr.id = msg.user_id
		WHERE cm.user_id = $1 AND msg.chat_id = $2 AND msg.text ILIKE '%' || $3 || '%'
		ORDER BY msg.created_at DESC`
)

type MessageRepository struct {
	db pgxinterface.PgxPool
}

func NewMessageRepository(db pgxinterface.PgxPool) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

// NewMessageRepositoryWithPool создает репозиторий с конкретным типом *pgxpool.Pool
func NewMessageRepositoryWithPool(db *pgxpool.Pool) *MessageRepository {
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
	err := r.db.QueryRow(ctx, insertMessageQuery, msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
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
	count := 0
	defer func() {
		logger.Debugf("db query: %s: status: %s, count: %d", query, queryStatus, count)
	}()

	logger.Debugf("starting: %s", query)

	rows, err := r.db.Query(ctx, getLastMessagesOfChatsQuery, userId)
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
			&message.Text, &message.CreatedAt, &message.UpdatedAt,
			&message.Type); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}

		result = append(result, message)
	}

	count = len(result)

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

	rows, err := r.db.Query(ctx, getMessagesOfChatQuery, chatId, offset, limit)
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
			&message.Text, &message.CreatedAt, &message.UpdatedAt,
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

func (r *MessageRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (modelsMessage.Message, error) {
	const op = "MessageRepository.GetMessageByID"
	const query = "SELECT message by ID"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("message_id", messageID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var message modelsMessage.Message
	err := r.db.QueryRow(ctx,
		`SELECT id, chat_id, user_id, text, created_at, updated_at, message_type::text
		 FROM message WHERE id = $1`,
		messageID,
	).Scan(&message.ID, &message.ChatID, &message.UserID, &message.Text, &message.CreatedAt, &message.UpdatedAt, &message.Type)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return modelsMessage.Message{}, err
	}

	return message, nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, messageID uuid.UUID, newText string) error {
	const op = "MessageRepository.UpdateMessage"
	const query = "UPDATE message"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("message_id", messageID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	_, err := r.db.Exec(ctx,
		`UPDATE message SET text = $1, updated_at = NOW() WHERE id = $2`,
		newText, messageID,
	)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return err
	}

	return nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	const op = "MessageRepository.DeleteMessage"
	const query = "DELETE message"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("message_id", messageID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	_, err := r.db.Exec(ctx,
		`DELETE FROM message WHERE id = $1`,
		messageID,
	)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return err
	}

	return nil
}

func (r *MessageRepository) GetLastMessagesOfChatsByIDs(ctx context.Context, chatsIDs []uuid.UUID) (map[uuid.UUID]modelsMessage.Message, error) {
	const op = "MessageRepository.GetLastMessagesOfChatsByIDs"
	const query = "SELECT last messages by IDs"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chats_ids", chatsIDs)

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	rows, err := r.db.Query(ctx, getLastMessagesOfChatsByIDsQuery, chatsIDs)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]modelsMessage.Message)
	for rows.Next() {
		var message modelsMessage.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.UserName,
			&message.Text, &message.CreatedAt, &message.UpdatedAt,
			&message.Type); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}

		result[message.ChatID] = message
	}

	return result, nil
}

func (r *MessageRepository) SearchMessagesInChat(ctx context.Context, userId, chatId uuid.UUID, searchText string) ([]modelsMessage.Message, error) {
	const op = "MessageRepository.SearchMessagesInChat"
	const query = "SEARCH messages"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("user_id", userId.String()).
		WithField("chat_id", chatId.String()).
		WithField("search_text", searchText)

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	rows, err := r.db.Query(ctx, searchMessagesInChatQuery, userId, chatId, searchText)
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
			&message.Text, &message.CreatedAt, &message.UpdatedAt,
			&message.Type); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}

		result = append(result, message)
	}

	return result, nil
}
