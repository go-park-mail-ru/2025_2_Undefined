package usecase

import (
	"context"
	"fmt"
	"log"

	ContactModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/google/uuid"
)

type ContactRepository interface {
	CreateContact(ctx context.Context, user_id uuid.UUID, contact_user_id uuid.UUID) error
	GetContactsByUserID(ctx context.Context, user_id uuid.UUID) ([]*ContactModels.Contact, error)
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error)
}

type ContactUsecase struct {
	contactrepo ContactRepository
	userrepo    UserRepository
}

func New(contactrepo ContactRepository, userrepo UserRepository) *ContactUsecase {
	return &ContactUsecase{
		contactrepo: contactrepo,
		userrepo:    userrepo,
	}
}

func (uc *ContactUsecase) CreateContact(ctx context.Context, req *ContactDTO.PostContactDTO, userID uuid.UUID) error {
	const op = "ContactUsecase.CreateContact"

	_, err := uc.userrepo.GetUserByID(ctx, userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	err = uc.contactrepo.CreateContact(ctx, userID, req.ContactUserID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}

func (uc *ContactUsecase) GetContacts(ctx context.Context, userID uuid.UUID) ([]*ContactDTO.GetContactsDTO, error) {
	const op = "ContactUsecase.GetContacts"
	contactsModels, err := uc.contactrepo.GetContactsByUserID(ctx, userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}

	if contactsModels == nil {
		return []*ContactDTO.GetContactsDTO{}, nil
	}

	ContactsDTO := make([]*ContactDTO.GetContactsDTO, len(contactsModels))
	for i, contact := range contactsModels {
		contactUserInfoModels, err := uc.userrepo.GetUserByID(ctx, contact.ContactUserID)
		if err != nil {
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, err
		}
		contactUserInfoDTO := &UserDTO.User{
			ID:          contactUserInfoModels.ID,
			PhoneNumber: contactUserInfoModels.PhoneNumber,
			Name:        contactUserInfoModels.Name,
			Username:    contactUserInfoModels.Username,
			Bio:         contactUserInfoModels.Bio,
			Avatar:      contactUserInfoModels.Avatar,
			AccountType: contactUserInfoModels.AccountType,
			CreatedAt:   contactUserInfoModels.CreatedAt,
			UpdatedAt:   contactUserInfoModels.UpdatedAt,
		}
		ContactsDTO[i] = &ContactDTO.GetContactsDTO{
			UserID:      contact.UserID,
			ContactUser: contactUserInfoDTO,
			CreatedAt:   contact.CreatedAt,
			UpdatedAt:   contact.UpdatedAt,
		}
	}
	return ContactsDTO, nil
}
