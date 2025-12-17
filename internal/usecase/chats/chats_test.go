package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
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

	service, mockChatsRepo, mockMessageRepo, _, _ := createTestHandler(ctrl)

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

	info, err := service.GetInformationAboutChat(context.Background(), userId, chatId, 0, 20)

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

	info, err := service.GetInformationAboutChat(context.Background(), userId, chatId, 0, 20)

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

	service, mockChatsRepo, _, _, _ := createTestHandler(ctrl)

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

	err := service.AddUsersToChat(context.Background(), chatID, adminUserID, users)

	assert.NoError(t, err)
}

func TestDeleteChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	service, mockChatsRepo, _, _, _ := createTestHandler(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		DeleteChat(gomock.Any(), userId, chatId).
		Return(nil)

	err := service.DeleteChat(context.Background(), userId, chatId)
	assert.NoError(t, err)
}

func TestUpdateChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	service, mockChatsRepo, _, _, _ := createTestHandler(ctrl)

	userId := uuid.New()
	chatId := uuid.New()

	mockChatsRepo.EXPECT().
		UpdateChat(gomock.Any(), userId, chatId, "NewName", "NewDesc").
		Return(nil)

	err := service.UpdateChat(context.Background(), userId, chatId, "NewName", "NewDesc")
	assert.NoError(t, err)
}

func TestGetUsersDialog_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	service, mockChatsRepo, _, _, _ := createTestHandler(ctrl)

	user1ID := uuid.New()
	user2ID := uuid.New()
	dialogID := uuid.New()

	mockChatsRepo.EXPECT().
		GetUsersDialog(gomock.Any(), user1ID, user2ID).
		Return(dialogID, nil)

	result, err := service.GetUsersDialog(context.Background(), user1ID, user2ID)

	assert.NoError(t, err)
	assert.Equal(t, dialogID, result.ID)
}

func TestSearchChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	service, mockChatsRepo, mockMessageRepo, _, _ := createTestHandler(ctrl)

	userID := uuid.New()
	chatID := uuid.New()
	query := "test"

	chats := []modelsChats.Chat{
		{ID: chatID, Name: "TestChat", Type: modelsChats.ChatTypeGroup},
	}

	mockChatsRepo.EXPECT().
		SearchChats(gomock.Any(), userID, query).
		Return(chats, nil)

	lastMessages := map[uuid.UUID]modelsMessage.Message{
		chatID: {
			ID:        uuid.New(),
			UserID:    &userID,
			Text:      "Last message",
			CreatedAt: time.Now(),
			ChatID:    chatID,
		},
	}

	mockMessageRepo.EXPECT().
		GetLastMessagesOfChatsByIDs(gomock.Any(), []uuid.UUID{chatID}).
		Return(lastMessages, nil)

	result, err := service.SearchChats(context.Background(), userID, query)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, chatID, result[0].ID)
	assert.Equal(t, "TestChat", result[0].Name)
}

func TestGetChatAvatars_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	service, mockChatsRepo, _, _, mockStorage := createTestHandler(ctrl)

	userID := uuid.New()
	chatID1 := uuid.New()
	chatID2 := uuid.New()
	chatIDs := []uuid.UUID{chatID1, chatID2}

	attachmentID1 := uuid.New()
	attachmentID2 := uuid.New()
	avatarsIDs := map[string]uuid.UUID{
		chatID1.String(): attachmentID1,
		chatID2.String(): attachmentID2,
	}

	url1 := "https://example.com/chat1.jpg"
	url2 := "https://example.com/chat2.jpg"

	mockChatsRepo.EXPECT().
		GetChatAvatars(gomock.Any(), userID, chatIDs).
		Return(avatarsIDs, nil)

	mockStorage.EXPECT().
		GetOne(gomock.Any(), &attachmentID1).
		Return(url1, nil)

	mockStorage.EXPECT().
		GetOne(gomock.Any(), &attachmentID2).
		Return(url2, nil)

	result, err := service.GetChatAvatars(context.Background(), userID, chatIDs)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, url1, *result[chatID1.String()])
	assert.Equal(t, url2, *result[chatID2.String()])
}

func TestUploadChatAvatar_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	service, mockChatsRepo, _, _, mockStorage := createTestHandler(ctrl)

	userID := uuid.New()
	chatID := uuid.New()
	fileData := minio.FileData{
		Name:        "avatar.jpg",
		Data:        []byte("fake image data"),
		ContentType: "image/jpeg",
	}
	avatarURL := "https://example.com/avatar.jpg"

	mockChatsRepo.EXPECT().
		CheckUserHasRole(gomock.Any(), userID, chatID, modelsChats.RoleAdmin).
		Return(true, nil)

	mockStorage.EXPECT().
		CreateOne(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(avatarURL, nil)

	mockChatsRepo.EXPECT().
		UpdateChatAvatar(gomock.Any(), chatID, gomock.Any(), int64(len(fileData.Data))).
		Return(nil)

	result, err := service.UploadChatAvatar(context.Background(), userID, chatID, fileData)

	assert.NoError(t, err)
	assert.Equal(t, avatarURL, result)
}
