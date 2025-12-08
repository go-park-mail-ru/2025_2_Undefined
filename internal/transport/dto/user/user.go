package dto

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	PhoneNumber  string    `json:"phone_number"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Bio          *string   `json:"bio,omitempty"`
	AccountType  string    `json:"account_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GetUserByPhone struct {
	PhoneNumber string `json:"phone_number"`
}

type GetUserByUsername struct {
	Username string `json:"username"`
}

type UpdateUserInfo struct {
	Name     *string `json:"name,omitempty"`
	Username *string `json:"username,omitempty"`
	Bio      *string `json:"bio,omitempty"`
}
