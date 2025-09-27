package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/dto"
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

func (m *mockChatsService) GetChats(userId uuid.UUID) ([]dto.ChatViewInformationDTO, error) {
	if m.getChatError != nil {
		return nil, m.getChatError
	}
	return m.chats, nil
}

func (m *mockChatsService) CreateChat(chatDTO dto.ChatCreateInformationDTO) (uuid.UUID, error) {
	if m.createChatError != nil {
		return uuid.Nil, m.createChatError
	}
	return uuid.New(), nil
}

func (m *mockChatsService) GetInformationAboutChat(userId, chatId uuid.UUID) (*dto.ChatDetailedInformationDTO, error) {
	if m.chatDetailedError != nil {
		return nil, m.chatDetailedError
	}
	return m.chatDetailed, nil
}

func TestPostChats_Success(t *testing.T) {
	mockService := &mockChatsService{}
	handler := NewChatsHandler(mockService)

	chatDTO := dto.ChatCreateInformationDTO{
		Name: "Test Chat",
		Type: 1,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: 1},
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
	handler := NewChatsHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.PostChats(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetInformationAboutChat_BadUUID(t *testing.T) {
	mockService := &mockChatsService{}
	handler := NewChatsHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/chats/invalid-uuid", nil)
	w := httptest.NewRecorder()

	handler.GetInformationAboutChat(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetInformationAboutChat_NoCookie(t *testing.T) {
	mockService := &mockChatsService{}
	handler := NewChatsHandler(mockService)

	testUUID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/chats/"+testUUID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetInformationAboutChat(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetChats_NoCookie(t *testing.T) {
	mockService := &mockChatsService{}
	handler := NewChatsHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	w := httptest.NewRecorder()

	handler.GetChats(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNewChatsHandler(t *testing.T) {
	mockService := &mockChatsService{}

	handler := NewChatsHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.chatService)
}

func TestPostChats_ServiceError(t *testing.T) {
	mockService := &mockChatsService{
		createChatError: errors.New("service error"),
	}
	handler := NewChatsHandler(mockService)

	chatDTO := dto.ChatCreateInformationDTO{
		Name: "Test Chat",
		Type: 1,
		Members: []dto.UserInfoChatDTO{
			{UserId: uuid.New(), Role: 1},
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
