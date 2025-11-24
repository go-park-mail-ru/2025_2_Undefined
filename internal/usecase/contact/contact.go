package usecase

import (
	"context"
	"fmt"

	ContactModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	contactES "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/elasticsearch/contact"
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
	esClient    contactES.ContactSearchRepositoryInterface
}

func New(contactrepo InterfaceContactRepository.ContactRepository, userrepo InterfaceUserRepository.UserRepository, fileStorage InterfaceFileStorage.FileStorage, esClient contactES.ContactSearchRepositoryInterface) *ContactUsecase {
	return &ContactUsecase{
		contactrepo: contactrepo,
		userrepo:    userrepo,
		fileStorage: fileStorage,
		esClient:    esClient,
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

	if uc.esClient != nil {
		contactUser, err := uc.userrepo.GetUserByID(ctx, req.ContactUserID)
		if err == nil {
			logger.Debug("indexing contact in elasticsearch")
			if indexErr := uc.esClient.IndexContact(
				ctx,
				userID.String(),
				contactUser.ID.String(),
				contactUser.Username,
				contactUser.Name,
				contactUser.PhoneNumber,
			); indexErr != nil {
				logger.WithError(indexErr).Warn("failed to index contact in elasticsearch")
			} else {
				logger.Debug("contact indexed successfully")
			}
		} else {
			logger.WithError(err).Warn("failed to get contact user for indexing")
		}
	} else {
		logger.Warn("elasticsearch client is nil, skipping indexing")
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

func (uc *ContactUsecase) SearchContacts(ctx context.Context, userID uuid.UUID, query string) ([]*ContactDTO.GetContactsDTO, error) {
	const op = "ContactUsecase.SearchContacts"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if query == "" {
		return []*ContactDTO.GetContactsDTO{}, nil
	}

	if uc.esClient != nil {
		logger.WithField("query", query).Debug("searching contacts in elasticsearch")
		results, err := uc.esClient.SearchContacts(ctx, userID.String(), query)
		if err != nil {
			logger.WithError(err).Warn("elasticsearch search failed, falling back to database")
		} else {
			logger.WithField("results_count", len(results)).Debug("elasticsearch search completed")
			contacts := make([]*ContactDTO.GetContactsDTO, 0, len(results))
			for _, result := range results {
				contactUserIDStr, ok := result["contact_user_id"].(string)
				if !ok {
					continue
				}

				contactUserID, err := uuid.Parse(contactUserIDStr)
				if err != nil {
					continue
				}

				user, err := uc.userrepo.GetUserByID(ctx, contactUserID)
				if err != nil {
					logger.WithError(err).Warn("failed to get user info for search result")
					continue
				}

				contactsModels, err := uc.contactrepo.GetContactsByUserID(ctx, userID)
				if err != nil {
					continue
				}

				var contactModel *ContactModels.Contact
				for _, contact := range contactsModels {
					if contact.ContactUserID == contactUserID {
						contactModel = contact
						break
					}
				}

				if contactModel == nil {
					continue
				}

				contacts = append(contacts, &ContactDTO.GetContactsDTO{
					UserID: userID,
					ContactUser: &UserDTO.User{
						ID:          user.ID,
						PhoneNumber: user.PhoneNumber,
						Name:        user.Name,
						Username:    user.Username,
						Bio:         user.Bio,
						AccountType: user.AccountType,
						CreatedAt:   user.CreatedAt,
						UpdatedAt:   user.UpdatedAt,
					},
					CreatedAt: contactModel.CreatedAt,
					UpdatedAt: contactModel.UpdatedAt,
				})
			}

			return contacts, nil
		}
	}

	logger.Warn("elasticsearch not available, returning empty results")
	return []*ContactDTO.GetContactsDTO{}, nil
}

func (uc *ContactUsecase) ReindexAllContacts(ctx context.Context) error {
	const op = "ContactUsecase.ReindexAllContacts"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if uc.esClient == nil {
		logger.Warn("elasticsearch client is nil, skipping reindexing")
		return nil
	}

	logger.Info("starting reindexing of all contacts")

	contacts, err := uc.contactrepo.GetAllContacts(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to get all contacts from database")
		return fmt.Errorf("%s: %w", op, err)
	}

	logger.WithField("total_contacts", len(contacts)).Info("retrieved contacts from database")

	indexed := 0
	failed := 0

	for _, contact := range contacts {
		user, err := uc.userrepo.GetUserByID(ctx, contact.ContactUserID)
		if err != nil {
			logger.WithError(err).WithField("contact_user_id", contact.ContactUserID).Warn("failed to get user info for contact, skipping")
			failed++
			continue
		}

		err = uc.esClient.IndexContact(
			ctx,
			contact.UserID.String(),
			contact.ContactUserID.String(),
			user.Username,
			user.Name,
			user.PhoneNumber,
		)
		if err != nil {
			logger.WithError(err).WithField("user_id", contact.UserID).WithField("contact_user_id", contact.ContactUserID).Warn("failed to index contact")
			failed++
			continue
		}
		indexed++
	}

	logger.WithField("indexed", indexed).WithField("failed", failed).WithField("total", len(contacts)).Info("reindexing completed")

	if failed > 0 {
		return fmt.Errorf("%s: reindexing completed with %d failures out of %d contacts", op, failed, len(contacts))
	}

	return nil
}
