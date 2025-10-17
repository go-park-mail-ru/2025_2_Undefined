package message

import (
	"testing"
	"time"

	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageUsecase_NewListenerMap(t *testing.T) {
	lm := NewListenerMap()
	assert.NotNil(t, lm)
	assert.NotNil(t, lm.data)
}

func TestMessageUsecase_ListenerMap_SubscribeUserToChat(t *testing.T) {
	lm := NewListenerMap()
	userId := uuid.New()
	chatId := uuid.New()

	// Первая подписка создает новый канал
	ch1 := lm.SubscribeUserToChat(userId, chatId)
	assert.NotNil(t, ch1)

	// Повторная подписка того же пользователя возвращает тот же канал
	ch2 := lm.SubscribeUserToChat(userId, chatId)
	assert.Equal(t, ch1, ch2)

	// Проверяем через внутренние данные, что канал создан с правильным размером буфера
	lm.mu.RLock()
	internalCh := lm.data[chatId][userId]
	lm.mu.RUnlock()
	assert.Equal(t, MessagesBufferForOneUserChat, cap(internalCh))
}

func TestMessageUsecase_ListenerMap_SubscribeUserToChat_MultipleChatsSameUser(t *testing.T) {
	lm := NewListenerMap()
	userId := uuid.New()
	chatId1 := uuid.New()
	chatId2 := uuid.New()

	ch1 := lm.SubscribeUserToChat(userId, chatId1)
	ch2 := lm.SubscribeUserToChat(userId, chatId2)

	// Каналы должны быть разными для разных чатов
	assert.NotEqual(t, ch1, ch2)
}

func TestMessageUsecase_ListenerMap_SubscribeUserToChat_MultipleUsersSameChat(t *testing.T) {
	lm := NewListenerMap()
	userId1 := uuid.New()
	userId2 := uuid.New()
	chatId := uuid.New()

	ch1 := lm.SubscribeUserToChat(userId1, chatId)
	ch2 := lm.SubscribeUserToChat(userId2, chatId)

	// Каналы должны быть разными для разных пользователей
	assert.NotEqual(t, ch1, ch2)
}

func TestMessageUsecase_ListenerMap_GetChatListeners(t *testing.T) {
	lm := NewListenerMap()
	chatId := uuid.New()

	// Для несуществующего чата возвращается nil
	listeners := lm.GetChatListeners(chatId)
	assert.Nil(t, listeners)

	// Добавляем пользователей в чат
	userId1 := uuid.New()
	userId2 := uuid.New()
	_ = lm.SubscribeUserToChat(userId1, chatId)
	_ = lm.SubscribeUserToChat(userId2, chatId)

	// Получаем слушателей
	listeners = lm.GetChatListeners(chatId)
	require.NotNil(t, listeners)
	assert.Len(t, listeners, 2)
	// Проверяем наличие пользователей в мапе
	assert.Contains(t, listeners, userId1)
	assert.Contains(t, listeners, userId2)

	// Проверяем, что возвращается копия
	delete(listeners, userId1)
	listenersAgain := lm.GetChatListeners(chatId)
	assert.Len(t, listenersAgain, 2) // Оригинальная мапа не изменилась
}

func TestMessageUsecase_ListenerMap_CloseAll(t *testing.T) {
	lm := NewListenerMap()
	userId1 := uuid.New()
	userId2 := uuid.New()
	chatId1 := uuid.New()
	chatId2 := uuid.New()

	// Подписываем пользователей на чаты
	ch1 := lm.SubscribeUserToChat(userId1, chatId1)
	ch2 := lm.SubscribeUserToChat(userId2, chatId1)
	ch3 := lm.SubscribeUserToChat(userId1, chatId2)

	// Проверяем, что есть подписчики перед закрытием
	listeners := lm.GetChatListeners(chatId1)
	assert.Len(t, listeners, 2)

	// Закрываем все каналы
	lm.CloseAll()

	// Проверяем, что каналы закрыты
	_, ok1 := <-ch1
	_, ok2 := <-ch2
	_, ok3 := <-ch3
	assert.False(t, ok1)
	assert.False(t, ok2)
	assert.False(t, ok3)

	// Проверяем, что данные очищены
	listeners = lm.GetChatListeners(chatId1)
	assert.Nil(t, listeners)
}

func TestMessageUsecase_ListenerMap_CleanInactiveChats(t *testing.T) {
	lm := NewListenerMap()
	chatId1 := uuid.New()
	chatId2 := uuid.New()
	userId := uuid.New()

	// Создаем чат с пользователем
	lm.SubscribeUserToChat(userId, chatId1)

	// Создаем пустой чат (имитируем ситуацию, когда пользователи ушли)
	lm.data[chatId2] = make(map[uuid.UUID]chan dtoMessage.MessageDTO)

	// Проверяем, что есть 2 чата
	assert.Len(t, lm.data, 2)

	// Очищаем неактивные чаты
	cleaned := lm.CleanInactiveChats()

	// Должен быть очищен 1 пустой чат
	assert.Equal(t, 1, cleaned)
	assert.Len(t, lm.data, 1)
	assert.Contains(t, lm.data, chatId1)
	assert.NotContains(t, lm.data, chatId2)
}

func TestMessageUsecase_ListenerMap_CleanInactiveReaders(t *testing.T) {
	lm := NewListenerMap()
	chatId := uuid.New()
	activeUserId := uuid.New()
	inactiveUserId := uuid.New()

	// Подписываем активного пользователя
	_ = lm.SubscribeUserToChat(activeUserId, chatId)

	// Подписываем неактивного пользователя
	inactiveCh := lm.SubscribeUserToChat(inactiveUserId, chatId)

	// Получаем доступ к внутреннему каналу и заполняем его буфер до максимума
	lm.mu.Lock()
	inactiveChBuffered := lm.data[chatId][inactiveUserId]
	for i := 0; i < cap(inactiveChBuffered); i++ {
		inactiveChBuffered <- dtoMessage.MessageDTO{}
	}
	lm.mu.Unlock()

	// Проверяем, что в чате 2 пользователя
	listeners := lm.GetChatListeners(chatId)
	assert.Len(t, listeners, 2)

	// Очищаем неактивных читателей
	cleaned := lm.CleanInactiveReaders()

	// Должен быть удален 1 неактивный читатель
	assert.Equal(t, 1, cleaned)

	// Проверяем, что остался только активный пользователь
	listeners = lm.GetChatListeners(chatId)
	require.NotNil(t, listeners)
	assert.Len(t, listeners, 1)
	assert.Contains(t, listeners, activeUserId)
	assert.NotContains(t, listeners, inactiveUserId)

	// Проверяем, что канал неактивного пользователя закрыт
	// Сначала вычитываем все сообщения из буфера
	for len(inactiveCh) > 0 {
		<-inactiveCh
	}
	// Теперь проверяем, что канал закрыт
	_, ok := <-inactiveCh
	assert.False(t, ok)
}

func TestMessageUsecase_ListenerMap_CleanInactiveReaders_EmptyChat(t *testing.T) {
	lm := NewListenerMap()
	chatId := uuid.New()
	userId := uuid.New()

	// Подписываем пользователя
	ch := lm.SubscribeUserToChat(userId, chatId)

	// Получаем доступ к внутреннему каналу и заполняем его буфер до максимума
	lm.mu.Lock()
	chBuffered := lm.data[chatId][userId]
	for i := 0; i < cap(chBuffered); i++ {
		chBuffered <- dtoMessage.MessageDTO{}
	}
	lm.mu.Unlock()

	// Проверяем, что чат существует
	assert.Contains(t, lm.data, chatId)

	// Очищаем неактивных читателей
	cleaned := lm.CleanInactiveReaders()

	// Должен быть удален 1 читатель и весь чат
	assert.Equal(t, 1, cleaned)
	assert.NotContains(t, lm.data, chatId)

	// Проверяем, что канал закрыт
	// Сначала вычитываем все сообщения из буфера
	for len(ch) > 0 {
		<-ch
	}
	// Теперь проверяем, что канал закрыт
	_, ok := <-ch
	assert.False(t, ok)
}

func TestMessageUsecase_ListenerMap_ConcurrentAccess(t *testing.T) {
	lm := NewListenerMap()
	chatId := uuid.New()

	// Тестируем concurrent доступ
	done := make(chan bool, 10)

	// Запускаем несколько горутин для подписки
	for i := 0; i < 5; i++ {
		go func(i int) {
			userId := uuid.New()
			ch := lm.SubscribeUserToChat(userId, chatId)
			assert.NotNil(t, ch)
			done <- true
		}(i)
	}

	// Запускаем горутины для получения слушателей
	for i := 0; i < 5; i++ {
		go func() {
			listeners := lm.GetChatListeners(chatId)
			_ = listeners // Используем результат
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Goroutine didn't finish in time")
		}
	}
}

func TestMessageUsecase_ListenerMap_CleanInactiveChats_NoChats(t *testing.T) {
	lm := NewListenerMap()

	// Очищаем пустую мапу
	cleaned := lm.CleanInactiveChats()
	assert.Equal(t, 0, cleaned)
}

func TestMessageUsecase_ListenerMap_CleanInactiveReaders_NoReaders(t *testing.T) {
	lm := NewListenerMap()

	// Очищаем пустую мапу
	cleaned := lm.CleanInactiveReaders()
	assert.Equal(t, 0, cleaned)
}
