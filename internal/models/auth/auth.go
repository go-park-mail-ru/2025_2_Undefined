package models

import (
	"time"

	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
)

type RegisterRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required,min=6"`
	Name        string `json:"name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string                `json:"token"`
	User  UserModels.PublicUser `json:"user"`
}

type TokenBlacklist struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
