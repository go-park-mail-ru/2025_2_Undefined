package chats

import (
	"context"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/google/uuid"
)

type ChatsRepository interface {
	GetChats(ctx context.Context, userID uuid.UUID) ([]modelsChats.Chat, error)
	GetChat(ctx context.Context, chatID uuid.UUID) (*modelsChats.Chat, error)
	GetUsersOfChat(ctx context.Context, chatID uuid.UUID) ([]modelsChats.UserInfo, error)
	GetUsersDialog(ctx context.Context, user1ID, user2ID uuid.UUID) (uuid.UUID, error)
	CreateChat(ctx context.Context, chat modelsChats.Chat, usersInfo []modelsChats.UserInfo, usersNames []string) error
	GetUserInfo(ctx context.Context, userID, chatID uuid.UUID) (*modelsChats.UserInfo, error)
	InsertUsersToChat(ctx context.Context, chatID uuid.UUID, usersInfo []modelsChats.UserInfo) error
	CheckUserHasRole(ctx context.Context, userID, chatID uuid.UUID, role string) (bool, error)
	DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error
	UpdateChat(ctx context.Context, userID, chatID uuid.UUID, name, description string) error
	GetChatAvatars(ctx context.Context, chatIDs []uuid.UUID) (map[string]uuid.UUID, error)
	UpdateChatAvatar(ctx context.Context, chatID uuid.UUID, attachmentID uuid.UUID, fileSize int64) error
	SearchChats(ctx context.Context, userID uuid.UUID, name string) ([]modelsChats.Chat, error)
}
