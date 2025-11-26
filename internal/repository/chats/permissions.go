package repository

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
)

func (r *ChatsRepository) CheckUserHasRole(ctx context.Context, userId, chatId uuid.UUID, role string) (bool, error) {
	const op = "ChatsRepository.CheckUserHasRole"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userId.String()).WithField("chat_id", chatId.String()).WithField("role", role)
	logger.Debug("Starting database operation: check user role in chat")

	var hasRole bool
	err := r.db.QueryRow(ctx, checkUserRoleQuery, userId, chatId, role).Scan(&hasRole)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check user role query")
		return false, err
	}

	logger.WithField("has_role", hasRole).Info("Database operation completed successfully: user role checked")
	return hasRole, nil
}
