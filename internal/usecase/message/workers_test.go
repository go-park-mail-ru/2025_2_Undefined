package message

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessageUsecase_WorkerDistribute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageRepo := mocks.NewMockMessageRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	mockListenerMap := mocks.NewMockListenerMapInterface(ctrl)

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

	// Настраиваем ожидания мока
	mockListenerMap.EXPECT().
		GetChatListeners(gomock.Any()).
		Return(userChannels).
		AnyTimes()

	uc := NewMessageUsecase(mockMessageRepo, mockUserRepo, mockChatsRepo, mockFileStorage, mockListenerMap)
	uc.distributeChannel <- message.MessageDTO{}

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, cnt.Load(), int32(3))
	uc.Stop()
}

func TestMessageUsecase_Distribute_NoListeners(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMessageRepo := mocks.NewMockMessageRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	mockListenerMap := mocks.NewMockListenerMapInterface(ctrl)

	// Настраиваем ожидания мока - возвращаем nil (нет слушателей)
	mockListenerMap.EXPECT().
		GetChatListeners(gomock.Any()).
		Return(nil).
		AnyTimes()

	uc := NewMessageUsecase(mockMessageRepo, mockUserRepo, mockChatsRepo, mockFileStorage, mockListenerMap)

	select {
	case uc.distributeChannel <- message.MessageDTO{ChatId: uuid.New()}:

	case <-time.After(500 * time.Millisecond):
		t.Fatal("sending to distributeChannel blocked")
	}

	time.Sleep(100 * time.Millisecond)
	uc.Stop()
}
