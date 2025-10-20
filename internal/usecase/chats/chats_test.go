package usecase

import (
	"errors"
	"testing"
	"time"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	modelsUsers "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockChatsRepo реализует интерфейс ChatsRepository, который используется в ChatsService
type MockChatsRepo struct {
	GetChatsFunc               func(userId uuid.UUID) ([]modelsChats.Chat, error)
	GetLastMessagesOfChatsFunc func(userId uuid.UUID) ([]modelsMessage.Message, error)
	GetChatFunc                func(userId, chatId uuid.UUID) (*modelsChats.Chat, error)
	GetMessagesOfChatFunc      func(chatId uuid.UUID, offset, limit int) ([]modelsMessage.Message, error)
	GetUsersOfChatFunc         func(chatId uuid.UUID) ([]modelsChats.UserInfo, error)
	GetUserInfoFunc            func(userId, chatId uuid.UUID) (*modelsChats.UserInfo, error)
	CreateChatFunc             func(chat modelsChats.Chat, usersInfo []modelsChats.UserInfo, usersNames []string) error
}

// MockUserRepo реализует интерфейс UserRepository, который используется в ChatsService
type MockUserRepo struct {
	GetUsersNamesFunc func(usersIds []uuid.UUID) ([]string, error)
}

func (m *MockUserRepo) GetUsersNames(usersIds []uuid.UUID) ([]string, error) {
	return m.GetUsersNamesFunc(usersIds)
}

func (m *MockUserRepo) GetUserByID(userId uuid.UUID) (*modelsUsers.User, error) {
	return &modelsUsers.User{}, nil
}

func (m *MockChatsRepo) GetChats(userId uuid.UUID) ([]modelsChats.Chat, error) {
	return m.GetChatsFunc(userId)
}
func (m *MockChatsRepo) GetLastMessagesOfChats(userId uuid.UUID) ([]modelsMessage.Message, error) {
	return m.GetLastMessagesOfChatsFunc(userId)
}
func (m *MockChatsRepo) GetChat(userId, chatId uuid.UUID) (*modelsChats.Chat, error) {
	return m.GetChatFunc(userId, chatId)
}
func (m *MockChatsRepo) GetMessagesOfChat(chatId uuid.UUID, offset, limit int) ([]modelsMessage.Message, error) {
	return m.GetMessagesOfChatFunc(chatId, offset, limit)
}
func (m *MockChatsRepo) GetUsersOfChat(chatId uuid.UUID) ([]modelsChats.UserInfo, error) {
	return m.GetUsersOfChatFunc(chatId)
}
func (m *MockChatsRepo) GetUserInfo(userId, chatId uuid.UUID) (*modelsChats.UserInfo, error) {
	return m.GetUserInfoFunc(userId, chatId)
}
func (m *MockChatsRepo) CreateChat(chat modelsChats.Chat, usersInfo []modelsChats.UserInfo, usersNames []string) error {
	return m.CreateChatFunc(chat, usersInfo, usersNames)
}

func TestGetChats_Success(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()
	mockChatsRepo := &MockChatsRepo{
		GetChatsFunc: func(userId uuid.UUID) ([]modelsChats.Chat, error) {
			return []modelsChats.Chat{{ID: chatId, Name: "TestChat"}}, nil
		},
		GetLastMessagesOfChatsFunc: func(userId uuid.UUID) ([]modelsMessage.Message, error) {
			return []modelsMessage.Message{{
				ChatID:    chatId,
				UserID:    userId,
				Text:      "Hello",
				CreatedAt: time.Now(),
			}}, nil
		},
	}
	mockUserRepo := &MockUserRepo{}
	service := NewChatsService(mockChatsRepo, mockUserRepo)
	chats, err := service.GetChats(userId)
	assert.NoError(t, err)
	assert.Len(t, chats, 1)
	assert.Equal(t, chatId, chats[0].ID)
	assert.Equal(t, "TestChat", chats[0].Name)
	assert.Equal(t, "Hello", chats[0].LastMessage.Text)
}

func TestGetChats_Error(t *testing.T) {
	mockChatsRepo := &MockChatsRepo{
		GetChatsFunc: func(userId uuid.UUID) ([]modelsChats.Chat, error) {
			return nil, errors.New("repo error")
		},
	}
	mockUserRepo := &MockUserRepo{}
	service := NewChatsService(mockChatsRepo, mockUserRepo)
	_, err := service.GetChats(uuid.New())
	assert.Error(t, err)
}

func TestGetInformationAboutChat_Success(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()
	mockChatsRepo := &MockChatsRepo{
		GetChatFunc: func(userId, chatId uuid.UUID) (*modelsChats.Chat, error) {
			return &modelsChats.Chat{ID: chatId, Name: "Chat1", Type: modelsChats.ChatTypeDialog}, nil
		},
		GetMessagesOfChatFunc: func(chatId uuid.UUID, offset, limit int) ([]modelsMessage.Message, error) {
			return []modelsMessage.Message{{
				UserID:    userId,
				Text:      "Hi",
				CreatedAt: time.Now(),
			}}, nil
		},
		GetUsersOfChatFunc: func(chatId uuid.UUID) ([]modelsChats.UserInfo, error) {
			return []modelsChats.UserInfo{{
				UserID: userId,
				Role:   modelsChats.RoleAdmin,
			}}, nil
		},
		GetUserInfoFunc: func(userId, chatId uuid.UUID) (*modelsChats.UserInfo, error) {
			return &modelsChats.UserInfo{UserID: userId, Role: modelsChats.RoleAdmin}, nil
		},
	}
	mockUserRepo := &MockUserRepo{}
	service := NewChatsService(mockChatsRepo, mockUserRepo)
	info, err := service.GetInformationAboutChat(userId, chatId)
	assert.NoError(t, err)
	assert.Equal(t, chatId, info.ID)
	assert.True(t, info.IsAdmin)
	assert.True(t, info.CanChat)
	assert.True(t, info.IsMember)
	assert.True(t, info.IsPrivate)
	assert.Len(t, info.Members, 1)
	assert.Len(t, info.Messages, 1)
}

func TestGetInformationAboutChat_Error(t *testing.T) {
	mockChatsRepo := &MockChatsRepo{
		GetChatFunc: func(userId, chatId uuid.UUID) (*modelsChats.Chat, error) {
			return nil, errors.New("not found")
		},
	}
	mockUserRepo := &MockUserRepo{}
	service := NewChatsService(mockChatsRepo, mockUserRepo)
	info, err := service.GetInformationAboutChat(uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestCreateChat_Success(t *testing.T) {
	mockChatsRepo := &MockChatsRepo{
		CreateChatFunc: func(chat modelsChats.Chat, usersInfo []modelsChats.UserInfo, usersNames []string) error {
			return nil
		},
	}
	mockUserRepo := &MockUserRepo{
		GetUsersNamesFunc: func(usersIds []uuid.UUID) ([]string, error) {
			return []string{"TestUser"}, nil
		},
	}
	service := NewChatsService(mockChatsRepo, mockUserRepo)
	chatDTO := dto.ChatCreateInformationDTO{
		Name: "NewChat",
		Type: modelsChats.ChatTypeDialog,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: modelsChats.RoleAdmin},
		},
	}
	id, err := service.CreateChat(chatDTO)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestCreateChat_Error(t *testing.T) {
	mockChatsRepo := &MockChatsRepo{
		CreateChatFunc: func(chat modelsChats.Chat, usersInfo []modelsChats.UserInfo, usersNames []string) error {
			return errors.New("fail")
		},
	}
	mockUserRepo := &MockUserRepo{
		GetUsersNamesFunc: func(usersIds []uuid.UUID) ([]string, error) {
			return []string{"TestUser"}, nil
		},
	}
	service := NewChatsService(mockChatsRepo, mockUserRepo)
	chatDTO := dto.ChatCreateInformationDTO{
		Name: "FailChat",
		Type: modelsChats.ChatTypeDialog,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: modelsChats.RoleAdmin},
		},
	}
	id, err := service.CreateChat(chatDTO)
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}
