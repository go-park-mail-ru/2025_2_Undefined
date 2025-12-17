package messages

import (
	"context"

	modelsAttachment "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/attachment"
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

	insertAttachmentQuery = `
		INSERT INTO attachment (id, attachment_type, file_name, file_size, content_disposition, duration)
		VALUES ($1, $2::attachment_type_enum, $3, $4, $5, $6)`

	insertMessageAttachmentQuery = `
		INSERT INTO message_attachment (message_id, attachment_id, user_id)
		VALUES ($1, $2, $3)`

	insertPendingAttachmentQuery = `
		INSERT INTO pending_attachment (attachment_id, user_id)
		VALUES ($1, $2)`

	getAttachmentByIDQuery = `
		SELECT id, attachment_type::text, file_name, file_size, content_disposition, 
		       duration, created_at, updated_at
		FROM attachment
		WHERE id = $1`

	checkAttachmentOwnershipQuery = `
		SELECT EXISTS(
			SELECT 1 FROM pending_attachment
			WHERE attachment_id = $1 AND user_id = $2
			UNION
			SELECT 1 FROM message_attachment
			WHERE attachment_id = $1 AND user_id = $2
		)`

	deletePendingAttachmentQuery = `
		DELETE FROM pending_attachment
		WHERE attachment_id = $1`

	updateAttachmentTypeQuery = `
		UPDATE attachment
		SET attachment_type = $2::attachment_type_enum
		WHERE id = $1`

	getMessageAttachmentsQuery = `
		SELECT a.id, a.attachment_type::text, a.file_name, a.file_size, a.content_disposition, 
		       a.duration, a.created_at, a.updated_at
		FROM attachment a
		JOIN message_attachment ma ON ma.attachment_id = a.id
		WHERE ma.message_id = $1`

	getLastMessagesOfChatsQuery = `
		SELECT DISTINCT ON (msg.chat_id)
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text,
			a.id, a.attachment_type::text, a.file_name, a.file_size, a.content_disposition, a.duration
		FROM message msg
		JOIN chat_member cm ON cm.chat_id = msg.chat_id
		LEFT JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN message_attachment ma ON ma.message_id = msg.id
		LEFT JOIN attachment a ON a.id = ma.attachment_id
		WHERE cm.user_id = $1
		ORDER BY msg.chat_id, msg.created_at DESC`

	getLastMessagesOfChatsByIDsQuery = `
		SELECT DISTINCT ON (msg.chat_id)
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text,
			a.id, a.attachment_type::text, a.file_name, a.file_size, a.content_disposition, a.duration
		FROM message msg
		LEFT JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN message_attachment ma ON ma.message_id = msg.id
		LEFT JOIN attachment a ON a.id = ma.attachment_id
		WHERE msg.chat_id = ANY($1)
		ORDER BY msg.chat_id, msg.created_at DESC`

	getMessagesOfChatQuery = `
		SELECT 
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text,
			a.id, a.attachment_type::text, a.file_name, a.file_size, a.content_disposition, a.duration
		FROM message msg
		LEFT JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN message_attachment ma ON ma.message_id = msg.id
		LEFT JOIN attachment a ON a.id = ma.attachment_id
		WHERE chat_id = $1
		ORDER BY msg.created_at DESC
		LIMIT $3 OFFSET $2`

	searchMessagesInChatQuery = `
		SELECT 
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			msg.text, msg.created_at, msg.updated_at, msg.message_type::text,
			a.id, a.attachment_type::text, a.file_name, a.file_size, a.content_disposition, a.duration
		FROM message msg
		JOIN chat_member cm ON cm.chat_id = msg.chat_id
		JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN message_attachment ma ON ma.message_id = msg.id
		LEFT JOIN attachment a ON a.id = ma.attachment_id
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

// scanMessageWithAttachment сканирует строку с сообщением и вложением
func scanMessageWithAttachment(scanner interface {
	Scan(dest ...interface{}) error
}, message *modelsMessage.Message) error {
	var attachmentID, attachmentType, attachmentFileName, attachmentContentDisposition *string
	var attachmentFileSize *int64
	var attachmentDuration *int

	err := scanner.Scan(
		&message.ID, &message.ChatID, &message.UserID, &message.UserName,
		&message.Text, &message.CreatedAt, &message.UpdatedAt, &message.Type,
		&attachmentID, &attachmentType, &attachmentFileName, &attachmentFileSize,
		&attachmentContentDisposition, &attachmentDuration,
	)
	if err != nil {
		return err
	}

	// Если есть вложение, добавляем его
	if attachmentID != nil && attachmentType != nil {
		id, _ := uuid.Parse(*attachmentID)
		message.Attachment = &modelsAttachment.Attachment{
			ID:                 id,
			Type:               attachmentType,
			FileName:           *attachmentFileName,
			FileSize:           *attachmentFileSize,
			ContentDisposition: *attachmentContentDisposition,
			Duration:           attachmentDuration,
		}
	}

	return nil
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
		if err := scanMessageWithAttachment(rows, &message); err != nil {
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
		if err := scanMessageWithAttachment(rows, &message); err != nil {
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
		if err := scanMessageWithAttachment(rows, &message); err != nil {
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
		if err := scanMessageWithAttachment(rows, &message); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}

		result = append(result, message)
	}

	return result, nil
}

func (r *MessageRepository) InsertMessageWithAttachment(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error) {
	const op = "MessageRepository.InsertMessageWithAttachment"
	const query = "INSERT message with attachment"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("chat_id", msg.ChatID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: begin transaction error: status: %s", query, queryStatus)
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	// Вставляем сообщение
	var messageID uuid.UUID
	err = tx.QueryRow(ctx, insertMessageQuery, msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
		Scan(&messageID)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert message error: status: %s", query, queryStatus)
		return uuid.Nil, err
	}

	// Если есть вложение, вставляем его
	if msg.Attachment != nil {
		_, err = tx.Exec(ctx, insertAttachmentQuery,
			msg.Attachment.ID,
			msg.Attachment.Type,
			msg.Attachment.FileName,
			msg.Attachment.FileSize,
			msg.Attachment.ContentDisposition,
			msg.Attachment.Duration,
		)
		if err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: insert attachment error: status: %s", query, queryStatus)
			return uuid.Nil, err
		}

		// Связываем сообщение с вложением
		_, err = tx.Exec(ctx, insertMessageAttachmentQuery, messageID, msg.Attachment.ID, msg.UserID)
		if err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: insert message_attachment error: status: %s", query, queryStatus)
			return uuid.Nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: commit transaction error: status: %s", query, queryStatus)
		return uuid.Nil, err
	}

	return messageID, nil
}

func (r *MessageRepository) InsertAttachment(ctx context.Context, attachment modelsAttachment.CreateAttachment, userID uuid.UUID) error {
	const op = "MessageRepository.InsertAttachment"
	const query = "INSERT attachment"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("attachment_id", attachment.ID.String()).
		WithField("user_id", userID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: begin transaction error: status: %s", query, queryStatus)
		return err
	}
	defer tx.Rollback(ctx)

	// Вставляем вложение
	_, err = tx.Exec(ctx, insertAttachmentQuery,
		attachment.ID,
		attachment.Type,
		attachment.FileName,
		attachment.FileSize,
		attachment.ContentDisposition,
		attachment.Duration,
	)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert attachment error: status: %s", query, queryStatus)
		return err
	}

	// Создаём запись в pending_attachment
	_, err = tx.Exec(ctx, insertPendingAttachmentQuery, attachment.ID, userID)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert pending_attachment error: status: %s", query, queryStatus)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: commit transaction error: status: %s", query, queryStatus)
		return err
	}

	return nil
}

func (r *MessageRepository) UpdateAttachmentType(ctx context.Context, attachmentID uuid.UUID, attachmentType string) error {
	const op = "MessageRepository.UpdateAttachmentType"
	const query = "UPDATE attachment type"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("attachment_id", attachmentID.String()).
		WithField("type", attachmentType)

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	_, err := r.db.Exec(ctx, updateAttachmentTypeQuery, attachmentID, attachmentType)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return err
	}

	return nil
}

func (r *MessageRepository) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*modelsAttachment.Attachment, error) {
	const op = "MessageRepository.GetAttachmentByID"
	const query = "SELECT attachment"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("attachment_id", attachmentID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var attachment modelsAttachment.Attachment
	err := r.db.QueryRow(ctx, getAttachmentByIDQuery, attachmentID).Scan(
		&attachment.ID,
		&attachment.Type,
		&attachment.FileName,
		&attachment.FileSize,
		&attachment.ContentDisposition,
		&attachment.Duration,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: status: %s", query, queryStatus)
		return nil, err
	}

	return &attachment, nil
}

func (r *MessageRepository) CheckAttachmentOwnership(ctx context.Context, attachmentID, userID uuid.UUID) (bool, error) {
	const op = "MessageRepository.CheckAttachmentOwnership"
	const query = "CHECK attachment ownership"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("attachment_id", attachmentID.String()).
		WithField("user_id", userID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var exists bool
	err := r.db.QueryRow(ctx, checkAttachmentOwnershipQuery, attachmentID, userID).Scan(&exists)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: status: %s", query, queryStatus)
		return false, err
	}

	return exists, nil
}

func (r *MessageRepository) LinkAttachmentToMessage(ctx context.Context, messageID, attachmentID, userID uuid.UUID) error {
	const op = "MessageRepository.LinkAttachmentToMessage"
	const query = "LINK attachment to message"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("message_id", messageID.String()).
		WithField("attachment_id", attachmentID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: begin transaction error: status: %s", query, queryStatus)
		return err
	}
	defer tx.Rollback(ctx)

	// Удаляем из pending_attachment
	_, err = tx.Exec(ctx, deletePendingAttachmentQuery, attachmentID)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: delete pending_attachment error: status: %s", query, queryStatus)
		return err
	}

	// Добавляем в message_attachment
	_, err = tx.Exec(ctx, insertMessageAttachmentQuery, messageID, attachmentID, userID)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert message_attachment error: status: %s", query, queryStatus)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: commit transaction error: status: %s", query, queryStatus)
		return err
	}

	return nil
}

func (r *MessageRepository) GetMessageAttachments(ctx context.Context, messageID uuid.UUID) (*modelsAttachment.Attachment, error) {
	const op = "MessageRepository.GetMessageAttachments"
	const query = "SELECT message attachments"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("message_id", messageID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var attachment modelsAttachment.Attachment
	err := r.db.QueryRow(ctx, getMessageAttachmentsQuery, messageID).
		Scan(&attachment.ID, &attachment.Type, &attachment.FileName, &attachment.FileSize,
			&attachment.ContentDisposition, &attachment.Duration, &attachment.CreatedAt, &attachment.UpdatedAt)
	if err != nil {
		// Если вложений нет, возвращаем nil
		return nil, nil
	}

	return &attachment, nil
}
