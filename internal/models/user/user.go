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
	ID           uuid.UUID
	PhoneNumber  string
	PasswordHash string
	Name         string
	Username     string
	Bio          string
	Avatar       *string
	AccountType  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
