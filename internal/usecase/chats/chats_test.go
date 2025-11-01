package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChats(gomock.Any(), userId).
		Return([]modelsChats.Chat{{ID: chatId, Name: "TestChat"}}, nil)

	mockChatsRepo.EXPECT().
		GetLastMessagesOfChats(gomock.Any(), userId).
		Return([]modelsMessage.Message{{
			ChatID:    chatId,
			UserID:    userId,
			Text:      "Hello",
			CreatedAt: time.Now(),
		}}, nil)

	service := NewChatsService(mockChatsRepo, mockUserRepo)
	chats, err := service.GetChats(context.Background(), userId)

	assert.NoError(t, err)
	assert.Len(t, chats, 1)
	assert.Equal(t, chatId, chats[0].ID)
	assert.Equal(t, "TestChat", chats[0].Name)
	assert.Equal(t, "Hello", chats[0].LastMessage.Text)
}

func TestGetChats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	userId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChats(gomock.Any(), userId).
		Return(nil, errors.New("repo error"))

	service := NewChatsService(mockChatsRepo, mockUserRepo)
	_, err := service.GetChats(context.Background(), userId)

	assert.Error(t, err)
}

func TestGetInformationAboutChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChat(gomock.Any(), userId, chatId).
		Return(&modelsChats.Chat{ID: chatId, Name: "Chat1", Type: modelsChats.ChatTypeDialog}, nil)

	mockChatsRepo.EXPECT().
		GetMessagesOfChat(gomock.Any(), chatId, gomock.Any(), gomock.Any()).
		Return([]modelsMessage.Message{{
			UserID:    userId,
			Text:      "Hi",
			CreatedAt: time.Now(),
		}}, nil)

	mockChatsRepo.EXPECT().
		GetUsersOfChat(gomock.Any(), chatId).
		Return([]modelsChats.UserInfo{{
			UserID: userId,
			Role:   modelsChats.RoleAdmin,
		}}, nil)

	mockChatsRepo.EXPECT().
		GetUserInfo(gomock.Any(), userId, chatId).
		Return(&modelsChats.UserInfo{UserID: userId, Role: modelsChats.RoleAdmin}, nil)

	service := NewChatsService(mockChatsRepo, mockUserRepo)
	info, err := service.GetInformationAboutChat(context.Background(), userId, chatId)

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
	ctrl := gomock.NewController(t)

	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChat(gomock.Any(), userId, chatId).
		Return(nil, errors.New("not found"))

	service := NewChatsService(mockChatsRepo, mockUserRepo)
	info, err := service.GetInformationAboutChat(context.Background(), userId, chatId)

	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestCreateChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	userIds := []uuid.UUID{uuid.New()}

	mockUserRepo.EXPECT().
		GetUsersNames(gomock.Any(), userIds).
		Return([]string{"TestUser"}, nil)

	mockChatsRepo.EXPECT().
		CreateChat(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	service := NewChatsService(mockChatsRepo, mockUserRepo)
	chatDTO := dto.ChatCreateInformationDTO{
		Name: "NewChat",
		Type: modelsChats.ChatTypeDialog,
		Members: []dto.AddChatMemberDTO{
			{UserId: userIds[0], Role: modelsChats.RoleAdmin},
		},
	}
	id, err := service.CreateChat(context.Background(), chatDTO)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestCreateChat_Error(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	userIds := []uuid.UUID{uuid.New()}

	mockUserRepo.EXPECT().
		GetUsersNames(gomock.Any(), userIds).
		Return([]string{"TestUser"}, nil)

	mockChatsRepo.EXPECT().
		CreateChat(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("fail"))

	service := NewChatsService(mockChatsRepo, mockUserRepo)
	chatDTO := dto.ChatCreateInformationDTO{
		Name: "FailChat",
		Type: modelsChats.ChatTypeDialog,
		Members: []dto.AddChatMemberDTO{
			{UserId: userIds[0], Role: modelsChats.RoleAdmin},
		},
	}
	id, err := service.CreateChat(context.Background(), chatDTO)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}

func TestAddUsersToChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	chatID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	users := []dto.AddChatMemberDTO{
		{UserId: userID1, Role: modelsChats.RoleAdmin},
		{UserId: userID2, Role: modelsChats.RoleMember},
	}

	expectedUsersInfo := []modelsChats.UserInfo{
		{UserID: userID1, ChatID: chatID, Role: modelsChats.RoleAdmin},
		{UserID: userID2, ChatID: chatID, Role: modelsChats.RoleMember},
	}

	mockChatsRepo.EXPECT().
		InsertUsersToChat(gomock.Any(), chatID, expectedUsersInfo).
		Return(nil)

	service := NewChatsService(mockChatsRepo, mockUserRepo)
	err := service.AddUsersToChat(context.Background(), chatID, users)

	assert.NoError(t, err)
}
