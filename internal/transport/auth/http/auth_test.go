package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockAuthServiceClient struct {
	mock.Mock
}

func (m *MockAuthServiceClient) Register(ctx context.Context, in *gen.RegisterReq, opts ...grpc.CallOption) (*gen.RegisterRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.RegisterRes), args.Error(1)
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *gen.LoginReq, opts ...grpc.CallOption) (*gen.LoginRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.LoginRes), args.Error(1)
}

func (m *MockAuthServiceClient) Logout(ctx context.Context, in *gen.LogoutReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAuthServiceClient) ValidateSession(ctx context.Context, in *gen.ValidateSessionReq, opts ...grpc.CallOption) (*gen.ValidateSessionRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.ValidateSessionRes), args.Error(1)
}

func (m *MockAuthServiceClient) GetSessionsByUserID(ctx context.Context, in *gen.GetSessionsByUserIDReq, opts ...grpc.CallOption) (*gen.GetSessionsByUserIDRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.GetSessionsByUserIDRes), args.Error(1)
}

func (m *MockAuthServiceClient) DeleteSession(ctx context.Context, in *gen.DeleteSessionReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAuthServiceClient) DeleteAllSessionsExceptCurrent(ctx context.Context, in *gen.DeleteAllSessionsExceptCurrentReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	req := AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	sessionID := uuid.New().String()
	csrfToken := "test-csrf-token"

	mockAuthClient.On("Register", mock.Anything, mock.MatchedBy(func(r *gen.RegisterReq) bool {
		return r.PhoneNumber == req.PhoneNumber &&
			r.Password == req.Password &&
			r.Name == req.Name
	})).Return(&gen.RegisterRes{
		SessionId: sessionID,
		CsrfToken: csrfToken,
	}, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	recorder := httptest.NewRecorder()
	handler.Register(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response AuthDTO.AuthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, csrfToken, response.CSRFToken)

	cookies := recorder.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, sessionConfig.Signature, cookies[0].Name)
	assert.Equal(t, sessionID, cookies[0].Value)

	mockAuthClient.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.Register(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAuthHandler_Register_GRPCError(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	req := AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	mockAuthClient.On("Register", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.AlreadyExists, "user already exists"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.Register(recorder, request)

	assert.Equal(t, http.StatusConflict, recorder.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	req := AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}

	sessionID := uuid.New().String()
	csrfToken := "test-csrf-token"

	mockAuthClient.On("Login", mock.Anything, mock.MatchedBy(func(r *gen.LoginReq) bool {
		return r.PhoneNumber == req.PhoneNumber && r.Password == req.Password
	})).Return(&gen.LoginRes{
		SessionId: sessionID,
		CsrfToken: csrfToken,
	}, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.Login(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response AuthDTO.AuthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, csrfToken, response.CSRFToken)

	mockAuthClient.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	req := AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "wrongpassword",
	}

	mockAuthClient.On("Login", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.Unauthenticated, "invalid credentials"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.Login(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	sessionID := uuid.New().String()

	mockAuthClient.On("Logout", mock.Anything, mock.MatchedBy(func(r *gen.LogoutReq) bool {
		return r.SessionId == sessionID
	})).Return(&emptypb.Empty{}, nil)

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	request.AddCookie(&http.Cookie{
		Name:  sessionConfig.Signature,
		Value: sessionID,
	})

	recorder := httptest.NewRecorder()
	handler.Logout(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestAuthHandler_Logout_NoSession(t *testing.T) {
	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	mockAuthClient := new(MockAuthServiceClient)
	handler := NewAuthGRPCProxyHandler(mockAuthClient, sessionConfig)

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	recorder := httptest.NewRecorder()
	handler.Logout(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
