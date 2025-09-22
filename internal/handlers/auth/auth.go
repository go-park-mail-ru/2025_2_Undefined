package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	service "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/auth"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthModels.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.PhoneNumber == "" || req.Email == "" || req.Username == "" || req.Password == "" || req.Name == "" {
		http.Error(w, `{"error": "All fields are required"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthModels.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "Email and password are required"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Извлекаем токен из заголовка
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, `{"error": "Invalid authorization header format"}`, http.StatusUnauthorized)
		return
	}

	tokenString := parts[1]
	if tokenString == "" {
		http.Error(w, `{"error": "Empty token"}`, http.StatusUnauthorized)
		return
	}

	// Добавляем токен в blacklist
	if err := h.authService.Logout(tokenString); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*UserModels.User)
	if !ok || user == nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusUnauthorized)
		return
	}

	publicUser := UserModels.PublicUser{
		ID:          user.ID,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		AccountType: user.AccountType,
		CreatedAt:   user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicUser)
}
