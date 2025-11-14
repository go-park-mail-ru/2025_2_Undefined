package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	SessionDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	userID := uuid.New()
	testBio := "Test bio"
	user := &UserDTO.User{
		ID:          userID,
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    "test_user",
		Bio:         &testBio,
		AccountType: "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockSessionUtils.EXPECT().GetUserIDFromSession(gomock.Any()).Return(userID, nil)
	mockUsecase.EXPECT().GetUserById(gomock.Any(), userID).Return(user, nil)

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	recorder := httptest.NewRecorder()

	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response UserDTO.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, userID, response.ID)
	assert.Equal(t, user.Name, response.Name)
}

func TestUserHandler_GetCurrentUser_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	mockSessionUtils.EXPECT().GetUserIDFromSession(gomock.Any()).Return(uuid.Nil, errors.New("unauthorized"))

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	recorder := httptest.NewRecorder()

	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_GetCurrentUser_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	userID := uuid.New()

	mockSessionUtils.EXPECT().GetUserIDFromSession(gomock.Any()).Return(userID, nil)
	mockUsecase.EXPECT().GetUserById(gomock.Any(), userID).Return(nil, errors.New("user not found"))

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	recorder := httptest.NewRecorder()

	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_GetSessionsByUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

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

	mockSessionUtils.EXPECT().GetUserIDFromSession(gomock.Any()).Return(userID, nil)
	mockSessionUsecase.EXPECT().GetSessionsByUserID(userID).Return(sessions, nil)

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
}

func TestUserHandler_GetSessionsByUser_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	mockSessionUtils.EXPECT().GetUserIDFromSession(gomock.Any()).Return(uuid.Nil, errors.New("unauthorized"))

	request := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	recorder := httptest.NewRecorder()

	handler.GetSessionsByUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_GetSessionsByUser_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	userID := uuid.New()

	mockSessionUtils.EXPECT().GetUserIDFromSession(gomock.Any()).Return(userID, nil)
	mockSessionUsecase.EXPECT().GetSessionsByUserID(userID).Return(nil, errors.New("session error"))

	request := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	recorder := httptest.NewRecorder()

	handler.GetSessionsByUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_GetUserByPhone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

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

	mockUsecase.EXPECT().GetUserByPhone(gomock.Any(), phone).Return(user, nil)

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
}

func TestUserHandler_GetUserByPhone_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUserHandler_GetUserByPhone_EmptyPhone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	phone := "+79998887766"
	req := UserDTO.GetUserByPhone{
		PhoneNumber: phone,
	}

	mockUsecase.EXPECT().GetUserByPhone(gomock.Any(), phone).Return(nil, errs.ErrUserNotFound)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestUserHandler_GetUserByPhone_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	phone := "+79998887766"
	req := UserDTO.GetUserByPhone{
		PhoneNumber: phone,
	}

	mockUsecase.EXPECT().GetUserByPhone(gomock.Any(), phone).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestUserHandler_GetUserByUsername_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

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

	mockUsecase.EXPECT().GetUserByUsername(gomock.Any(), username).Return(user, nil)

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
}

func TestUserHandler_GetUserByUsername_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUserHandler_GetUserByUsername_EmptyUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	username := "nonexistent_user"
	req := UserDTO.GetUserByUsername{
		Username: username,
	}

	mockUsecase.EXPECT().GetUserByUsername(gomock.Any(), username).Return(nil, errs.ErrUserNotFound)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestUserHandler_GetUserByUsername_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUserUsecase(ctrl)
	mockSessionUsecase := mocks.NewMockSessionUsecase(ctrl)
	mockSessionUtils := mocks.NewMockSessionUtils(ctrl)

	handler := New(mockUsecase, mockSessionUsecase, mockSessionUtils)

	username := "test_user"
	req := UserDTO.GetUserByUsername{
		Username: username,
	}

	mockUsecase.EXPECT().GetUserByUsername(gomock.Any(), username).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
