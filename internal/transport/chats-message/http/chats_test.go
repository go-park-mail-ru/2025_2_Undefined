package chats

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/http/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGRPCGetChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	expectedResponse := &gen.GetChatsRes{
		Chats: []*gen.Chat{
			{
				Id:   chatID.String(),
				Name: "Test Chat",
				Type: "private",
			},
		},
	}

	mockClient.EXPECT().
		GetChats(gomock.Any(), &gen.GetChatsReq{UserId: userID.String()}).
		Return(expectedResponse, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetChats(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCGetChats_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	request := httptest.NewRequest(http.MethodGet, "/chats", nil)
	recorder := httptest.NewRecorder()
	handler.GetChats(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestGRPCPostChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	participantID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatCreateInformationDTO{
		Name: "New Chat",
		Type: "private",
		Members: []dtoChats.AddChatMemberDTO{
			{
				UserId: participantID,
				Role:   "writer",
			},
		},
	}

	mockClient.EXPECT().
		CreateChat(gomock.Any(), gomock.Any()).
		Return(&gen.IdRes{Id: chatID.String()}, nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.PostChats(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response dtoUtils.IdDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, chatID, response.ID)
}

func TestGRPCPostChats_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.PostChats(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGRPCGetInformationAboutChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	expectedResponse := &gen.ChatDetailedInformation{
		Id:   chatID.String(),
		Name: "Test Chat",
		Type: "private",
	}

	mockClient.EXPECT().
		GetChat(gomock.Any(), &gen.GetChatReq{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).
		Return(expectedResponse, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCGetInformationAboutChat_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodGet, "/chats/invalid-uuid", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGRPCGetInformationAboutChat_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	chatID := uuid.New()

	request := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String(), nil)
	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestGRPCGetUsersDialog_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	otherUserID := uuid.New()
	dialogID := uuid.New()

	mockClient.EXPECT().
		GetUsersDialog(gomock.Any(), &gen.GetUsersDialogReq{
			User1Id: userID.String(),
			User2Id: otherUserID.String(),
		}).
		Return(&gen.IdRes{Id: dialogID.String()}, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats/dialog/"+otherUserID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetUsersDialog(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response dtoUtils.IdDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, dialogID, response.ID)
}

func TestGRPCUpdateChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatUpdateDTO{
		Name:        "Updated Chat",
		Description: "Updated description",
	}

	mockClient.EXPECT().
		UpdateChat(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String(), bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UpdateChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCUpdateChat_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatUpdateDTO{
		Name: "Updated Chat",
	}

	mockClient.EXPECT().
		UpdateChat(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.PermissionDenied, "user is not admin"))

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String(), bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UpdateChat(recorder, request)

	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestGRPCDeleteChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	mockClient.EXPECT().
		DeleteChat(gomock.Any(), &gen.GetChatReq{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).
		Return(&emptypb.Empty{}, nil)

	request := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCDeleteChat_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	mockClient.EXPECT().
		DeleteChat(gomock.Any(), &gen.GetChatReq{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).
		Return(nil, status.Error(codes.PermissionDenied, "user is not admin"))

	request := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteChat(recorder, request)

	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestGRPCAddUsersToChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()
	participantID := uuid.New()

	reqBody := dtoChats.AddUsersToChatDTO{
		Users: []dtoChats.AddChatMemberDTO{
			{
				UserId: participantID,
				Role:   "writer",
			},
		},
	}

	mockClient.EXPECT().
		AddUserToChat(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String()+"/members", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.AddUsersToChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCAddUsersToChat_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String()+"/members", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.AddUsersToChat(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGRPCAddUsersToChat_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	chatID := uuid.New()
	participantID := uuid.New()

	reqBody := dtoChats.AddUsersToChatDTO{
		Users: []dtoChats.AddChatMemberDTO{
			{
				UserId: participantID,
				Role:   "writer",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String()+"/members", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.AddUsersToChat(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
