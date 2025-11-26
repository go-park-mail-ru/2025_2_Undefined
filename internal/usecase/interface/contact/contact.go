package contact

import (
	"context"

	ContactModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	"github.com/google/uuid"
)

type ContactRepository interface {
	CreateContact(ctx context.Context, user_id uuid.UUID, contact_user_id uuid.UUID) error
	GetContactsByUserID(ctx context.Context, user_id uuid.UUID) ([]*ContactModels.Contact, error)
	GetAllContacts(ctx context.Context) ([]*ContactModels.Contact, error)
}
