package message

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessageUsecase_WorkerDistribute(t *testing.T) {
	mockMessageRepo := &MockMessageRepo{}
	mockUserRepo := &MockUserRepo{}

	cnt := atomic.Int32{}

	userChannels := make(map[uuid.UUID]chan message.MessageDTO)
	for i := 0; i < 3; i++ {
		userId := uuid.New()
		ch := make(chan message.MessageDTO, 1)
		userChannels[userId] = ch
		go func() {
			<-ch
			cnt.Add(1)
		}()
	}

	mockListenerMap := &MockListenerMap{
		GetChatListenersFunc: func(chatId uuid.UUID) map[uuid.UUID]chan message.MessageDTO {
			return userChannels
		},
	}

	uc := NewMessageUsecase(mockMessageRepo, mockUserRepo, mockListenerMap)
	uc.distributeChannel <- message.MessageDTO{}

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, cnt.Load(), int32(3))
	uc.Stop()
}

func TestMessageUsecase_Distribute_NoListeners(t *testing.T) {
	mockMessageRepo := &MockMessageRepo{}
	mockUserRepo := &MockUserRepo{}

	mockListenerMap := &MockListenerMap{
		GetChatListenersFunc: func(chatId uuid.UUID) map[uuid.UUID]chan message.MessageDTO {
			return nil
		},
	}

	uc := NewMessageUsecase(mockMessageRepo, mockUserRepo, mockListenerMap)

	select {
	case uc.distributeChannel <- message.MessageDTO{ChatId: uuid.New()}:

	case <-time.After(500 * time.Millisecond):
		t.Fatal("sending to distributeChannel blocked")
	}

	time.Sleep(100 * time.Millisecond)
	uc.Stop()
}
