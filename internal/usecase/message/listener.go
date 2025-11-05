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
	// хранит словарь слушателей сообщений: chatID -> connectionID -> chan MessageDTO
	data map[uuid.UUID]map[uuid.UUID]chan dtoMessage.MessageDTO
}

func NewListenerMap() *ListenerMap {
	logger := domains.GetLogger(context.Background())

	return &ListenerMap{
		data:   make(map[uuid.UUID]map[uuid.UUID]chan dtoMessage.MessageDTO),
		logger: logger,
	}
}

func (lm *ListenerMap) SubscribeConnectionToChat(connectionID uuid.UUID, chatId uuid.UUID) <-chan dtoMessage.MessageDTO {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.data[chatId] == nil {
		lm.data[chatId] = make(map[uuid.UUID]chan dtoMessage.MessageDTO)
	}

	ch, ok := lm.data[chatId][connectionID]
	if !ok {
		ch = make(chan dtoMessage.MessageDTO, MessagesBufferForOneUserChat)
		lm.data[chatId][connectionID] = ch
	}

	return ch
}

func (lm *ListenerMap) GetChatListeners(chatId uuid.UUID) map[uuid.UUID]chan dtoMessage.MessageDTO {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	chatMap, ok := lm.data[chatId]
	if !ok {
		return nil
	}

	// Копируем, чтобы не было гонки данных
	result := make(map[uuid.UUID]chan dtoMessage.MessageDTO)
	maps.Copy(result, chatMap)

	return result
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

	lm.data = make(map[uuid.UUID]map[uuid.UUID]chan dtoMessage.MessageDTO)
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
