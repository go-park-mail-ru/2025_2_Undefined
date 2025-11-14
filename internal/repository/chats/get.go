package repository

import (
	"context"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
)

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

func (r *ChatsRepository) GetChat(ctx context.Context, chatID uuid.UUID) (*modelsChats.Chat, error) {
	const op = "ChatsRepository.GetChat"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatID.String())
	logger.Debug("Starting database operation: get specific chat")

	chat := &modelsChats.Chat{}

	err := r.db.QueryRow(getChatQuery, chatID).
		Scan(&chat.ID, &chat.Type, &chat.Name, &chat.Description)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get chat query")
		return nil, err
	}

	logger.Info("Database operation completed successfully: chat retrieved")
	return chat, nil
}

func (r *ChatsRepository) GetChatWithAvatar(ctx context.Context, chatID, userID uuid.UUID) (*modelsChats.Chat, error) {
	const op = "ChatsRepository.GetChatWithAvatar"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_id", chatID.String()).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: get specific chat with avatar")

	chat := &modelsChats.Chat{}

	err := r.db.QueryRow(getChatWithAvatarQuery, chatID).
		Scan(&chat.ID, &chat.Type, &chat.Name, &chat.Description, &chat.AvatarID)

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
