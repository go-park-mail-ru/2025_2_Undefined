package message

import (
	"log"
	"maps"
	"sync"

	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
)

// ListenersMap реализует ListenerMapInterface с безопасным доступом для конкурентных операций
type ListenerMap struct {
	mu sync.RWMutex
	// хранит словарь слушателей сообщений: chatID -> userID -> chan MessageDTO
	data map[uuid.UUID]map[uuid.UUID]chan dtoMessage.MessageDTO
}

func NewListenerMap() *ListenerMap {
	return &ListenerMap{
		data: make(map[uuid.UUID]map[uuid.UUID]chan dtoMessage.MessageDTO),
	}
}

func (lm *ListenerMap) SubscribeUserToChat(userId uuid.UUID, chatId uuid.UUID) <-chan dtoMessage.MessageDTO {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.data[chatId] == nil {
		lm.data[chatId] = make(map[uuid.UUID]chan dtoMessage.MessageDTO)
	}

	ch, ok := lm.data[chatId][userId]
	if !ok {
		ch = make(chan dtoMessage.MessageDTO, MessagesBufferForOneUserChat)
		lm.data[chatId][userId] = ch
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

	for chatId, users := range lm.data {
		for userId, ch := range users {
			close(ch)
			log.Printf("Closed channel for user %d in chat %d", userId, chatId)
		}
	}

	lm.data = make(map[uuid.UUID]map[uuid.UUID]chan dtoMessage.MessageDTO)
}

func (lm *ListenerMap) CleanInactiveChats() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	cleaned := 0
	for chatId, users := range lm.data {
		if len(users) == 0 {
			delete(lm.data, chatId)
			cleaned++
			log.Printf("Cleaned up inactive chat %d", chatId)
		}
	}

	return cleaned
}

func (lm *ListenerMap) CleanInactiveReaders() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	cleanedCount := 0

	for chatId, users := range lm.data {
		for userId, ch := range users {
			// Проверяем, заполнен ли буфер канала (читатель не успевает обрабатывать)
			if len(ch) >= cap(ch) {
				close(ch)
				delete(users, userId)
				cleanedCount++
				log.Printf("Cleaned inactive reader: user %s in chat %s", userId, chatId)
			}
		}

		// Если в чате не осталось пользователей, удаляем сам чат
		if len(users) == 0 {
			delete(lm.data, chatId)
			log.Printf("Removed empty chat %s after cleaning readers", chatId)
		}
	}

	log.Printf("Cleaned %d inactive readers", cleanedCount)
	return cleanedCount
}
