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
	Bio          string    `json:"bio"`
	Avatar       *string   `json:"avatar"`
	AccountType  string    `json:"account_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
