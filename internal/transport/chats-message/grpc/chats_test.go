package chats

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockChatsUsecase struct {
	mock.Mock
}

func (m *MockChatsUsecase) GetChats(ctx context.Context, userId uuid.UUID) ([]dtoChats.ChatViewInformationDTO, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]dtoChats.ChatViewInformationDTO), args.Error(1)
}

func (m *MockChatsUsecase) CreateChat(ctx context.Context, chatDTO dtoChats.ChatCreateInformationDTO) (uuid.UUID, error) {
	args := m.Called(ctx, chatDTO)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockChatsUsecase) GetInformationAboutChat(ctx context.Context, userId, chatId uuid.UUID, offset, limit int) (*dtoChats.ChatDetailedInformationDTO, error) {
	args := m.Called(ctx, userId, chatId, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtoChats.ChatDetailedInformationDTO), args.Error(1)
}

func (m *MockChatsUsecase) GetUsersDialog(ctx context.Context, user1ID, user2ID uuid.UUID) (*dtoUtils.IdDTO, error) {
	args := m.Called(ctx, user1ID, user2ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtoUtils.IdDTO), args.Error(1)
}

func (m *MockChatsUsecase) AddUsersToChat(ctx context.Context, chatID, userID uuid.UUID, users []dtoChats.AddChatMemberDTO) error {
	args := m.Called(ctx, chatID, userID, users)
	return args.Error(0)
}

func (m *MockChatsUsecase) DeleteChat(ctx context.Context, userId, chatId uuid.UUID) error {
	args := m.Called(ctx, userId, chatId)
	return args.Error(0)
}

func (m *MockChatsUsecase) UpdateChat(ctx context.Context, userId, chatId uuid.UUID, name, description string) error {
	args := m.Called(ctx, userId, chatId, name, description)
	return args.Error(0)
}

func (m *MockChatsUsecase) GetChatAvatars(ctx context.Context, userId uuid.UUID, chatIDs []uuid.UUID) (map[string]*string, error) {
	args := m.Called(ctx, userId, chatIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*string), args.Error(1)
}

func (m *MockChatsUsecase) UploadChatAvatar(ctx context.Context, userID, chatID uuid.UUID, fileData minio.FileData) (string, error) {
	args := m.Called(ctx, userID, chatID, fileData)
	return args.String(0), args.Error(1)
}

func (m *MockChatsUsecase) SearchChats(ctx context.Context, userID uuid.UUID, name string) ([]dtoChats.ChatViewInformationDTO, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).([]dtoChats.ChatViewInformationDTO), args.Error(1)
}

type MockMessageUsecase struct {
	mock.Mock
}

func (m *MockMessageUsecase) AddMessage(ctx context.Context, message dtoMessage.CreateMessageDTO, userID uuid.UUID) error {
	args := m.Called(ctx, message, userID)
	return args.Error(0)
}

func (m *MockMessageUsecase) EditMessage(ctx context.Context, message dtoMessage.EditMessageDTO, userID uuid.UUID) error {
	args := m.Called(ctx, message, userID)
	return args.Error(0)
}

func (m *MockMessageUsecase) DeleteMessage(ctx context.Context, message dtoMessage.DeleteMessageDTO, userID uuid.UUID) error {
	args := m.Called(ctx, message, userID)
	return args.Error(0)
}

func (m *MockMessageUsecase) SubscribeConnectionToChats(ctx context.Context, connectionID uuid.UUID, userID uuid.UUID, chatsDTO []dtoChats.ChatViewInformationDTO) <-chan dtoMessage.WebSocketMessageDTO {
	args := m.Called(ctx, connectionID, userID, chatsDTO)
	return args.Get(0).(<-chan dtoMessage.WebSocketMessageDTO)
}

func (m *MockMessageUsecase) SubscribeUsersOnChat(ctx context.Context, chatID uuid.UUID, members []dtoChats.AddChatMemberDTO) error {
	args := m.Called(ctx, chatID, members)
	return args.Error(0)
}

func (m *MockMessageUsecase) GetMessagesBySearch(ctx context.Context, userID uuid.UUID, chatID uuid.UUID, text string) ([]dtoMessage.MessageDTO, error) {
	args := m.Called(ctx, userID, chatID, text)
	return args.Get(0).([]dtoMessage.MessageDTO), args.Error(1)
}

func (m *MockMessageUsecase) GetChatMessages(ctx context.Context, userID uuid.UUID, chatID uuid.UUID, offset, limit int) ([]dtoMessage.MessageDTO, error) {
	args := m.Called(ctx, userID, chatID, offset, limit)
	return args.Get(0).([]dtoMessage.MessageDTO), args.Error(1)
}

func (m *MockMessageUsecase) AddMessageJoinUsers(ctx context.Context, chatID uuid.UUID, users []dtoChats.AddChatMemberDTO) error {
	return nil
}

func (m *MockMessageUsecase) UploadAttachment(ctx context.Context, userID, chatID uuid.UUID, contentType string, fileData []byte, filename string, duration *int) (*dtoMessage.AttachmentDTO, error) {
	args := m.Called(ctx, userID, chatID, contentType, fileData, filename, duration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtoMessage.AttachmentDTO), args.Error(1)
}

func setupContext() context.Context {
	ctx := context.Background()
	_ = domains.GetLogger(ctx)
	return ctx
}

