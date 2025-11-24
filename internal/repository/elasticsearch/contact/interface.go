package contact

import (
	"context"
)

type ContactSearchRepositoryInterface interface {
	CreateIndex(ctx context.Context) error
	IndexContact(ctx context.Context, userID, contactUserID, username, name, phoneNumber string) error
	SearchContacts(ctx context.Context, userID, query string) ([]map[string]interface{}, error)
	DeleteContact(ctx context.Context, userID, contactUserID string) error
}
