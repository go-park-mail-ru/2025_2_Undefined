package session

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSessionUsecase struct {
	mock.Mock
}

func (m *MockSessionUsecase) GetSession(sessionID uuid.UUID) (*dto.Session, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Session), args.Error(1)
}

func (m *MockSessionUsecase) GetSessionsByUserID(userID uuid.UUID) ([]*dto.Session, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.Session), args.Error(1)
}

func TestSessionUtils_GetUserIDFromSession_Success(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	sessionID := uuid.New()
	userID := uuid.New()
	device := "Chrome on Windows"

	session := &dto.Session{
		ID:         sessionID,
		UserID:     userID,
		Device:     device,
		Created_at: time.Now(),
		Last_seen:  time.Now(),
	}

	mockUsecase.On("GetSession", sessionID).Return(session, nil)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID.String(),
	}
	request.AddCookie(cookie)

	result, err := utils.GetUserIDFromSession(request)

	assert.NoError(t, err)
	assert.Equal(t, userID, result)
	mockUsecase.AssertExpectations(t)
}

func TestSessionUtils_GetUserIDFromSession_NoCookie(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	request := httptest.NewRequest(http.MethodGet, "/", nil)

	result, err := utils.GetUserIDFromSession(request)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Equal(t, "session required", err.Error())
}

func TestSessionUtils_GetUserIDFromSession_InvalidSessionID(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: "invalid-uuid",
	}
	request.AddCookie(cookie)

	result, err := utils.GetUserIDFromSession(request)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Equal(t, "invalid session ID", err.Error())
}

func TestSessionUtils_GetUserIDFromSession_SessionNotFound(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	sessionID := uuid.New()

	mockUsecase.On("GetSession", sessionID).Return(nil, errors.New("session not found"))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID.String(),
	}
	request.AddCookie(cookie)

	result, err := utils.GetUserIDFromSession(request)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Equal(t, "invalid session", err.Error())
	mockUsecase.AssertExpectations(t)
}

func TestSessionUtils_GetSessionsByUserID_Success(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	userID := uuid.New()
	sessionID1 := uuid.New()
	sessionID2 := uuid.New()

	sessions := []*dto.Session{
		{
			ID:         sessionID1,
			UserID:     userID,
			Device:     "Chrome on Windows",
			Created_at: time.Now(),
			Last_seen:  time.Now(),
		},
		{
			ID:         sessionID2,
			UserID:     userID,
			Device:     "Safari on iPhone",
			Created_at: time.Now().Add(-time.Hour),
			Last_seen:  time.Now().Add(-time.Minute),
		},
	}

	mockUsecase.On("GetSessionsByUserID", userID).Return(sessions, nil)

	result, err := utils.GetSessionsByUserID(userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, sessionID1, result[0].ID)
	assert.Equal(t, sessionID2, result[1].ID)
	mockUsecase.AssertExpectations(t)
}

func TestSessionUtils_GetSessionsByUserID_NilUserID(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	result, err := utils.GetSessionsByUserID(uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestSessionUtils_GetSessionsByUserID_UsecaseError(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	userID := uuid.New()

	mockUsecase.On("GetSessionsByUserID", userID).Return(nil, errors.New("redis error"))

	result, err := utils.GetSessionsByUserID(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "redis error")
	mockUsecase.AssertExpectations(t)
}

func TestSessionUtils_GetSessionsByUserID_EmptyResult(t *testing.T) {
	mockUsecase := new(MockSessionUsecase)
	sessionConfig := &config.SessionConfig{
		Signature: "session_id",
	}
	utils := NewSessionUtils(mockUsecase, sessionConfig)

	userID := uuid.New()

	mockUsecase.On("GetSessionsByUserID", userID).Return([]*dto.Session{}, nil)

	result, err := utils.GetSessionsByUserID(userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockUsecase.AssertExpectations(t)
}
