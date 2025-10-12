package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockAuthUsecase struct {
	RegisterFunc    func(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO)
	LoginFunc       func(req *AuthModels.LoginRequest) (string, error)
	LogoutFunc      func(tokenString string) error
	GetUserByIdFunc func(id uuid.UUID) (*UserModels.User, error)
}

func (m *MockAuthUsecase) Register(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(req)
	}
	return "", nil
}

func (m *MockAuthUsecase) Login(req *AuthModels.LoginRequest) (string, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(req)
	}
	return "", nil
}

func (m *MockAuthUsecase) Logout(tokenString string) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(tokenString)
	}
	return nil
}

func (m *MockAuthUsecase) GetUserById(id uuid.UUID) (*UserModels.User, error) {
	if m.GetUserByIdFunc != nil {
		return m.GetUserByIdFunc(id)
	}
	return nil, nil
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockUsecase := &MockAuthUsecase{
		RegisterFunc: func(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO) {
			assert.Equal(t, "+79998887766", req.PhoneNumber)
			assert.Equal(t, "Test User", req.Name)
			assert.Equal(t, "password123", req.Password)
			return "test.jwt.token", nil
		},
	}

	handler := New(mockUsecase)

	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Password:    "password123",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	cookies := resp.Cookies()
	var tokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == domains.TokenCookieName {
			tokenCookie = cookie
			break
		}
	}
	assert.NotNil(t, tokenCookie)
	assert.Equal(t, "test.jwt.token", tokenCookie.Value)
}

func TestAuthHandler_Register_BadJSON(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_Register_ValidationErrors(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "invalid-phone",
		Name:        "",
		Password:    "123",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var validationResponse dto.ValidationErrorsDTO
	err := json.NewDecoder(resp.Body).Decode(&validationResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Ошибка валидации", validationResponse.Message)
	assert.True(t, len(validationResponse.Errors) > 0)
}

func TestAuthHandler_Register_UsecaseError(t *testing.T) {
	mockUsecase := &MockAuthUsecase{
		RegisterFunc: func(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO) {
			return "", &dto.ValidationErrorsDTO{
				Message: "phone already exists",
				Errors: []dto.ValidationErrorDTO{
					{Field: "phone_number", Message: "пользователь с таким телефоном уже существует"},
				},
			}
		},
	}

	handler := New(mockUsecase)

	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Password:    "password123",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_Register_EmptyToken(t *testing.T) {
	mockUsecase := &MockAuthUsecase{
		RegisterFunc: func(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO) {
			return "", nil
		},
	}

	handler := New(mockUsecase)

	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Password:    "password123",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUsecase := &MockAuthUsecase{
		LoginFunc: func(req *AuthModels.LoginRequest) (string, error) {
			assert.Equal(t, "+79998887766", req.PhoneNumber)
			assert.Equal(t, "password123", req.Password)
			return "test.jwt.token", nil
		},
	}

	handler := New(mockUsecase)

	loginReq := AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cookies := resp.Cookies()
	var tokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == domains.TokenCookieName {
			tokenCookie = cookie
			break
		}
	}
	assert.NotNil(t, tokenCookie)
	assert.Equal(t, "test.jwt.token", tokenCookie.Value)
}

func TestAuthHandler_Login_BadJSON(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_Login_ValidationErrors(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	loginReq := AuthModels.LoginRequest{
		PhoneNumber: "invalid-phone",
		Password:    "кириллица",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var validationResponse dto.ValidationErrorsDTO
	err := json.NewDecoder(resp.Body).Decode(&validationResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Ошибка валидации", validationResponse.Message)
	assert.True(t, len(validationResponse.Errors) > 0)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockUsecase := &MockAuthUsecase{
		LoginFunc: func(req *AuthModels.LoginRequest) (string, error) {
			return "", errs.ErrInvalidCredentials
		},
	}

	handler := New(mockUsecase)

	loginReq := AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "wrongpassword",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockUsecase := &MockAuthUsecase{
		LogoutFunc: func(tokenString string) error {
			assert.Equal(t, "test.jwt.token", tokenString)
			return nil
		},
	}

	handler := New(mockUsecase)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: "test.jwt.token"})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cookies := resp.Cookies()
	var tokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == domains.TokenCookieName {
			tokenCookie = cookie
			break
		}
	}
	assert.NotNil(t, tokenCookie)
	assert.Equal(t, "", tokenCookie.Value)
	assert.True(t, tokenCookie.Expires.Before(time.Now()))
}

func TestAuthHandler_Logout_NoCookie(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Logout_UsecaseError(t *testing.T) {
	mockUsecase := &MockAuthUsecase{
		LogoutFunc: func(tokenString string) error {
			return errors.New("logout failed")
		},
	}

	handler := New(mockUsecase)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: "test.jwt.token"})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_GetCurrentUser_Success(t *testing.T) {
	userID := uuid.New()
	expectedUser := &UserModels.User{
		ID:          userID,
		Name:        "Test User",
		PhoneNumber: "+79998887766",
	}

	mockUsecase := &MockAuthUsecase{
		GetUserByIdFunc: func(id uuid.UUID) (*UserModels.User, error) {
			assert.Equal(t, userID, id)
			return expectedUser, nil
		},
	}

	handler := New(mockUsecase)

	tokenator := jwt.NewTokenator()
	token, _ := tokenator.CreateJWT(userID.String())

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: token})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var user UserModels.User
	err := json.NewDecoder(resp.Body).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)
}

