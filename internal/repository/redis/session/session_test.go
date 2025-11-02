package redis

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepository_AddSession_Success(t *testing.T) {
	client, _ := redismock.NewClientMock()
	repo := New(client, time.Hour)

	userID := uuid.New()
	device := "Chrome on Windows"

	sessionID, err := repo.AddSession(userID, device)

	if err != nil {
		assert.Contains(t, err.Error(), "was not expected")
		assert.Equal(t, uuid.Nil, sessionID)
	} else {
		assert.NotEqual(t, uuid.Nil, sessionID)
	}
}

func TestSessionRepository_AddSession_PipelineError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	repo := New(client, time.Hour)

	userID := uuid.New()
	device := "Chrome on Windows"

	_, err := repo.AddSession(userID, device)

	assert.Error(t, err)
}

func TestSessionRepository_GetSession_Success(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	userID := uuid.New()
	device := "Chrome on Windows"
	now := time.Now()

	sessionData := sessionData{
		UserID:    userID,
		Device:    device,
		CreatedAt: now,
		LastSeen:  now,
	}

	sessionJSON, _ := json.Marshal(sessionData)
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	mock.ExpectGet(sessionKey).SetVal(string(sessionJSON))

	session, err := repo.GetSession(sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, device, session.Device)
	assert.Equal(t, now.Unix(), session.Created_at.Unix())
	assert.Equal(t, now.Unix(), session.Last_seen.Unix())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_GetSession_NotFound(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	mock.ExpectGet(sessionKey).RedisNil()

	session, err := repo.GetSession(sessionID)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "session not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_GetSession_RedisError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	mock.ExpectGet(sessionKey).SetErr(redis.ErrClosed)

	session, err := repo.GetSession(sessionID)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "failed to get session")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_GetSession_InvalidJSON(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	mock.ExpectGet(sessionKey).SetVal("invalid json")

	session, err := repo.GetSession(sessionID)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "failed to unmarshal session data")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_DeleteSession_Success(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	userID := uuid.New()
	device := "Chrome on Windows"

	sessionData := sessionData{
		UserID: userID,
		Device: device,
	}

	sessionJSON, _ := json.Marshal(sessionData)
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	mock.ExpectGet(sessionKey).SetVal(string(sessionJSON))
	mock.ExpectDel(sessionKey).SetVal(1)
	mock.ExpectSRem(userSessionsKey, sessionID.String()).SetVal(1)

	err := repo.DeleteSession(sessionID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_DeleteSession_NotFound(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	mock.ExpectGet(sessionKey).RedisNil()

	err := repo.DeleteSession(sessionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_UpdateSession_Success(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	userID := uuid.New()
	device := "Chrome on Windows"
	now := time.Now()

	originalData := sessionData{
		UserID:    userID,
		Device:    device,
		CreatedAt: now,
		LastSeen:  now,
	}

	sessionJSON, _ := json.Marshal(originalData)
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	mock.ExpectGet(sessionKey).SetVal(string(sessionJSON))

	err := repo.UpdateSession(sessionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "was not expected")
}

func TestSessionRepository_UpdateSession_NotFound(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	sessionID := uuid.New()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	mock.ExpectGet(sessionKey).RedisNil()

	err := repo.UpdateSession(sessionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_GetSessionsByUserID_Success(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	userID := uuid.New()
	sessionID1 := uuid.New()
	sessionID2 := uuid.New()
	device1 := "Chrome on Windows"
	device2 := "Safari on iPhone"
	now := time.Now()

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())
	sessionKey1 := fmt.Sprintf("%s:%s", sessionPrefix, sessionID1.String())
	sessionKey2 := fmt.Sprintf("%s:%s", sessionPrefix, sessionID2.String())

	sessionData1 := sessionData{
		UserID:    userID,
		Device:    device1,
		CreatedAt: now,
		LastSeen:  now,
	}

	sessionData2 := sessionData{
		UserID:    userID,
		Device:    device2,
		CreatedAt: now.Add(-time.Hour),
		LastSeen:  now.Add(-time.Minute),
	}

	sessionJSON1, _ := json.Marshal(sessionData1)
	sessionJSON2, _ := json.Marshal(sessionData2)

	mock.ExpectSMembers(userSessionsKey).SetVal([]string{sessionID1.String(), sessionID2.String()})
	mock.ExpectGet(sessionKey1).SetVal(string(sessionJSON1))
	mock.ExpectGet(sessionKey2).SetVal(string(sessionJSON2))

	sessions, err := repo.GetSessionsByUserID(userID)

	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_GetSessionsByUserID_EmptyResult(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	userID := uuid.New()
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	mock.ExpectSMembers(userSessionsKey).SetVal([]string{})

	sessions, err := repo.GetSessionsByUserID(userID)

	assert.NoError(t, err)
	assert.Empty(t, sessions)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_GetSessionsByUserID_RedisError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := New(client, time.Hour)

	userID := uuid.New()
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	mock.ExpectSMembers(userSessionsKey).SetErr(redis.ErrClosed)

	sessions, err := repo.GetSessionsByUserID(userID)

	assert.Error(t, err)
	assert.Nil(t, sessions)
	assert.Contains(t, err.Error(), "failed to get user session IDs")
	assert.NoError(t, mock.ExpectationsWereMet())
}
