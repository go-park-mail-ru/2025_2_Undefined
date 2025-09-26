package service

import (
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/dto"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockChatsRepo реализует интерфейс ChatsRepository, который используется в ChatsService
type MockChatsRepo struct {
	GetChatsFunc               func(userId uuid.UUID) ([]models.Chat, error)
	GetLastMessagesOfChatsFunc func(userId uuid.UUID) ([]models.Message, error)
	GetChatFunc                func(userId, chatId uuid.UUID) (models.Chat, error)
	GetMessagesOfChatFunc      func(chatId uuid.UUID, limit, offset int) ([]models.Message, error)
	GetUsersOfChatFunc         func(chatId uuid.UUID) ([]models.UserInfo, error)
	GetUserInfoFunc            func(userId, chatId uuid.UUID) (models.UserInfo, error)
	CreateChatFunc             func(chat models.Chat, users []models.UserInfo) error
}

func (m *MockChatsRepo) GetChats(userId uuid.UUID) ([]models.Chat, error) {
	return m.GetChatsFunc(userId)
}
func (m *MockChatsRepo) GetLastMessagesOfChats(userId uuid.UUID) ([]models.Message, error) {
	return m.GetLastMessagesOfChatsFunc(userId)
}
func (m *MockChatsRepo) GetChat(userId, chatId uuid.UUID) (models.Chat, error) {
	return m.GetChatFunc(userId, chatId)
}
func (m *MockChatsRepo) GetMessagesOfChat(chatId uuid.UUID, limit, offset int) ([]models.Message, error) {
	return m.GetMessagesOfChatFunc(chatId, limit, offset)
}
func (m *MockChatsRepo) GetUsersOfChat(chatId uuid.UUID) ([]models.UserInfo, error) {
	return m.GetUsersOfChatFunc(chatId)
}
func (m *MockChatsRepo) GetUserInfo(userId, chatId uuid.UUID) (models.UserInfo, error) {
	return m.GetUserInfoFunc(userId, chatId)
}
func (m *MockChatsRepo) CreateChat(chat models.Chat, users []models.UserInfo) error {
	return m.CreateChatFunc(chat, users)
}

func TestGetChats_Success(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()
	mockRepo := &MockChatsRepo{
		GetChatsFunc: func(userId uuid.UUID) ([]models.Chat, error) {
			return []models.Chat{{ID: chatId, Name: "TestChat"}}, nil
		},
		GetLastMessagesOfChatsFunc: func(userId uuid.UUID) ([]models.Message, error) {
			return []models.Message{{
				UserID:    userId,
				Text:      "Hello",
				CreatedAt: time.Now(),
			}}, nil
		},
	}
	service := NewChatsService(mockRepo)
	chats, err := service.GetChats(userId)
	assert.NoError(t, err)
	assert.Len(t, chats, 1)
	assert.Equal(t, chatId, chats[0].ID)
	assert.Equal(t, "TestChat", chats[0].Name)
	assert.Equal(t, "Hello", chats[0].LastMessage.Text)
}

func TestGetChats_Error(t *testing.T) {
	mockRepo := &MockChatsRepo{
		GetChatsFunc: func(userId uuid.UUID) ([]models.Chat, error) {
			return nil, errors.New("repo error")
		},
	}
	service := NewChatsService(mockRepo)
	_, err := service.GetChats(uuid.New())
	assert.Error(t, err)
}

func TestGetInformationAboutChat_Success(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()
	mockRepo := &MockChatsRepo{
		GetChatFunc: func(userId, chatId uuid.UUID) (models.Chat, error) {
			return models.Chat{ID: chatId, Name: "Chat1", Type: models.ChatDialog}, nil
		},
		GetMessagesOfChatFunc: func(chatId uuid.UUID, limit, offset int) ([]models.Message, error) {
			return []models.Message{{
				UserID:    userId,
				Text:      "Hi",
				CreatedAt: time.Now(),
			}}, nil
		},
		GetUsersOfChatFunc: func(chatId uuid.UUID) ([]models.UserInfo, error) {
			return []models.UserInfo{{
				UserID: userId,
				Role:   models.RoleAdmin,
			}}, nil
		},
		GetUserInfoFunc: func(userId, chatId uuid.UUID) (models.UserInfo, error) {
			return models.UserInfo{UserID: userId, Role: models.RoleAdmin}, nil
		},
	}
	service := NewChatsService(mockRepo)
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
	mockRepo := &MockChatsRepo{
		GetChatFunc: func(userId, chatId uuid.UUID) (models.Chat, error) {
			return models.Chat{}, errors.New("not found")
		},
	}
	service := NewChatsService(mockRepo)
	info, err := service.GetInformationAboutChat(uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestCreateChat_Success(t *testing.T) {
	mockRepo := &MockChatsRepo{
		CreateChatFunc: func(chat models.Chat, users []models.UserInfo) error {
			return nil
		},
	}
	service := NewChatsService(mockRepo)
	chatDTO := dto.ChatCreateInformationDTO{
		Name: "NewChat",
		Type: models.ChatDialog,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: models.RoleAdmin},
		},
	}
	id, err := service.CreateChat(chatDTO)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestCreateChat_Error(t *testing.T) {
	mockRepo := &MockChatsRepo{
		CreateChatFunc: func(chat models.Chat, users []models.UserInfo) error {
			return errors.New("fail")
		},
	}
	service := NewChatsService(mockRepo)
	chatDTO := dto.ChatCreateInformationDTO{
		Name: "FailChat",
		Type: models.ChatDialog,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: models.RoleAdmin},
		},
	}
	id, err := service.CreateChat(chatDTO)
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}
