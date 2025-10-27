package message

import (
	"context"

	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"

	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
)

type MockMessageRepo struct {
	InsertMessageFunc func(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error)
}

func (r *MockMessageRepo) InsertMessage(ctx context.Context, msg modelsMessage.CreateMessage) (uuid.UUID, error) {
	return r.InsertMessageFunc(ctx, msg)
}

type MockUserRepo struct {
	GetUserByIDFunc   func(ctx context.Context, id uuid.UUID) (*UserModels.User, error)
	GetUsersNamesFunc func(ctx context.Context, usersIds []uuid.UUID) ([]string, error)
}

func (r *MockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error) {
	return r.GetUserByIDFunc(ctx, id)
}

func (r *MockUserRepo) GetUsersNames(ctx context.Context, usersIds []uuid.UUID) ([]string, error) {
	return r.GetUsersNamesFunc(ctx, usersIds)
}

type MockListenerMap struct {
	SubscribeUserToChatFunc  func(userId uuid.UUID, chatId uuid.UUID) <-chan dtoMessage.MessageDTO
	GetChatListenersFunc     func(chatId uuid.UUID) map[uuid.UUID]chan dtoMessage.MessageDTO
	CloseAllFunc             func()
	CleanInactiveChatsFunc   func() int
	CleanInactiveReadersFunc func() int
}

func (lm *MockListenerMap) SubscribeUserToChat(userId uuid.UUID, chatId uuid.UUID) <-chan dtoMessage.MessageDTO {
	return lm.SubscribeUserToChatFunc(userId, chatId)
}

func (lm *MockListenerMap) GetChatListeners(chatId uuid.UUID) map[uuid.UUID]chan dtoMessage.MessageDTO {
	return lm.GetChatListenersFunc(chatId)
}

func (lm *MockListenerMap) CloseAll() {
	lm.CloseAllFunc()
}

func (lm *MockListenerMap) CleanInactiveChats() int {
	return lm.CleanInactiveChatsFunc()
}

func (lm *MockListenerMap) CleanInactiveReaders() int {
	return lm.CleanInactiveReadersFunc()
}
