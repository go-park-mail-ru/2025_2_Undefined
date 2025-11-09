package dto

import (
	"time"

	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/google/uuid"
)

type ContactDTO struct {
	UserID        uuid.UUID `json:"user_id"`
	ContactUserID uuid.UUID `json:"contact_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"update_at"`
}

type PostContactDTO struct {
	ContactUserID uuid.UUID `json:"contact_id"`
}

type GetContactsDTO struct {
	UserID      uuid.UUID `json:"user_id"`
	ContactUser *dto.User `json:"contact"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"update_at"`
}
