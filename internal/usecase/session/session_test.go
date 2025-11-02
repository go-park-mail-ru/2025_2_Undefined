package usecase

import (
	"errors"
	"testing"
	"time"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) AddSession(userID uuid.UUID, device string) (uuid.UUID, error) {
	args := m.Called(userID, device)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockSessionRepository) DeleteSession(sessionID uuid.UUID) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSession(sessionID uuid.UUID) (*models.Session, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetSessionsByUserID(userID uuid.UUID) ([]*models.Session, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) UpdateSession(sessionID uuid.UUID) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func TestSessionUsecase_GetSession_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	sessionID := uuid.New()
	userID := uuid.New()
	device := "Chrome on Windows"
	now := time.Now()

	session := &models.Session{
		ID:         sessionID,
		UserID:     userID,
		Device:     device,
		Created_at: now,
		Last_seen:  now,
	}

	mockRepo.On("GetSession", sessionID).Return(session, nil)

	result, err := uc.GetSession(sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, sessionID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, device, result.Device)
	assert.Equal(t, now, result.Created_at)
	assert.Equal(t, now, result.Last_seen)
	mockRepo.AssertExpectations(t)
}

func TestSessionUsecase_GetSession_NilSessionID(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	result, err := uc.GetSession(uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "session required")
}

func TestSessionUsecase_GetSession_RepositoryError(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	sessionID := uuid.New()
	mockRepo.On("GetSession", sessionID).Return(nil, errors.New("session not found"))

	result, err := uc.GetSession(sessionID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "session not found")
	mockRepo.AssertExpectations(t)
}

func TestSessionUsecase_GetSessionsByUserID_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	userID := uuid.New()
	sessionID1 := uuid.New()
	sessionID2 := uuid.New()
	device1 := "Chrome on Windows"
	device2 := "Safari on iPhone"
	now := time.Now()

	sessions := []*models.Session{
		{
			ID:         sessionID1,
			UserID:     userID,
			Device:     device1,
			Created_at: now,
			Last_seen:  now,
		},
		{
			ID:         sessionID2,
			UserID:     userID,
			Device:     device2,
			Created_at: now.Add(-time.Hour),
			Last_seen:  now.Add(-time.Minute),
		},
	}

	mockRepo.On("GetSessionsByUserID", userID).Return(sessions, nil)

	result, err := uc.GetSessionsByUserID(userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, sessionID1, result[0].ID)
	assert.Equal(t, device1, result[0].Device)
	assert.Equal(t, sessionID2, result[1].ID)
	assert.Equal(t, device2, result[1].Device)
	mockRepo.AssertExpectations(t)
}

func TestSessionUsecase_GetSessionsByUserID_NilUserID(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	result, err := uc.GetSessionsByUserID(uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestSessionUsecase_GetSessionsByUserID_RepositoryError(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetSessionsByUserID", userID).Return(nil, errors.New("redis error"))

	result, err := uc.GetSessionsByUserID(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "redis error")
	mockRepo.AssertExpectations(t)
}

func TestSessionUsecase_GetSessionsByUserID_EmptyResult(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetSessionsByUserID", userID).Return([]*models.Session{}, nil)

	result, err := uc.GetSessionsByUserID(userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestSessionUsecase_UpdateSession_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	sessionID := uuid.New()
	mockRepo.On("UpdateSession", sessionID).Return(nil)

	err := uc.UpdateSession(sessionID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSessionUsecase_UpdateSession_NilSessionID(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	err := uc.UpdateSession(uuid.Nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session ID is required")
}

func TestSessionUsecase_UpdateSession_RepositoryError(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := New(mockRepo)

	sessionID := uuid.New()
	mockRepo.On("UpdateSession", sessionID).Return(errors.New("session not found"))

	err := uc.UpdateSession(sessionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
	mockRepo.AssertExpectations(t)
}
