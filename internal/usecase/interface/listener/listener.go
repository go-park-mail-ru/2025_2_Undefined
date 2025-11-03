package listener

import (
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
)

type ListenerMapInterface interface {
	SubscribeUserToChat(userId uuid.UUID, chatId uuid.UUID) <-chan message.MessageDTO
	GetChatListeners(chatId uuid.UUID) map[uuid.UUID]chan message.MessageDTO
	CloseAll()
	CleanInactiveChats() int
	CleanInactiveReaders() int
}
