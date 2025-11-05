package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"

	"github.com/google/uuid"
)

const (
	getChatsQuery = `
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1`

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

	getChatQuery = `
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1 AND c.id = $2`

	getUsersOfChat = `
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT 
			cm.user_id, cm.chat_id, usr.name, 
			la.attachment_id,
			cm.chat_member_role::text
		FROM chat_member cm
		JOIN "user" usr ON usr.id = cm.user_id
		LEFT JOIN latest_avatars la ON la.user_id = cm.user_id
		WHERE cm.chat_id = $1`

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

	getUserInfo = `
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT 
			cm.user_id, cm.chat_id, usr.name, 
			la.attachment_id,
			cm.chat_member_role::text
		FROM chat_member cm
		JOIN "user" usr ON usr.id = cm.user_id
		LEFT JOIN latest_avatars la ON la.user_id = cm.user_id
		WHERE cm.user_id = $1 AND cm.chat_id = $2`

	getUsersDialogQuery = `
		SELECT chat.id 
		FROM chat
		LEFT JOIN chat_member cm1 ON cm1.chat_id = chat.id
		LEFT JOIN chat_member cm2 ON cm2.chat_id = chat.id
		WHERE cm1.user_id = $1 AND cm2.user_id = $2`

	checkUserRoleQuery = `
		SELECT EXISTS(
			SELECT 1 FROM chat_member 
			WHERE user_id = $1 AND chat_id = $2 AND chat_member_role = $3::chat_member_role_enum
		)`

	deleteChatQuery = `DELETE FROM chat WHERE id = $1`

	updateChatQuery = `UPDATE chat SET name = $1, description = $2 WHERE id = $3`
)

type ChatsRepository struct {
	db *sql.DB
}

func NewChatsRepository(db *sql.DB) *ChatsRepository {
	return &ChatsRepository{
		db: db,
	}
}

func (r *ChatsRepository) GetChats(ctx context.Context, userId uuid.UUID) ([]modelsChats.Chat, error) {
	const op = "ChatsRepository.GetChats"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String())
	logger.Debug("Starting database operation: get user chats")

	rows, err := r.db.Query(getChatsQuery, userId)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get user chats query")
		return nil, err
	}
	defer rows.Close()

	result := make([]modelsChats.Chat, 0)
	for rows.Next() {
		var chat modelsChats.Chat
		if err := rows.Scan(&chat.ID, &chat.Type, &chat.Name, &chat.Description); err != nil {
			logger.WithError(err).Error("Database operation failed: scan chat row")
			return nil, err
		}

		result = append(result, chat)
	}

	logger.WithField("chats_count", len(result)).Info("Database operation completed successfully: user chats retrieved")
	return result, nil
}

func (r *ChatsRepository) GetLastMessagesOfChats(ctx context.Context, userId uuid.UUID) ([]modelsMessage.Message, error) {
	const op = "ChatsRepository.GetLastMessagesOfChats"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String())
	logger.Debug("Starting database operation: get last messages of user chats")

	rows, err := r.db.Query(getLastMessagesOfChatsQuery, userId)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get last messages query")
		return nil, err
	}
	defer rows.Close()

	result := make([]modelsMessage.Message, 0)
	for rows.Next() {
		var message modelsMessage.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.UserName,
			&message.UserAvatarID, &message.Text, &message.CreatedAt,
			&message.Type); err != nil {
			logger.WithError(err).Error("Database operation failed: scan message row")
			return nil, err
		}

		result = append(result, message)
	}

	logger.WithField("messages_count", len(result)).Info("Database operation completed successfully: last messages retrieved")
	return result, nil
}

func (r *ChatsRepository) GetChat(ctx context.Context, userId, chatId uuid.UUID) (*modelsChats.Chat, error) {
	const op = "ChatsRepository.GetChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String()).WithField("chat_id", chatId.String())
	logger.Debug("Starting database operation: get specific chat")

	chat := &modelsChats.Chat{}

	err := r.db.QueryRow(getChatQuery, userId, chatId).
		Scan(&chat.ID, &chat.Type, &chat.Name, &chat.Description)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get chat query")
		return nil, err
	}

	logger.Info("Database operation completed successfully: chat retrieved")
	return chat, nil
}

func (r *ChatsRepository) GetUsersOfChat(ctx context.Context, chatId uuid.UUID) ([]modelsChats.UserInfo, error) {
	const op = "ChatsRepository.GetUsersOfChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatId.String())
	logger.Debug("Starting database operation: get chat users")

	rows, err := r.db.Query(getUsersOfChat, chatId)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get chat users query")
		return nil, err
	}
	defer rows.Close()

	result := make([]modelsChats.UserInfo, 0)
	for rows.Next() {
		var userInfo modelsChats.UserInfo
		if err := rows.Scan(&userInfo.UserID, &userInfo.ChatID, &userInfo.UserName,
			&userInfo.UserAvatarID, &userInfo.Role); err != nil {
			logger.WithError(err).Error("Database operation failed: scan user info row")
			return nil, err
		}

		result = append(result, userInfo)
	}

	logger.WithField("users_count", len(result)).Info("Database operation completed successfully: chat users retrieved")
	return result, nil
}

