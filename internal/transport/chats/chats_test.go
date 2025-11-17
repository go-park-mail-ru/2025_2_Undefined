package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()
	chatID := uuid.New()

	expectedChats := []dtoChats.ChatViewInformationDTO{
		{
			ID:   chatID,
			Name: "Test Chat",
			Type: "private",
		},
	}

	mockChatUsecase.EXPECT().
		GetChats(gomock.Any(), userID).
		Return(expectedChats, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetChats(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []dtoChats.ChatViewInformationDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, chatID, response[0].ID)
	assert.Equal(t, "Test Chat", response[0].Name)
}

func TestGetChats_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	request := httptest.NewRequest(http.MethodGet, "/chats", nil)
	recorder := httptest.NewRecorder()
	handler.GetChats(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestPostChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

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

	mockChatUsecase.EXPECT().
		CreateChat(gomock.Any(), reqBody).
		Return(chatID, nil)

	mockMessageUsecase.EXPECT().
		SubscribeUsersOnChat(gomock.Any(), chatID, gomock.Any()).
		Return(nil)

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

func TestPostChats_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.PostChats(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetInformationAboutChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()
	chatID := uuid.New()

	expectedChat := &dtoChats.ChatDetailedInformationDTO{
		ID:   chatID,
		Name: "Test Chat",
		Type: "private",
	}

	mockChatUsecase.EXPECT().
		GetInformationAboutChat(gomock.Any(), userID, chatID).
		Return(expectedChat, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response dtoChats.ChatDetailedInformationDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, chatID, response.ID)
	assert.Equal(t, "Test Chat", response.Name)
}

func TestGetInformationAboutChat_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodGet, "/chats/invalid-uuid", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUpdateChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatUpdateDTO{
		Name:        "Updated Chat",
		Description: "Updated description",
	}

	mockChatUsecase.EXPECT().
		UpdateChat(gomock.Any(), userID, chatID, reqBody.Name, reqBody.Description).
		Return(nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String(), bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UpdateChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestUpdateChat_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatUpdateDTO{
		Name: "Updated Chat",
	}

	mockChatUsecase.EXPECT().
		UpdateChat(gomock.Any(), userID, chatID, reqBody.Name, reqBody.Description).
		Return(errors.New("user is not admin"))

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String(), bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UpdateChat(recorder, request)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestDeleteChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()
	chatID := uuid.New()

	mockChatUsecase.EXPECT().
		DeleteChat(gomock.Any(), userID, chatID).
		Return(nil)

	request := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestDeleteChat_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

	userID := uuid.New()
	chatID := uuid.New()

	mockChatUsecase.EXPECT().
		DeleteChat(gomock.Any(), userID, chatID).
		Return(errors.New("user is not admin"))

	request := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteChat(recorder, request)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAddUsersToChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

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

	mockChatUsecase.EXPECT().
		AddUsersToChat(gomock.Any(), chatID, userID, gomock.Any()).
		Return(nil)

	mockMessageUsecase.EXPECT().
		SubscribeUsersOnChat(gomock.Any(), chatID, gomock.Any()).
		Return(nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String()+"/members", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.AddUsersToChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestAddUsersToChat_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockMessageUsecase := mocks.NewMockMessageUsecase(ctrl)
	handler := NewChatsHandler(mockMessageUsecase, mockChatUsecase)

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
