package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Мок сервиса для тестирования
type mockChatsService struct {
	chats             []dto.ChatViewInformationDTO
	createChatError   error
	getChatError      error
	chatDetailedError error
	chatDetailed      *dto.ChatDetailedInformationDTO
}

// Мок утилит сессий для тестирования
type mockSessionUtils struct {
	userID uuid.UUID
	err    error
}

func (m *mockSessionUtils) GetUserIDFromSession(r *http.Request) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return m.userID, nil
}

func (m *mockChatsService) GetChats(ctx context.Context, userId uuid.UUID) ([]dto.ChatViewInformationDTO, error) {
	if m.getChatError != nil {
		return nil, m.getChatError
	}
	return m.chats, nil
}

func (m *mockChatsService) CreateChat(ctx context.Context, chatDTO dto.ChatCreateInformationDTO) (uuid.UUID, error) {
	if m.createChatError != nil {
		return uuid.Nil, m.createChatError
	}
	return uuid.New(), nil
}

func (m *mockChatsService) GetInformationAboutChat(ctx context.Context, userId, chatId uuid.UUID) (*dto.ChatDetailedInformationDTO, error) {
	if m.chatDetailedError != nil {
		return nil, m.chatDetailedError
	}
	return m.chatDetailed, nil
}

func TestPostChats_Success(t *testing.T) {
	mockService := &mockChatsService{}
	mockSessionUtils := &mockSessionUtils{userID: uuid.New()}
	handler := NewChatsHandler(mockService, mockSessionUtils)

	chatDTO := dto.ChatCreateInformationDTO{
		Name: "Test Chat",
		Type: models.ChatTypeDialog,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: models.RoleMember},
		},
	}

	body, err := json.Marshal(chatDTO)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var result dto.IdDTO
	err = json.Unmarshal(respBody, &result)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ID)
}

func TestPostChats_BadJSON(t *testing.T) {
	mockService := &mockChatsService{}
	mockSessionUtils := &mockSessionUtils{userID: uuid.New()}
	handler := NewChatsHandler(mockService, mockSessionUtils)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetInformationAboutChat_BadUUID(t *testing.T) {
	mockService := &mockChatsService{}
	mockSessionUtils := &mockSessionUtils{userID: uuid.New()}
	handler := NewChatsHandler(mockService, mockSessionUtils)

	req := httptest.NewRequest(http.MethodGet, "/chats/invalid-uuid", nil)
	w := httptest.NewRecorder()

	handler.GetInformationAboutChat(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetInformationAboutChat_NoCookie(t *testing.T) {
	mockService := &mockChatsService{}
	mockSessionUtils := &mockSessionUtils{err: errors.New("unauthorized")}
	handler := NewChatsHandler(mockService, mockSessionUtils)

	testUUID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/chats/"+testUUID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetInformationAboutChat(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetChats_NoCookie(t *testing.T) {
	mockService := &mockChatsService{}
	mockSessionUtils := &mockSessionUtils{err: errors.New("unauthorized")}
	handler := NewChatsHandler(mockService, mockSessionUtils)

	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	w := httptest.NewRecorder()

	handler.GetChats(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNewChatsHandler(t *testing.T) {
	mockService := &mockChatsService{}
	mockSessionUtils := &mockSessionUtils{userID: uuid.New()}

	handler := NewChatsHandler(mockService, mockSessionUtils)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.chatService)
	assert.Equal(t, mockSessionUtils, handler.sessionUtils)
}

func TestPostChats_ServiceError(t *testing.T) {
	mockService := &mockChatsService{
		createChatError: errors.New("service error"),
	}
	mockSessionUtils := &mockSessionUtils{userID: uuid.New()}
	handler := NewChatsHandler(mockService, mockSessionUtils)

	chatDTO := dto.ChatCreateInformationDTO{
		Name: "Test Chat",
		Type: models.ChatTypeChannel,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: models.RoleMember},
		},
	}

	body, err := json.Marshal(chatDTO)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
