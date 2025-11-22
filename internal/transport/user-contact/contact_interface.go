package user

import (
	"context"

	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contact_interface.go -destination=../../usecase/mocks/mock_contact_usecase.go -package=mocks IContactUsecase
type IContactUsecase interface {
	CreateContact(ctx context.Context, req *ContactDTO.PostContactDTO, userID uuid.UUID) error
	GetContacts(ctx context.Context, userID uuid.UUID) ([]*ContactDTO.GetContactsDTO, error)
}
