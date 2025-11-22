package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestSessionHandler_GetSessionsByUser_Success(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	userID := uuid.New()
	sessionID1 := uuid.New()
	sessionID2 := uuid.New()

	mockAuthClient.On("GetSessionsByUserID", mock.Anything, mock.MatchedBy(func(r *gen.GetSessionsByUserIDReq) bool {
		return r.UserId == userID.String()
	})).Return(&gen.GetSessionsByUserIDRes{
		Sessions: []*gen.Session{
			{
				Id:       sessionID1.String(),
				UserId:   userID.String(),
				Device:   "Chrome on Windows",
				LastSeen: "2024-01-01T00:00:00Z",
			},
			{
				Id:       sessionID2.String(),
				UserId:   userID.String(),
				Device:   "Firefox on Linux",
				LastSeen: "2024-01-02T00:00:00Z",
			},
		},
	}, nil)

	request := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetSessionsByUser(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestSessionHandler_GetSessionsByUser_Unauthorized(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	request := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	recorder := httptest.NewRecorder()
	handler.GetSessionsByUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestSessionHandler_DeleteSession_Success(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	userID := uuid.New()
	sessionID := uuid.New()

	mockAuthClient.On("DeleteSession", mock.Anything, mock.MatchedBy(func(r *gen.DeleteSessionReq) bool {
		return r.UserId == userID.String() && r.SessionId == sessionID.String()
	})).Return(&emptypb.Empty{}, nil)

	reqBody := map[string]string{"id": sessionID.String()}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodDelete, "/session", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteSession(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_InvalidJSON(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodDelete, "/session", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteSession(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestSessionHandler_DeleteAllSessionsExceptCurrent_Success(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	userID := uuid.New()
	currentSessionID := uuid.New()

	mockAuthClient.On("DeleteAllSessionsExceptCurrent", mock.Anything, mock.MatchedBy(func(r *gen.DeleteAllSessionsExceptCurrentReq) bool {
		return r.UserId == userID.String() && r.CurrentSessionId == currentSessionID.String()
	})).Return(&emptypb.Empty{}, nil)

	request := httptest.NewRequest(http.MethodDelete, "/sessions", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)
	request.AddCookie(&http.Cookie{
		Name:  sessionConfig.Signature,
		Value: currentSessionID.String(),
	})

	recorder := httptest.NewRecorder()
	handler.DeleteAllSessionsExceptCurrent(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestSessionHandler_DeleteAllSessionsExceptCurrent_NoSession(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodDelete, "/sessions", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteAllSessionsExceptCurrent(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestSessionHandler_DeleteSession_GRPCError(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	userID := uuid.New()
	sessionID := uuid.New()

	mockAuthClient.On("DeleteSession", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.NotFound, "session not found"))

	reqBody := map[string]string{"id": sessionID.String()}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodDelete, "/session", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteSession(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockAuthClient.AssertExpectations(t)
}
