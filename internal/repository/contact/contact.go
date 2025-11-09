package repository

import (
	"context"
	"database/sql"
	"time"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	createContactQuery = `
		INSERT INTO contact (user_id, contact_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		 RETURNING user_id`

	getContactsByUserIDQuery = `
		SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact
		WHERE user_id = $1`
)

type ContactRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *ContactRepository {
	return &ContactRepository{
		db: db,
	}
}

func (r *ContactRepository) CreateContact(ctx context.Context, user_id uuid.UUID, contact_user_id uuid.UUID) error {
	const op = "ContactRepository.CreateContact"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", user_id.String()).WithField("contact_user_id", contact_user_id.String())
	logger.Debug("Starting database operation: create contact")

	err := r.db.QueryRow(createContactQuery,
		user_id, contact_user_id, time.Now(), time.Now()).
		Scan(&user_id)
	if err != nil {
		// Проверяем является ли ошибка наруением уникального ограничения
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			err = errs.ErrIsDuplicateKey
			logger.WithError(err).Error("Database operation failed: duplicate key constraint violation")
			return err
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			err = errs.ErrUserNotFound
			logger.WithError(err).Error("Database operation failed: user not found")
			return err
		}
		logger.WithError(err).Error("Database operation failed: create contact query")
		return err
	}

	logger.Info("Database operation completed successfully: contact created")
	return nil
}

func (r *ContactRepository) GetContactsByUserID(ctx context.Context, user_id uuid.UUID) ([]*models.Contact, error) {
	const op = "ContactRepository.GetContactsByUserID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", user_id.String())
	logger.Debug("Starting database operation: get contacts by user ID")

	rows, err := r.db.Query(getContactsByUserIDQuery, user_id)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get contacts query")
		return nil, err
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		var contact models.Contact
		err := rows.Scan(&contact.UserID, &contact.ContactUserID, &contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan contact row")
			return nil, err
		}
		contacts = append(contacts, &contact)
	}
	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("contacts_count", len(contacts)).Info("Database operation completed successfully: contacts retrieved")
	return contacts, nil
}
