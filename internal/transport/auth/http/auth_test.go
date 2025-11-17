package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) Register(ctx context.Context, req *AuthDTO.RegisterRequest, device string) (uuid.UUID, *dto.ValidationErrorsDTO) {
	args := m.Called(ctx, req, device)
	if args.Get(1) == nil {
		return args.Get(0).(uuid.UUID), nil
	}
	return args.Get(0).(uuid.UUID), args.Get(1).(*dto.ValidationErrorsDTO)
}

func (m *MockAuthUsecase) Login(ctx context.Context, req *AuthDTO.LoginRequest, device string) (uuid.UUID, error) {
	args := m.Called(ctx, req, device)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAuthUsecase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

type MockSessionUtils struct {
	mock.Mock
}

func (m *MockSessionUtils) GetUserIDFromSession(r *http.Request) (uuid.UUID, error) {
	args := m.Called(r)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	req := AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	sessionID := uuid.New()
	mockUsecase.On("Register", mock.Anything, &req, mock.AnythingOfType("string")).
		Return(sessionID, nil)

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
	assert.NotEmpty(t, response.CSRFToken)

	cookies := recorder.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, sessionConfig.Signature, cookies[0].Name)

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Register(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	req := AuthDTO.RegisterRequest{
		PhoneNumber: "invalid_phone",
		Password:    "123",
		Name:        "",
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Register(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAuthHandler_Register_UsecaseError(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	req := AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	validationErr := &dto.ValidationErrorsDTO{
		Message: "phone already exists",
		Errors: []dto.ValidationErrorDTO{
			{Field: "phone_number", Message: "already exists"},
		},
	}

	mockUsecase.On("Register", mock.Anything, &req, mock.AnythingOfType("string")).
		Return(uuid.Nil, validationErr)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Register(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Register_NilSessionID(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	req := AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	mockUsecase.On("Register", mock.Anything, &req, mock.AnythingOfType("string")).
		Return(uuid.Nil, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Register(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	req := AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}

	sessionID := uuid.New()
	mockUsecase.On("Login", mock.Anything, &req, mock.AnythingOfType("string")).
		Return(sessionID, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")

	recorder := httptest.NewRecorder()

	handler.Login(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response AuthDTO.AuthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.CSRFToken)

	cookies := recorder.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, sessionConfig.Signature, cookies[0].Name)

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Login(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	req := AuthDTO.LoginRequest{
		PhoneNumber: "",
		Password:    "",
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Login(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	req := AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "wrongpassword",
	}

	mockUsecase.On("Login", mock.Anything, &req, mock.AnythingOfType("string")).
		Return(uuid.Nil, errs.ErrInvalidCredentials)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Login(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	userID := uuid.New()
	sessionID := uuid.New()

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("Logout", mock.Anything, sessionID).Return(nil)

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	request.AddCookie(&http.Cookie{
		Name:  sessionConfig.Signature,
		Value: sessionID.String(),
	})

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Logout_Unauthorized(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(uuid.Nil, errors.New("unauthorized"))

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestAuthHandler_Logout_NoCookie(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	userID := uuid.New()
	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestAuthHandler_Logout_InvalidSessionID(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	userID := uuid.New()
	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	request.AddCookie(&http.Cookie{
		Name:  sessionConfig.Signature,
		Value: "invalid-uuid",
	})

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestAuthHandler_Logout_UsecaseError(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	mockSessionUtils := new(MockSessionUtils)

	sessionConfig := &config.SessionConfig{Signature: "test_signature"}
	csrfConfig := &config.CSRFConfig{Secret: "test_secret"}

	handler := New(mockUsecase, sessionConfig, csrfConfig, mockSessionUtils)

	userID := uuid.New()
	sessionID := uuid.New()

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("Logout", mock.Anything, sessionID).Return(errors.New("logout error"))

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	request.AddCookie(&http.Cookie{
		Name:  sessionConfig.Signature,
		Value: sessionID.String(),
	})

	recorder := httptest.NewRecorder()

	handler.Logout(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func Test_getDeviceFromUserAgent(t *testing.T) {
	tests := []struct {
		userAgent string
		contains  []string
	}{
		{
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			contains:  []string{"Chrome", "Windows"},
		},
		{
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			contains:  []string{"Safari", "CPU iPhone OS"},
		},
		{
			userAgent: "",
			contains:  []string{"Unknown Device"},
		},
		{
			userAgent: "curl/7.68.0",
			contains:  []string{"curl"},
		},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		if tt.userAgent != "" {
			request.Header.Set("User-Agent", tt.userAgent)
		}

		result := getDeviceFromUserAgent(request)

		found := false
		for _, expectedPart := range tt.contains {
			if assert.Contains(t, result, expectedPart) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected result '%s' to contain one of %v", result, tt.contains)
		}
	}
}
