package repository

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
)

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
