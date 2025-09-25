package repository

import (
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/google/uuid"
)

type ChatsRepository interface {
	GetChats(userId uuid.UUID) ([]models.Chat, error)
	GetLastMessagesOfChats(userId uuid.UUID) ([]models.Message, error)
	GetChat(userId, chatId uuid.UUID) (models.Chat, error)
	GetUsersOfChat(chatId uuid.UUID) ([]models.UserInfo, error)
	GetMessagesOfChat(chatId uuid.UUID, limit, offset int) ([]models.Message, error)
	CreateChat(chat models.Chat, usersInfo []models.UserInfo) error
	GetUserInfo(userId, chatId uuid.UUID) (models.UserInfo, error)
}
