package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	InterfaceContactRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/contact"
	InterfaceFileStorage "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/storage"
	InterfaceUserRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/user"
	"github.com/google/uuid"
)

type ContactUsecase struct {
	contactrepo InterfaceContactRepository.ContactRepository
	userrepo    InterfaceUserRepository.UserRepository
	fileStorage InterfaceFileStorage.FileStorage
}

func New(contactrepo InterfaceContactRepository.ContactRepository, userrepo InterfaceUserRepository.UserRepository, fileStorage InterfaceFileStorage.FileStorage) *ContactUsecase {
	return &ContactUsecase{
		contactrepo: contactrepo,
		userrepo:    userrepo,
		fileStorage: fileStorage,
	}
}

func (uc *ContactUsecase) CreateContact(ctx context.Context, req *ContactDTO.PostContactDTO, userID uuid.UUID) error {
	const op = "ContactUsecase.CreateContact"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	_, err := uc.userrepo.GetUserByID(ctx, userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to get user by ID")
		return wrappedErr
	}

	err = uc.contactrepo.CreateContact(ctx, userID, req.ContactUserID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to create contact")
		return wrappedErr
	}

	return nil
}

func (uc *ContactUsecase) GetContacts(ctx context.Context, userID uuid.UUID) ([]*ContactDTO.GetContactsDTO, error) {
	const op = "ContactUsecase.GetContacts"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	contactsModels, err := uc.contactrepo.GetContactsByUserID(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get contacts by user ID")
		return nil, err
	}

	if contactsModels == nil {
		return []*ContactDTO.GetContactsDTO{}, nil
	}

	ContactsDTO := make([]*ContactDTO.GetContactsDTO, len(contactsModels))
	for i, contact := range contactsModels {
		contactUserInfoModels, err := uc.userrepo.GetUserByID(ctx, contact.ContactUserID)
		if err != nil {
			logger.WithError(err).Error("failed to get contact user info by ID")
			return nil, err
		}

		contactUserInfoDTO := &UserDTO.User{
			ID:          contactUserInfoModels.ID,
			PhoneNumber: contactUserInfoModels.PhoneNumber,
			Name:        contactUserInfoModels.Name,
			Username:    contactUserInfoModels.Username,
			Bio:         contactUserInfoModels.Bio,
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
