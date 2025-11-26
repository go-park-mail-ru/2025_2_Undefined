package repository

import (
	"context"
	"fmt"
	"strings"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: begin transaction")
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Вставка чата
	logger.Debug("Executing database query: INSERT chat")
	_, err = tx.Exec(ctx, `INSERT INTO chat (id, chat_type, name, description) 
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

	placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d::message_type_enum)",
		len(values)+1, len(values)+2, len(values)+3, len(values)+4))

	var text string
	if chat.Type == modelsChats.ChatTypeChannel {
		text = "Канал создан"
	} else {
		text = "Чат создан"
	}

	values = append(values, chat.ID, nil, text, "system")

	if chat.Type == modelsChats.ChatTypeGroup {
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
		_, err = tx.Exec(ctx, query, values...)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: insert system messages")
			return err
		}
	}

	logger.Debug("Committing database transaction")
	if err := tx.Commit(ctx); err != nil {
		logger.WithError(err).Error("Database operation failed: commit transaction")
		return err
	}

	logger.Info("Database operation completed successfully: chat created with transaction")
	return nil
}

func (r *ChatsRepository) InsertUsersToChat(ctx context.Context, chatID uuid.UUID, usersInfo []modelsChats.UserInfo) error {
	const op = "ChatsRepository.InsertUsersToChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatID.String()).WithField("users_count", len(usersInfo))
	logger.Debug("Starting database operation: insert users to chat")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: begin transaction")
		return err
	}
	defer tx.Rollback(ctx)

	err = r.insertUsersToChat(ctx, tx, chatID, usersInfo)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert users to chat")
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.WithError(err).Error("Database operation failed: commit transaction")
		return err
	}

	logger.Info("Database operation completed successfully: users inserted to chat")

	return nil
}

func (r *ChatsRepository) insertUsersToChat(ctx context.Context, tx pgx.Tx, chatID uuid.UUID, usersInfo []modelsChats.UserInfo) error {
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
	_, err := tx.Exec(ctx, query, values...)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert chat members")
		return err
	}

	return nil
}
