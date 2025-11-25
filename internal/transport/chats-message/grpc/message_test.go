package chats

import (
	"context"
	"errors"
	"testing"

	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestHandleSendMessage_NewMessage_Success(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	chatID := uuid.New()
	ctx := setupContext()

	mockMessageUC.On("AddMessage", ctx, mock.AnythingOfType("dto.CreateMessageDTO"), userID).Return(nil)

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event: &gen.MessageEventReq_NewChatMessage{
			NewChatMessage: &gen.CreateMessage{
				ChatId: chatID.String(),
				Text:   "Hello, World!",
			},
		},
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockMessageUC.AssertExpectations(t)
}

func TestHandleSendMessage_EditMessage_Success(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	messageID := uuid.New()
	ctx := setupContext()

	mockMessageUC.On("EditMessage", ctx, mock.AnythingOfType("dto.EditMessageDTO"), userID).Return(nil)

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event: &gen.MessageEventReq_EditChatMessage{
			EditChatMessage: &gen.EditMessage{
				MessageId: messageID.String(),
				Text:      "Updated text",
			},
		},
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockMessageUC.AssertExpectations(t)
}

func TestHandleSendMessage_DeleteMessage_Success(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	messageID := uuid.New()
	ctx := setupContext()

	mockMessageUC.On("DeleteMessage", ctx, mock.AnythingOfType("dto.DeleteMessageDTO"), userID).Return(nil)

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event: &gen.MessageEventReq_DeleteChatMessage{
			DeleteChatMessage: &gen.DeleteMessage{
				MessageId: messageID.String(),
			},
		},
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockMessageUC.AssertExpectations(t)
}

func TestHandleSendMessage_InvalidUserID(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	ctx := setupContext()
	chatID := uuid.New()

	req := &gen.MessageEventReq{
		UserId: "invalid-uuid",
		Event: &gen.MessageEventReq_NewChatMessage{
			NewChatMessage: &gen.CreateMessage{
				ChatId: chatID.String(),
				Text:   "Hello",
			},
		},
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestHandleSendMessage_ValidationError(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	ctx := setupContext()
	userID := uuid.New()

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event:  nil,
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestHandleSendMessage_AddMessageError(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	chatID := uuid.New()
	ctx := setupContext()

	mockMessageUC.On("AddMessage", ctx, mock.AnythingOfType("dto.CreateMessageDTO"), userID).
		Return(errors.New("database error"))

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event: &gen.MessageEventReq_NewChatMessage{
			NewChatMessage: &gen.CreateMessage{
				ChatId: chatID.String(),
				Text:   "Hello",
			},
		},
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSearchMessages_Success(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	chatID := uuid.New()
	ctx := setupContext()
	textQuery := "test"

	expectedMessages := []dtoMessage.MessageDTO{}
	mockMessageUC.On("GetMessagesBySearch", ctx, userID, chatID, textQuery).Return(expectedMessages, nil)

	req := &gen.SearchMessagesReq{
		UserId: userID.String(),
		ChatId: chatID.String(),
		Text:   textQuery,
	}

	resp, err := handler.SearchMessages(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Messages)
	mockMessageUC.AssertExpectations(t)
}

func TestSearchMessages_InvalidUserID(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	chatID := uuid.New()
	ctx := setupContext()

	req := &gen.SearchMessagesReq{
		UserId: "invalid-uuid",
		ChatId: chatID.String(),
		Text:   "test",
	}

	resp, err := handler.SearchMessages(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSearchMessages_InvalidChatID(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	ctx := setupContext()

	req := &gen.SearchMessagesReq{
		UserId: userID.String(),
		ChatId: "invalid-uuid",
		Text:   "test",
	}

	resp, err := handler.SearchMessages(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSearchMessages_UsecaseError(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	chatID := uuid.New()
	ctx := setupContext()
	textQuery := "test"

	mockMessageUC.On("GetMessagesBySearch", ctx, userID, chatID, textQuery).
		Return([]dtoMessage.MessageDTO{}, errors.New("search error"))

	req := &gen.SearchMessagesReq{
		UserId: userID.String(),
		ChatId: chatID.String(),
		Text:   textQuery,
	}

	resp, err := handler.SearchMessages(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestCreateMessage(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	ctx := setupContext()

	message := dtoMessage.CreateMessageDTO{
		Text:   "Test message",
		ChatId: uuid.New(),
	}

	mockMessageUC.On("AddMessage", ctx, message, userID).Return(nil)

	err := handler.createMessage(ctx, userID, message)

	assert.NoError(t, err)
	mockMessageUC.AssertExpectations(t)
}

func TestEditMessage(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	ctx := setupContext()

	message := dtoMessage.EditMessageDTO{
		ID:   uuid.New(),
		Text: "Updated message",
	}

	mockMessageUC.On("EditMessage", ctx, message, userID).Return(nil)

	err := handler.editMessage(ctx, userID, message)

	assert.NoError(t, err)
	mockMessageUC.AssertExpectations(t)
}

func TestDeleteMessage(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	ctx := setupContext()

	message := dtoMessage.DeleteMessageDTO{
		ID: uuid.New(),
	}

	mockMessageUC.On("DeleteMessage", ctx, message, userID).Return(nil)

	err := handler.deleteMessage(ctx, userID, message)

	assert.NoError(t, err)
	mockMessageUC.AssertExpectations(t)
}

func TestHandleSendMessage_EmptyEvent(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	ctx := setupContext()

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event:  nil,
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestNewMessageGRPCHandler(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)

	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	assert.NotNil(t, handler)
	assert.Equal(t, mockMessageUC, handler.messageUsecase)
	assert.Equal(t, mockChatsUC, handler.chatsUsecase)
}

func TestHandleSendMessage_EditMessageError(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	messageID := uuid.New()
	ctx := setupContext()

	mockMessageUC.On("EditMessage", ctx, mock.AnythingOfType("dto.EditMessageDTO"), userID).
		Return(errors.New("edit error"))

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event: &gen.MessageEventReq_EditChatMessage{
			EditChatMessage: &gen.EditMessage{
				MessageId: messageID.String(),
				Text:      "Updated text",
			},
		},
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestHandleSendMessage_DeleteMessageError(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	messageID := uuid.New()
	ctx := setupContext()

	mockMessageUC.On("DeleteMessage", ctx, mock.AnythingOfType("dto.DeleteMessageDTO"), userID).
		Return(errors.New("delete error"))

	req := &gen.MessageEventReq{
		UserId: userID.String(),
		Event: &gen.MessageEventReq_DeleteChatMessage{
			DeleteChatMessage: &gen.DeleteMessage{
				MessageId: messageID.String(),
			},
		},
	}

	resp, err := handler.HandleSendMessage(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

type MockStreamServer struct {
	mock.Mock
	ctx context.Context
}

func (m *MockStreamServer) Send(msg *gen.MessageEventRes) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockStreamServer) Context() context.Context {
	return m.ctx
}

func (m *MockStreamServer) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockStreamServer) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockStreamServer) SetHeader(md metadata.MD) error {
	return nil
}

func (m *MockStreamServer) SendHeader(md metadata.MD) error {
	return nil
}

func (m *MockStreamServer) SetTrailer(md metadata.MD) {
}

func TestStreamMessagesForUser_InvalidUserID(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	ctx := setupContext()
	stream := &MockStreamServer{ctx: ctx}

	req := &gen.StreamMessagesForUserReq{
		UserId: "invalid-uuid",
	}

	err := handler.StreamMessagesForUser(req, stream)

	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestStreamMessagesForUser_GetChatsError(t *testing.T) {
	mockMessageUC := new(MockMessageUsecase)
	mockChatsUC := new(MockChatsUsecase)
	handler := NewMessageGRPCHandler(mockMessageUC, mockChatsUC)

	userID := uuid.New()
	ctx := setupContext()
	stream := &MockStreamServer{ctx: ctx}

	mockChatsUC.On("GetChats", ctx, userID).Return([]dtoChats.ChatViewInformationDTO{}, errors.New("database error"))

	req := &gen.StreamMessagesForUserReq{
		UserId: userID.String(),
	}

	err := handler.StreamMessagesForUser(req, stream)

	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	mockChatsUC.AssertExpectations(t)
}