func TestAuthHandler_GetCurrentUser_NoCookie(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_GetCurrentUser_InvalidToken(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: "invalid-token"})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_GetCurrentUser_InvalidUserID(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	tokenator := jwt.NewTokenator()
	token, _ := tokenator.CreateJWT("invalid-uuid")

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: token})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_GetCurrentUser_UserNotFound(t *testing.T) {
	userID := uuid.New()

	mockUsecase := &MockAuthUsecase{
		GetUserByIdFunc: func(id uuid.UUID) (*UserModels.User, error) {
			return nil, errors.New("user not found")
		},
	}

	handler := New(mockUsecase)

	tokenator := jwt.NewTokenator()
	token, _ := tokenator.CreateJWT(userID.String())

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: token})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	cookies := resp.Cookies()
	var tokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == domains.TokenCookieName {
			tokenCookie = cookie
			break
		}
	}
	assert.NotNil(t, tokenCookie)
	assert.Equal(t, "", tokenCookie.Value)
	assert.True(t, tokenCookie.Expires.Before(time.Now()))
}

func TestAuthHandler_Register_MultipleValidationErrors(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	// Создаем запрос с множественными ошибками валидации
	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "invalid-phone",
		Name:        "",
		Password:    "123",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var validationResponse dto.ValidationErrorsDTO
	err := json.NewDecoder(resp.Body).Decode(&validationResponse)
	assert.NoError(t, err)

	assert.Equal(t, "Ошибка валидации", validationResponse.Message)
	assert.True(t, len(validationResponse.Errors) > 1, "Should return multiple validation errors")

	errorFields := make(map[string]bool)
	for _, errorDTO := range validationResponse.Errors {
		errorFields[errorDTO.Field] = true
	}

	expectedFields := []string{"phone_number", "name", "password"}
	for _, field := range expectedFields {
		assert.True(t, errorFields[field], "Should have validation error for field: %s", field)
	}
}

func TestAuthHandler_Login_MultipleValidationErrors(t *testing.T) {
	mockUsecase := &MockAuthUsecase{}
	handler := New(mockUsecase)

	loginReq := AuthModels.LoginRequest{
		PhoneNumber: "invalid-phone",
		Password:    "кириллица",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var validationResponse dto.ValidationErrorsDTO
	err := json.NewDecoder(resp.Body).Decode(&validationResponse)
	assert.NoError(t, err)

	assert.Equal(t, "Ошибка валидации", validationResponse.Message)
	assert.True(t, len(validationResponse.Errors) > 1, "Should return multiple validation errors")

	errorFields := make(map[string]bool)
	for _, errorDTO := range validationResponse.Errors {
		errorFields[errorDTO.Field] = true
	}

	assert.True(t, errorFields["phone_number"], "Should have validation error for phone_number")
	assert.True(t, errorFields["password"], "Should have validation error for password")
}

func TestAuthHandler_Register_EdgeCases(t *testing.T) {
	t.Run("Empty request body", func(t *testing.T) {
		mockUsecase := &MockAuthUsecase{}
		handler := New(mockUsecase)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Null JSON", func(t *testing.T) {
		mockUsecase := &MockAuthUsecase{}
		handler := New(mockUsecase)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("null"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuthHandler_Login_EdgeCases(t *testing.T) {
	t.Run("Empty request body", func(t *testing.T) {
		mockUsecase := &MockAuthUsecase{}
		handler := New(mockUsecase)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Null JSON", func(t *testing.T) {
		mockUsecase := &MockAuthUsecase{}
		handler := New(mockUsecase)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("null"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
