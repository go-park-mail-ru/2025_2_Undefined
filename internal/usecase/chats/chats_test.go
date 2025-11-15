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

func createTestHandler(ctrl *gomock.Controller) (*ChatsUsecase, *mocks.MockChatsRepository, *mocks.MockMessageRepository, *mocks.MockUserRepository, *mocks.MockFileStorage) {
	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockMessageRepo := mocks.NewMockMessageRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockStorage := mocks.NewMockFileStorage(ctrl)

	service := NewChatsUsecase(mockChatsRepo, mockUserRepo, mockMessageRepo, mockStorage)
	return service, mockChatsRepo, mockMessageRepo, mockUserRepo, mockStorage
}

func TestGetChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	service, mockChatsRepo, mockMessageRepo, _, mockStorage := createTestHandler(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChats(gomock.Any(), userId).
		Return([]modelsChats.Chat{{ID: chatId, Name: "TestChat"}}, nil)

	mockMessageRepo.EXPECT().
		GetLastMessagesOfChats(gomock.Any(), userId).
		Return([]modelsMessage.Message{{
			ChatID:    chatId,
			UserID:    &userId,
			Text:      "Hello",
			CreatedAt: time.Now(),
		}}, nil)

	mockStorage.EXPECT().
		GetOne(gomock.Any(), gomock.Any()).
		Return("", nil)

	chats, err := service.GetChats(context.Background(), userId)

	assert.NoError(t, err)
	assert.Len(t, chats, 1)
	assert.Equal(t, chatId, chats[0].ID)
	assert.Equal(t, "TestChat", chats[0].Name)
	assert.Equal(t, "Hello", chats[0].LastMessage.Text)
}

func TestGetChats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)

	service, mockChatsRepo, _, _, _ := createTestHandler(ctrl)

	userId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChats(gomock.Any(), userId).
		Return(nil, errors.New("repo error"))

	_, err := service.GetChats(context.Background(), userId)

	assert.Error(t, err)
}

func TestGetInformationAboutChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	service, mockChatsRepo, mockMessageRepo, _, mockStorage := createTestHandler(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChat(gomock.Any(), chatId).
		Return(&modelsChats.Chat{ID: chatId, Name: "Chat1", Type: modelsChats.ChatTypeDialog}, nil)

	mockMessageRepo.EXPECT().
		GetMessagesOfChat(gomock.Any(), chatId, gomock.Any(), gomock.Any()).
		Return([]modelsMessage.Message{{
			UserID:    &userId,
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

	mockStorage.EXPECT().
		GetOne(gomock.Any(), gomock.Any()).
		Return("", nil).
		AnyTimes()

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

	service, mockChatsRepo, _, _, _ := createTestHandler(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		GetChat(gomock.Any(), chatId).
		Return(nil, errors.New("not found"))

	info, err := service.GetInformationAboutChat(context.Background(), userId, chatId)

	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestCreateChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	service, mockChatsRepo, _, mockUserRepo, _ := createTestHandler(ctrl)

	userIds := []uuid.UUID{uuid.New()}

	mockUserRepo.EXPECT().
		GetUsersNames(gomock.Any(), userIds).
		Return([]string{"TestUser"}, nil)

	mockChatsRepo.EXPECT().
		CreateChat(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

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

	service, mockChatsRepo, _, mockUserRepo, _ := createTestHandler(ctrl)

	userIds := []uuid.UUID{uuid.New()}

	mockUserRepo.EXPECT().
		GetUsersNames(gomock.Any(), userIds).
		Return([]string{"TestUser"}, nil)

	mockChatsRepo.EXPECT().
		CreateChat(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("fail"))

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

	service, mockChatsRepo, mockMessageRepo, mockUserRepo, _ := createTestHandler(ctrl)

	chatID := uuid.New()
	adminUserID := uuid.New()
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
		CheckUserHasRole(gomock.Any(), adminUserID, chatID, modelsChats.RoleAdmin).
		Return(true, nil)

	mockChatsRepo.EXPECT().
		InsertUsersToChat(gomock.Any(), chatID, expectedUsersInfo).
		Return(nil)

	// В реальной логике usecase после вставки участников мы получаем информацию о чате
	// и добавляем системные сообщения для групп. Мокируем эти вызовы здесь.
	mockChatsRepo.EXPECT().
		GetChat(gomock.Any(), chatID).
		Return(&modelsChats.Chat{Type: modelsChats.ChatTypeGroup}, nil)

	usersIDs := []uuid.UUID{userID1, userID2}
	mockUserRepo.EXPECT().
		GetUsersNames(gomock.Any(), usersIDs).
		Return([]string{"User1", "User2"}, nil)

	mockMessageRepo.EXPECT().
		InsertMessage(gomock.Any(), gomock.Any()).
		Return(uuid.New(), nil).
		Times(2)

	err := service.AddUsersToChat(context.Background(), chatID, adminUserID, users)

	assert.NoError(t, err)
}
