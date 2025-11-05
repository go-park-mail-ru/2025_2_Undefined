package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewChatsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	assert.NotNil(t, handler)
}

func TestGetChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	userId := uuid.New()
	expectedChats := []dto.ChatViewInformationDTO{
		{ID: uuid.New(), Name: "Test Chat"},
	}

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(userId, nil)

	mockChatUsecase.EXPECT().
		GetChats(gomock.Any(), userId).
		Return(expectedChats, nil)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	w := httptest.NewRecorder()

	handler.GetChats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []dto.ChatViewInformationDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, expectedChats[0].Name, response[0].Name)
}

func TestGetChats_SessionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(uuid.UUID{}, errors.New("session error"))

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	w := httptest.NewRecorder()

	handler.GetChats(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	newChatId := uuid.New()

	chatDTO := dto.ChatCreateInformationDTO{
		Name: "Test Chat",
		Type: models.ChatTypeDialog,
		Members: []dto.AddChatMemberDTO{
			{UserId: uuid.New(), Role: models.RoleMember},
		},
	}

	mockChatUsecase.EXPECT().
		CreateChat(gomock.Any(), chatDTO).
		Return(newChatId, nil)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	body, err := json.Marshal(chatDTO)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	respBody, err := io.ReadAll(w.Body)
	assert.NoError(t, err)

	var result dtoUtils.IdDTO
	err = json.Unmarshal(respBody, &result)
	assert.NoError(t, err)
	assert.Equal(t, newChatId, result.ID)
}

func TestPostChats_BadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetInformationAboutChat_BadUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodGet, "/chats/invalid-uuid", nil)
	w := httptest.NewRecorder()

	handler.GetInformationAboutChat(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetInformationAboutChat_SessionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	testUUID := uuid.New()

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(uuid.UUID{}, errors.New("session error"))

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodGet, "/chats/"+testUUID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetInformationAboutChat(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetChats_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	userId := uuid.New()

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(userId, nil)

	mockChatUsecase.EXPECT().
		GetChats(gomock.Any(), userId).
		Return(nil, errors.New("service error"))

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	w := httptest.NewRecorder()

	handler.GetChats(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostChats_SessionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodPost, "/chats", nil)
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostChats_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	chatDTO := dto.ChatCreateInformationDTO{
		Name: "Test Chat",
		Type: models.ChatTypeChannel,
		Members: []dto.AddChatMemberDTO{
			{UserId: uuid.New(), Role: models.RoleMember},
		},
	}

	mockChatUsecase.EXPECT().
		CreateChat(gomock.Any(), chatDTO).
		Return(uuid.UUID{}, errors.New("service error"))

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	body, err := json.Marshal(chatDTO)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddUsersToChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	chatId := uuid.New()
	userId := uuid.New()
	usersToAdd := []uuid.UUID{uuid.New(), uuid.New()}

	usersToAddDTO := []dto.AddChatMemberDTO{
		{UserId: usersToAdd[0], Role: "writer"},
		{UserId: usersToAdd[1], Role: "writer"},
	}

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(userId, nil)

	mockChatUsecase.EXPECT().
		AddUsersToChat(gomock.Any(), chatId, userId, usersToAddDTO).
		Return(nil)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	addUsersDTO := dto.AddUsersToChatDTO{
		Users: usersToAddDTO,
	}
	jsonData, _ := json.Marshal(addUsersDTO)

	req := httptest.NewRequest(http.MethodPatch, "/chats/"+chatId.String()+"/members", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddUsersToChat(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddUsersToChat_BadUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodPatch, "/chats/invalid-uuid/members", nil)
	w := httptest.NewRecorder()

	handler.AddUsersToChat(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddUsersToChat_BadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	chatId := uuid.New()
	userId := uuid.New()

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(userId, nil)

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	req := httptest.NewRequest(http.MethodPatch, "/chats/"+chatId.String()+"/members", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddUsersToChat(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddUsersToChat_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	chatId := uuid.New()
	userId := uuid.New()
	usersToAdd := []uuid.UUID{uuid.New()}

	usersToAddDTO := []dto.AddChatMemberDTO{
		{UserId: usersToAdd[0], Role: "writer"},
	}

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(userId, nil)

	mockChatUsecase.EXPECT().
		AddUsersToChat(gomock.Any(), chatId, userId, usersToAddDTO).
		Return(errors.New("service error"))

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	addUsersDTO := dto.AddUsersToChatDTO{
		Users: usersToAddDTO,
	}
	jsonData, _ := json.Marshal(addUsersDTO)

	req := httptest.NewRequest(http.MethodPatch, "/chats/"+chatId.String()+"/members", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddUsersToChat(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddUsersToChat_DuplicateUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	chatId := uuid.New()
	userId := uuid.New()
	duplicateUserId := uuid.New()

	mockSessionUtils.EXPECT().
		GetUserIDFromSession(gomock.Any()).
		Return(userId, nil)

	// Один и тот же пользователь дважды
	usersToAddDTO := []dto.AddChatMemberDTO{
		{UserId: duplicateUserId, Role: "writer"},
		{UserId: duplicateUserId, Role: "admin"}, // Дубликат
	}

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	addUsersDTO := dto.AddUsersToChatDTO{
		Users: usersToAddDTO,
	}
	jsonData, _ := json.Marshal(addUsersDTO)

	req := httptest.NewRequest(http.MethodPatch, "/chats/"+chatId.String()+"/members", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddUsersToChat(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostChats_DuplicateMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockChatUsecase := mocks.NewMockChatsUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	duplicateUserId := uuid.New()

	chatDTO := dto.ChatCreateInformationDTO{
		Name: "Test Chat",
		Type: models.ChatTypeGroup,
		Members: []dto.AddChatMemberDTO{
			{UserId: duplicateUserId, Role: "admin"},
			{UserId: duplicateUserId, Role: "writer"}, // Дубликат
		},
	}

	handler := NewChatsHandler(mockChatUsecase, mockSessionUtils)

	body, err := json.Marshal(chatDTO)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
