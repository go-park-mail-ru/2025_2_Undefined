package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	UserAccount int = iota
	PremiumAccount
	VerifiedAccount
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email" validate:"required,email"`
	PhoneNumber  string    `json:"phone_number"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Bio          string    `json:"bio"`
	AccountType  int       `json:"account_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PublicUser struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	Bio         string    `json:"bio"`
	AccountType int       `json:"account_type"`
	CreatedAt   time.Time `json:"created_at"`
}
