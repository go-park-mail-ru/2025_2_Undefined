package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	service "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockUserRepository struct {
	CreateFunc        func(user *UserModels.User) error
	GetByIDFunc       func(id uuid.UUID) (*UserModels.User, error)
	GetByPhoneFunc    func(phone string) (*UserModels.User, error)
	GetByEmailFunc    func(email string) (*UserModels.User, error)
	GetByUsernameFunc func(username string) (*UserModels.User, error)
	UpdateFunc        func(user *UserModels.User) error
}

func (m *MockUserRepository) Create(user *UserModels.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*UserModels.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) GetByPhone(phone string) (*UserModels.User, error) {
	if m.GetByPhoneFunc != nil {
		return m.GetByPhoneFunc(phone)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) GetByEmail(email string) (*UserModels.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) GetByUsername(username string) (*UserModels.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(username)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) Update(user *UserModels.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}

type MockTokenRepository struct {
	AddToBlacklistFunc func(token string) error
	IsInBlacklistFunc  func(token string) bool
}

func (m *MockTokenRepository) AddToBlacklist(token string) error {
	if m.AddToBlacklistFunc != nil {
		return m.AddToBlacklistFunc(token)
	}
	return nil
}

func (m *MockTokenRepository) IsInBlacklist(token string) bool {
	if m.IsInBlacklistFunc != nil {
		return m.IsInBlacklistFunc(token)
	}
	return false
}

func (m *MockTokenRepository) CleanupExpiredTokens() {}

func TestAuthHandler_Register_Integration(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	userRepo.GetByEmailFunc = func(email string) (*UserModels.User, error) {
		return nil, errors.New("not found")
	}
	userRepo.GetByPhoneFunc = func(phone string) (*UserModels.User, error) {
		return nil, errors.New("not found")
	}
	userRepo.GetByUsernameFunc = func(username string) (*UserModels.User, error) {
		return nil, errors.New("not found")
	}
	userRepo.CreateFunc = func(user *UserModels.User) error {
		return nil
	}

	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Email:       "test@mail.ru",
		Username:    "testuser",
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
}

func TestAuthHandler_Register_BadJSON(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_Register_MissingFields(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "",
		Email:       "test@mail.ru",
		Username:    "testuser",
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

func TestAuthHandler_Register_InvalidPhone(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	registerReq := AuthModels.RegisterRequest{
		PhoneNumber: "invalid-phone",
		Email:       "test@mail.ru",
		Username:    "testuser",
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

func TestAuthHandler_Login_Integration(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	userRepo.GetByPhoneFunc = func(phone string) (*UserModels.User, error) {
		hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
		return &UserModels.User{
			ID:           uuid.New(),
			PhoneNumber:  "+79998887766",
			PasswordHash: hashedPassword,
		}, nil
	}

	loginReq := AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	userRepo.GetByPhoneFunc = func(phone string) (*UserModels.User, error) {
		return nil, errors.New("not found")
	}

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
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{
		AddToBlacklistFunc: func(token string) error {
			return nil
		},
	}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	token, _ := tokenator.CreateJWT(uuid.New().String())

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: token})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Logout_NoCookie(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_GetCurrentUser_Success(t *testing.T) {
	userID := uuid.New()
	expectedUser := &UserModels.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@mail.ru",
	}

	userRepo := &MockUserRepository{
		GetByIDFunc: func(id uuid.UUID) (*UserModels.User, error) {
			return expectedUser, nil
		},
	}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	token, _ := tokenator.CreateJWT(userID.String())

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: token})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthHandler_GetCurrentUser_NoCookie(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_GetCurrentUser_InvalidToken(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenator := jwt.NewTokenator()
	tokenRepo := &MockTokenRepository{}

	authService := service.NewAuthService(userRepo, tokenator, tokenRepo)
	handler := NewAuthHandler(authService)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: domains.TokenCookieName, Value: "invalid-token"})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
