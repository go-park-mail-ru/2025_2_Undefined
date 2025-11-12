package message

import (
	"context"
	"maps"
	"sync"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ListenersMap реализует ListenerMapInterface с безопасным доступом для конкурентных операций
type ListenerMap struct {
	logger *logrus.Entry
	mu     sync.RWMutex
	// хранит словарь слушателей сообщений: chatID -> connectionID -> chan WebSocketMessageDTO
	data map[uuid.UUID]map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO
	// хранит словарь соединений пользователя: userID -> connectionID -> chan WebSocketMessageDTO
	userConnections map[uuid.UUID]map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO
	// хранит словарь соединения и канала, в который сливаются сообщения: connectionID -> chan WebSocketMessageDTO
	outgoingChannels map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO
}

func NewListenerMap() *ListenerMap {
	logger := domains.GetLogger(context.Background())

	return &ListenerMap{
		data:             make(map[uuid.UUID]map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO),
		userConnections:  make(map[uuid.UUID]map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO),
		outgoingChannels: make(map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO),
		logger:           logger,
	}
}

func (lm *ListenerMap) SubscribeConnectionToChat(connectionID uuid.UUID, chatID, userID uuid.UUID) <-chan dtoMessage.WebSocketMessageDTO {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.data[chatID] == nil {
		lm.data[chatID] = make(map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO)
	}

	ch, ok := lm.data[chatID][connectionID]
	if !ok {
		ch = make(chan dtoMessage.WebSocketMessageDTO, MessagesBufferForOneUserChat)
		lm.data[chatID][connectionID] = ch
	}

	// Добавляем ко всем соединениям пользователя новое соединение
	if lm.userConnections[userID] == nil {
		lm.userConnections[userID] = make(map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO)
	}

	lm.userConnections[userID][connectionID] = lm.data[chatID][connectionID]

	return ch
}

func (lm *ListenerMap) GetChatListeners(chatId uuid.UUID) map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	chatMap, ok := lm.data[chatId]
	if !ok {
		return nil
	}

	// Копируем, чтобы не было гонки данных
	result := make(map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO)
	maps.Copy(result, chatMap)

	return result
}

// Добавляет чат к подпискам пользователя и возвращает все его соединения
func (lm *ListenerMap) AddChatToUserSubscription(userID, chatID uuid.UUID) map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	result := make(map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO)

	// Если у пользователя нет соединений, тогда ничего не делаем для него
	if lm.userConnections[userID] == nil {
		return result
	}

	for connectionID := range lm.userConnections[userID] {

		// Аналогично функции SubscribeConnectionToChat добавляем соединение.
		if lm.data[chatID] == nil {
			lm.data[chatID] = make(map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO)
		} else {
			lm.logger.Warningf("AddChatToUserSubscription: chat %s already exists for user %s!", chatID, userID)
		}

		ch, ok := lm.data[chatID][connectionID]
		if !ok {
			ch = make(chan dtoMessage.WebSocketMessageDTO, MessagesBufferForOneUserChat)
			lm.data[chatID][connectionID] = ch
		}

		result[connectionID] = ch
	}

	return result
}

func (lm *ListenerMap) GetOutgoingChannel(connectionID uuid.UUID) chan dtoMessage.WebSocketMessageDTO {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	ch, ok := lm.outgoingChannels[connectionID]
	if !ok {
		ch = make(chan dtoMessage.WebSocketMessageDTO, MessagesBufferForAllUserChats)
		lm.outgoingChannels[connectionID] = ch
	}

	return ch
}

func (lm *ListenerMap) CloseAll() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	for chatId, connections := range lm.data {
		for connectionId, ch := range connections {
			close(ch)
			lm.logger.Infof("Closed channel for connection %d in chat %d", connectionId, chatId)
		}
	}

	lm.data = make(map[uuid.UUID]map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO)
}

func (lm *ListenerMap) CleanInactiveChats() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	cleaned := 0
	for chatId, connections := range lm.data {
		if len(connections) == 0 {
			delete(lm.data, chatId)
			cleaned++
			lm.logger.Infof("Cleaned up inactive chat %d", chatId)
		}
	}

	return cleaned
}

func (lm *ListenerMap) CleanInactiveReaders() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	cleanedCount := 0

	for chatId, connections := range lm.data {
		for connectionId, ch := range connections {
			// Проверяем, заполнен ли буфер канала (читатель не успевает обрабатывать)
			if len(ch) >= cap(ch) {
				close(ch)
				delete(connections, connectionId)
				cleanedCount++
				lm.logger.Infof("Cleaned inactive reader: connection %s in chat %s", connectionId, chatId)
			}
		}

		// Если в чате не осталось соединений, то очищаем запись
		if len(connections) == 0 {
			delete(lm.data, chatId)
			lm.logger.Infof("Removed empty chat %s after cleaning readers", chatId)
		}
	}

	lm.logger.Infof("Cleaned %d inactive readers", cleanedCount)
	return cleanedCount
}
