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

	result, err := r.db.Exec(ctx, updateChatQuery, name, description, chatId)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update chat")
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		err := fmt.Errorf("chat not found")
		logger.WithError(err).Error("Database operation failed: chat not found")
		return err
	}

	logger.Info("Database operation completed successfully: chat updated")
	return nil
}

func (r *ChatsRepository) UpdateChatAvatar(ctx context.Context, chatID uuid.UUID, attachmentID uuid.UUID, fileSize int64) error {
	const op = "ChatsRepository.UpdateChatAvatar"
	const query = "UPDATE chat avatar"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: begin transaction: status: %s", query, queryStatus)
		return err
	}
	defer tx.Rollback(ctx)

	// Вставляем запись в таблицу attachment
	logger.WithField("attachment_id", attachmentID.String()).WithField("file_size", fileSize).Debug("Inserting into attachment table")
	_, err = tx.Exec(ctx, insertChatAvatarInAttachmentTableQuery, attachmentID, "avatar_"+attachmentID.String(), fileSize, "inline")
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert attachment: status: %s", query, queryStatus)
		return err
	}

	// Вставляем запись в таблицу avatar_chat
	logger.WithField("attachment_id", attachmentID.String()).WithField("chat_id", chatID.String()).Debug("Inserting into avatar_chat table")
	_, err = tx.Exec(ctx, insertChatAvatarInAvatarChatTableQuery, chatID, attachmentID)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert chat avatar: status: %s", query, queryStatus)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: commit transaction: status: %s", query, queryStatus)
		return err
	}

	logger.Info("Chat avatar updated successfully")
	return nil
}
