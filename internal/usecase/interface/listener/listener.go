package listener

import (
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
)

type ListenerMapInterface interface {
	SubscribeConnectionToChat(connectionID uuid.UUID, chatID, userID uuid.UUID) <-chan dto.WebSocketMessageDTO
	GetChatListeners(chatId uuid.UUID) map[uuid.UUID]chan dto.WebSocketMessageDTO
	AddChatToUserSubscription(userID, chatID uuid.UUID) map[uuid.UUID]chan dto.WebSocketMessageDTO
	GetOutgoingChannel(connectionID uuid.UUID) chan dto.WebSocketMessageDTO
	RegisterUserConnection(userID, connectionID uuid.UUID, outgoingChan chan dto.WebSocketMessageDTO)
	CloseAll()
	CleanInactiveChats() int
	CleanInactiveReaders() int
}
