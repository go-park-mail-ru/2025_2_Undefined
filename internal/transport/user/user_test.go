package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	SessionDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) GetUserById(ctx context.Context, id uuid.UUID) (*UserDTO.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserDTO.User), args.Error(1)
}

func (m *MockUserUsecase) GetUserByPhone(ctx context.Context, phone string) (*UserDTO.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserDTO.User), args.Error(1)
}

func (m *MockUserUsecase) GetUserByUsername(ctx context.Context, username string) (*UserDTO.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserDTO.User), args.Error(1)
}

type MockSessionUtils struct {
	mock.Mock
}

func (m *MockSessionUtils) GetUserIDFromSession(r *http.Request) (uuid.UUID, error) {
	args := m.Called(r)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockSessionUtils) GetSessionsByUserID(userID uuid.UUID) ([]*SessionDTO.Session, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*SessionDTO.Session), args.Error(1)
}

func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	user := &UserDTO.User{
		ID:          userID,
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    "test_user",
		Bio:         "Test bio",
		AccountType: "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("GetUserById", mock.Anything, userID).Return(user, nil)

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	recorder := httptest.NewRecorder()

	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response UserDTO.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, userID, response.ID)
	assert.Equal(t, user.Name, response.Name)

	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_GetCurrentUser_Unauthorized(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(uuid.Nil, errors.New("unauthorized"))

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	recorder := httptest.NewRecorder()

	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestUserHandler_GetCurrentUser_UserNotFound(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("GetUserById", mock.Anything, userID).Return(nil, errors.New("user not found"))

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	recorder := httptest.NewRecorder()

	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_GetSessionsByUser_Success(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	sessionID := uuid.New()
	sessions := []*SessionDTO.Session{
		{
			ID:         sessionID,
			UserID:     userID,
			Device:     "Chrome on Windows",
			Created_at: time.Now(),
			Last_seen:  time.Now(),
		},
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockSessionUtils.On("GetSessionsByUserID", userID).Return(sessions, nil)

	request := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	recorder := httptest.NewRecorder()

	handler.GetSessionsByUser(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []*SessionDTO.Session
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, sessionID, response[0].ID)
	assert.Equal(t, "Chrome on Windows", response[0].Device)

	mockSessionUtils.AssertExpectations(t)
}

func TestUserHandler_GetSessionsByUser_Unauthorized(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(uuid.Nil, errors.New("unauthorized"))

	request := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	recorder := httptest.NewRecorder()

	handler.GetSessionsByUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestUserHandler_GetSessionsByUser_Error(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockSessionUtils.On("GetSessionsByUserID", userID).Return(nil, errors.New("session error"))

	request := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	recorder := httptest.NewRecorder()

	handler.GetSessionsByUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestUserHandler_GetUserByPhone_Success(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	phone := "+79998887766"
	user := &UserDTO.User{
		ID:          uuid.New(),
		PhoneNumber: phone,
		Name:        "Test User",
		Username:    "test_user",
		AccountType: "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	req := UserDTO.GetUserByPhone{
		PhoneNumber: phone,
	}

	mockUsecase.On("GetUserByPhone", mock.Anything, phone).Return(user, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response UserDTO.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, phone, response.PhoneNumber)
	assert.Equal(t, user.Name, response.Name)

	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_GetUserByPhone_InvalidJSON(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUserHandler_GetUserByPhone_EmptyPhone(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	req := UserDTO.GetUserByPhone{
		PhoneNumber: "",
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUserHandler_GetUserByPhone_UserNotFound(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	phone := "+79998887766"
	req := UserDTO.GetUserByPhone{
		PhoneNumber: phone,
	}

	mockUsecase.On("GetUserByPhone", mock.Anything, phone).Return(nil, errs.ErrUserNotFound)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_GetUserByPhone_InternalError(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	phone := "+79998887766"
	req := UserDTO.GetUserByPhone{
		PhoneNumber: phone,
	}

	mockUsecase.On("GetUserByPhone", mock.Anything, phone).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_GetUserByUsername_Success(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	username := "test_user"
	user := &UserDTO.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    username,
		AccountType: "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	req := UserDTO.GetUserByUsername{
		Username: username,
	}

	mockUsecase.On("GetUserByUsername", mock.Anything, username).Return(user, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response UserDTO.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, username, response.Username)
	assert.Equal(t, user.Name, response.Name)

	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_GetUserByUsername_InvalidJSON(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUserHandler_GetUserByUsername_EmptyUsername(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	req := UserDTO.GetUserByUsername{
		Username: "",
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUserHandler_GetUserByUsername_UserNotFound(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	username := "nonexistent_user"
	req := UserDTO.GetUserByUsername{
		Username: username,
	}

	mockUsecase.On("GetUserByUsername", mock.Anything, username).Return(nil, errs.ErrUserNotFound)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockUsecase.AssertExpectations(t)
}

func TestUserHandler_GetUserByUsername_InternalError(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	username := "test_user"
	req := UserDTO.GetUserByUsername{
		Username: username,
	}

	mockUsecase.On("GetUserByUsername", mock.Anything, username).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockUsecase.AssertExpectations(t)
}
