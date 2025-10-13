package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	UserAccount     = "user"
	PremiumAccount  = "premium"
	VerifiedAccount = "verified"
)

type User struct {
	ID           uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	PhoneNumber  string    `json:"phone_number"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Bio          string    `json:"bio"`
	AccountType  string    `json:"account_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