func (r *ChatsRepository) GetMessagesOfChat(ctx context.Context, chatId uuid.UUID, offset, limit int) ([]modelsMessage.Message, error) {
	const op = "ChatsRepository.GetMessagesOfChats"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatId.String()).WithField("offset", offset).WithField("limit", limit)
	logger.Debug("Starting database operation: get chat messages")

	rows, err := r.db.Query(getMessagesOfChatQuery, chatId, offset, limit)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get chat messages query")
		return nil, err
	}
	defer rows.Close()

	result := make([]modelsMessage.Message, 0)
	for rows.Next() {
		var message modelsMessage.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.UserName,
			&message.UserAvatarID, &message.Text, &message.CreatedAt,
			&message.Type); err != nil {
			logger.WithError(err).Error("Database operation failed: scan message row")
			return nil, err
		}

		message.ChatID = chatId
		result = append(result, message)
	}

	logger.WithField("messages_count", len(result)).Info("Database operation completed successfully: chat messages retrieved")
	return result, nil
}

func (r *ChatsRepository) GetUsersDialog(ctx context.Context, user1ID, user2ID uuid.UUID) (uuid.UUID, error) {
	const op = "ChatsRepository.GetUsersDialog"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Starting database operation: get users dialog")

	var chatID uuid.UUID

	err := r.db.QueryRow(getUsersDialogQuery, user1ID, user2ID).Scan(&chatID)
	if err != nil {
		logger.WithError(err).Errorf("error getting dialog users: %s and %s", user1ID.String(), user2ID.String())
		return uuid.Nil, err
	}

	return chatID, nil
}

func (r *ChatsRepository) CreateChat(ctx context.Context, chat modelsChats.Chat, usersInfo []modelsChats.UserInfo, usersNames []string) error {
	const op = "ChatsRepository.CreateChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chat.ID.String()).WithField("users_count", len(usersInfo))
	logger.Debug("Starting database operation: create chat with transaction")

	if len(usersInfo) != len(usersNames) || len(usersInfo) == 0 {
		err := fmt.Errorf("invalid input of users ids: usersInfo and usersNames must have the same non-zero length")
		logger.WithError(err).Error("Database operation failed: invalid input parameters")
		return err
	}

	logger.Debug("Starting database transaction")
	tx, err := r.db.Begin()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: begin transaction")
		return err
	}
	defer tx.Rollback()

	// 1. Вставка чата
	logger.Debug("Executing database query: INSERT chat")
	_, err = tx.Exec(`INSERT INTO chat (id, chat_type, name, description) 
        VALUES ($1, $2::chat_type_enum, $3, $4)`,
		chat.ID, chat.Type, chat.Name, chat.Description)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert chat")
		return err
	}

	// 2. Вставка участников чата
	err = r.insertUsersToChat(ctx, tx, chat.ID, usersInfo)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert chat members")
		return err
	}

	// 3. Вставка системных сообщений
	query := `INSERT INTO message (chat_id, user_id, text, message_type) VALUES `
	values := []interface{}{}
	placeholders := []string{}

	if chat.Type == modelsChats.ChatTypeDialog {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d::message_type_enum)",
			len(values)+1, len(values)+2, len(values)+3, len(values)+4))
		text := "Чат создан"
		values = append(values, chat.ID, usersInfo[0].UserID, text, "system")
	} else {
		for i, userName := range usersNames {

			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d::message_type_enum)",
				len(values)+1, len(values)+2, len(values)+3, len(values)+4))
			text := fmt.Sprintf("Пользователь %s вступил в чат", userName)
			values = append(values, chat.ID, usersInfo[i].UserID, text, "system")

		}
	}

	if len(placeholders) > 0 {
		query += strings.Join(placeholders, ", ")
		logger.Debug("Executing database query: INSERT system messages")
		_, err = tx.Exec(query, values...)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: insert system messages")
			return err
		}
	}

	logger.Debug("Committing database transaction")
	if err := tx.Commit(); err != nil {
		logger.WithError(err).Error("Database operation failed: commit transaction")
		return err
	}

	logger.Info("Database operation completed successfully: chat created with transaction")
	return nil
}