func TestGetChats_Success(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	ctx := setupContext()

	expectedChats := []dtoChats.ChatViewInformationDTO{}
	mockChatsUC.On("GetChats", ctx, userID).Return(expectedChats, nil)

	req := &gen.GetChatsReq{UserId: userID.String()}
	resp, err := handler.GetChats(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockChatsUC.AssertExpectations(t)
}

func TestGetChats_InvalidUserID(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	ctx := setupContext()
	req := &gen.GetChatsReq{UserId: "invalid-uuid"}
	resp, err := handler.GetChats(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetChats_UsecaseError(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	ctx := setupContext()

	mockChatsUC.On("GetChats", ctx, userID).Return([]dtoChats.ChatViewInformationDTO{}, errors.New("database error"))

	req := &gen.GetChatsReq{UserId: userID.String()}
	resp, err := handler.GetChats(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetChat_Success(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	chatID := uuid.New()
	ctx := setupContext()

	expectedChat := &dtoChats.ChatDetailedInformationDTO{ID: chatID}
	mockChatsUC.On("GetInformationAboutChat", ctx, userID, chatID, 0, 20).Return(expectedChat, nil)

	req := &gen.GetChatReq{UserId: userID.String(), ChatId: chatID.String()}
	resp, err := handler.GetChat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockChatsUC.AssertExpectations(t)
}

func TestGetChat_InvalidChatID(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	ctx := setupContext()

	req := &gen.GetChatReq{UserId: userID.String(), ChatId: "invalid-uuid"}
	resp, err := handler.GetChat(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestDeleteChat_Success(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	chatID := uuid.New()
	ctx := setupContext()

	mockChatsUC.On("DeleteChat", ctx, userID, chatID).Return(nil)

	req := &gen.GetChatReq{UserId: userID.String(), ChatId: chatID.String()}
	resp, err := handler.DeleteChat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockChatsUC.AssertExpectations(t)
}

func TestDeleteChat_UsecaseError(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	chatID := uuid.New()
	ctx := setupContext()

	mockChatsUC.On("DeleteChat", ctx, userID, chatID).Return(errors.New("permission denied"))

	req := &gen.GetChatReq{UserId: userID.String(), ChatId: chatID.String()}
	resp, err := handler.DeleteChat(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetUsersDialog_Success(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	user1ID := uuid.New()
	user2ID := uuid.New()
	dialogID := uuid.New()
	ctx := setupContext()

	expectedDialog := &dtoUtils.IdDTO{ID: dialogID}
	mockChatsUC.On("GetUsersDialog", ctx, user1ID, user2ID).Return(expectedDialog, nil)

	req := &gen.GetUsersDialogReq{User1Id: user1ID.String(), User2Id: user2ID.String()}
	resp, err := handler.GetUsersDialog(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, dialogID.String(), resp.Id)
	mockChatsUC.AssertExpectations(t)
}

func TestGetUsersDialog_InvalidUser1ID(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	user2ID := uuid.New()
	ctx := setupContext()

	req := &gen.GetUsersDialogReq{User1Id: "invalid-uuid", User2Id: user2ID.String()}
	resp, err := handler.GetUsersDialog(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSearchChats_Success(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	ctx := setupContext()
	nameQuery := "test"

	expectedChats := []dtoChats.ChatViewInformationDTO{}
	mockChatsUC.On("SearchChats", ctx, userID, nameQuery).Return(expectedChats, nil)

	req := &gen.SearchChatsReq{UserId: userID.String(), Name: nameQuery}
	resp, err := handler.SearchChats(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockChatsUC.AssertExpectations(t)
}

func TestSearchChats_UsecaseError(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	ctx := setupContext()
	nameQuery := "test"

	mockChatsUC.On("SearchChats", ctx, userID, nameQuery).Return([]dtoChats.ChatViewInformationDTO{}, errors.New("search error"))

	req := &gen.SearchChatsReq{UserId: userID.String(), Name: nameQuery}
	resp, err := handler.SearchChats(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestGetChatAvatars_Success(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	chatID1 := uuid.New()
	chatID2 := uuid.New()
	ctx := setupContext()

	avatarURL1 := "http://example.com/avatar1.jpg"
	avatarURL2 := "http://example.com/avatar2.jpg"
	expectedAvatars := map[string]*string{
		chatID1.String(): &avatarURL1,
		chatID2.String(): &avatarURL2,
	}

	mockChatsUC.On("GetChatAvatars", ctx, userID, []uuid.UUID{chatID1, chatID2}).Return(expectedAvatars, nil)

	req := &gen.GetChatAvatarsReq{
		UserId:  userID.String(),
		ChatIds: []string{chatID1.String(), chatID2.String()},
	}
	resp, err := handler.GetChatAvatars(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Avatars))
	mockChatsUC.AssertExpectations(t)
}

func TestGetChatAvatars_EmptyChatIDs(t *testing.T) {
	mockChatsUC := new(MockChatsUsecase)
	mockMessageUC := new(MockMessageUsecase)
	handler := NewChatsGRPCHandler(mockChatsUC, mockMessageUC)

	userID := uuid.New()
	ctx := setupContext()

	req := &gen.GetChatAvatarsReq{
		UserId:  userID.String(),
		ChatIds: []string{},
	}
	resp, err := handler.GetChatAvatars(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.Avatars))
}
