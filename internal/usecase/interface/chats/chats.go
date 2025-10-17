package chats

import (
	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/google/uuid"
)

type ChatsRepository interface {
	GetChats(userId uuid.UUID) ([]modelsChats.Chat, error)
	GetLastMessagesOfChats(userId uuid.UUID) ([]modelsMessage.Message, error)
	GetChat(userId, chatId uuid.UUID) (*modelsChats.Chat, error)
	GetUsersOfChat(chatId uuid.UUID) ([]modelsChats.UserInfo, error)
	GetMessagesOfChat(chatId uuid.UUID, offset, limit int) ([]modelsMessage.Message, error)
	CreateChat(chat modelsChats.Chat, usersInfo []modelsChats.UserInfo, usersNames []string) error
	GetUserInfo(userId, chatId uuid.UUID) (*modelsChats.UserInfo, error)
}