func (r *ChatsRepository) GetUserInfo(ctx context.Context, userId, chatId uuid.UUID) (*modelsChats.UserInfo, error) {
	const op = "ChatsRepository.GetUserInfo"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String()).WithField("chat_id", chatId.String())
	logger.Debug("Starting database operation: get user info in chat")

	userInfo := &modelsChats.UserInfo{}

	err := r.db.QueryRow(getUserInfo, userId, chatId).
		Scan(&userInfo.UserID, &userInfo.ChatID, &userInfo.UserName,
			&userInfo.UserAvatarID, &userInfo.Role)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get user info query")
		return nil, err
	}

	logger.Info("Database operation completed successfully: user info retrieved")
	return userInfo, nil
}

func (r *ChatsRepository) InsertUsersToChat(ctx context.Context, chatID uuid.UUID, usersInfo []modelsChats.UserInfo) error {
	const op = "ChatsRepository.InsertUsersToChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatID.String()).WithField("users_count", len(usersInfo))
	logger.Debug("Starting database operation: insert users to chat")

	tx, err := r.db.Begin()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: begin transaction")
		return err
	}
	defer tx.Rollback()

	err = r.insertUsersToChat(ctx, tx, chatID, usersInfo)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert users to chat")
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.WithError(err).Error("Database operation failed: commit transaction")
		return err
	}

	logger.Info("Database operation completed successfully: users inserted to chat")

	return nil
}

func (r *ChatsRepository) insertUsersToChat(ctx context.Context, tx *sql.Tx, chatID uuid.UUID, usersInfo []modelsChats.UserInfo) error {
	const op = "ChatsRepository.insertUsersToChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatID.String()).WithField("users_count", len(usersInfo))

	query := `INSERT INTO chat_member (user_id, chat_id, chat_member_role) VALUES `
	values := []interface{}{}
	placeholders := []string{}

	for _, userInfo := range usersInfo {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d::chat_member_role_enum)",
			len(values)+1, len(values)+2, len(values)+3))
		values = append(values, userInfo.UserID, chatID, userInfo.Role)
	}

	query += strings.Join(placeholders, ", ")
	logger.Debug("Executing database query: INSERT chat members")
	_, err := tx.Exec(query, values...)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert chat members")
		return err
	}

	return nil
}

func (r *ChatsRepository) CheckUserHasRole(ctx context.Context, userId, chatId uuid.UUID, role string) (bool, error) {
	const op = "ChatsRepository.CheckUserHasRole"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String()).WithField("chat_id", chatId.String()).WithField("role", role)
	logger.Debug("Starting database operation: check user role in chat")

	var hasRole bool
	err := r.db.QueryRow(checkUserRoleQuery, userId, chatId, role).Scan(&hasRole)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check user role query")
		return false, err
	}

	logger.WithField("has_role", hasRole).Info("Database operation completed successfully: user role checked")
	return hasRole, nil
}

func (r *ChatsRepository) DeleteChat(ctx context.Context, userId, chatId uuid.UUID) error {
	const op = "ChatsRepository.DeleteChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String()).WithField("chat_id", chatId.String())
	logger.Debug("Starting database operation: delete chat")

	// Проверяем, является ли пользователь администратором чата
	isAdmin, err := r.CheckUserHasRole(ctx, userId, chatId, "admin")
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check user admin status")
		return err
	}

	if !isAdmin {
		err := fmt.Errorf("user is not admin of the chat")
		logger.WithError(err).Error("Database operation failed: user permission denied")
		return err
	}

	// Удаляем чат (связанные записи удалятся каскадно)
	result, err := r.db.Exec(deleteChatQuery, chatId)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: delete chat")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check rows affected")
		return err
	}

	if rowsAffected == 0 {
		err := fmt.Errorf("chat not found")
		logger.WithError(err).Error("Database operation failed: chat not found")
		return err
	}

	logger.Info("Database operation completed successfully: chat deleted")
	return nil
}

func (r *ChatsRepository) UpdateChat(ctx context.Context, userId, chatId uuid.UUID, name, description string) error {
	const op = "ChatsRepository.UpdateChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String()).WithField("chat_id", chatId.String())
	logger.Debug("Starting database operation: update chat")

	// Проверяем, является ли пользователь администратором чата
	isAdmin, err := r.CheckUserHasRole(ctx, userId, chatId, "admin")
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check user admin status")
		return err
	}

	if !isAdmin {
		err := fmt.Errorf("user is not admin of the chat")
		logger.WithError(err).Error("Database operation failed: user permission denied")
		return err
	}

	result, err := r.db.Exec(updateChatQuery, name, description, chatId)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update chat")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check rows affected")
		return err
	}

	if rowsAffected == 0 {
		err := fmt.Errorf("chat not found")
		logger.WithError(err).Error("Database operation failed: chat not found")
		return err
	}

	logger.Info("Database operation completed successfully: chat updated")
	return nil
}
