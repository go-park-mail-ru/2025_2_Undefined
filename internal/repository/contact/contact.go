package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

	deleteContactQuery = `
		DELETE FROM contact 
		WHERE user_id = $1 AND contact_user_id = $2`

	checkContactExistsQuery = `
		SELECT EXISTS(
			SELECT 1 FROM contact 
			WHERE user_id = $1 AND contact_user_id = $2
		)`
)

type ContactRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{
		pool: pool,
	}
}

func (r *ContactRepository) CreateContact(ctx context.Context, userID uuid.UUID, contactUserID uuid.UUID) error {
	const op = "ContactRepository.CreateContact"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String()).WithField("contact_user_id", contactUserID.String())
	logger.Debug("Starting database operation: create contact")

	var resultUserID uuid.UUID
	err := r.pool.QueryRow(ctx, createContactQuery,
		userID, contactUserID, time.Now(), time.Now()).
		Scan(&resultUserID)
	if err != nil {
		// Проверяем является ли ошибка нарушением уникального ограничения
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				logger.WithError(err).Error("Database operation failed: duplicate key constraint violation")
				return errs.ErrIsDuplicateKey
			}
			if pgErr.Code == "23503" { // foreign_key_violation
				logger.WithError(err).Error("Database operation failed: user not found")
				return errs.ErrUserNotFound
			}
		}
		logger.WithError(err).Error("Database operation failed: create contact query")
		return fmt.Errorf("failed to create contact: %w", err)
	}

	logger.Info("Database operation completed successfully: contact created")
	return nil
}

func (r *ContactRepository) GetContactsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Contact, error) {
	const op = "ContactRepository.GetContactsByUserID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: get contacts by user ID")

	rows, err := r.pool.Query(ctx, getContactsByUserIDQuery, userID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get contacts query")
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		var contact models.Contact
		err := rows.Scan(&contact.UserID, &contact.ContactUserID, &contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan contact row")
			return nil, fmt.Errorf("failed to scan contact row: %w", err)
		}
		contacts = append(contacts, &contact)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	logger.WithField("contacts_count", len(contacts)).Info("Database operation completed successfully: contacts retrieved")
	return contacts, nil
}

func (r *ContactRepository) DeleteContact(ctx context.Context, userID, contactUserID uuid.UUID) error {
	const op = "ContactRepository.DeleteContact"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String()).WithField("contact_user_id", contactUserID.String())
	logger.Debug("Starting database operation: delete contact")

	cmdTag, err := r.pool.Exec(ctx, deleteContactQuery, userID, contactUserID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: delete contact")
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Debug("Database operation completed: contact not found")
		return fmt.Errorf("contact not found")
	}

	logger.Info("Database operation completed successfully: contact deleted")
	return nil
}

func (r *ContactRepository) CheckContactExists(ctx context.Context, userID, contactUserID uuid.UUID) (bool, error) {
	const op = "ContactRepository.CheckContactExists"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String()).WithField("contact_user_id", contactUserID.String())
	logger.Debug("Starting database operation: check contact exists")

	var exists bool
	err := r.pool.QueryRow(ctx, checkContactExistsQuery, userID, contactUserID).Scan(&exists)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check contact exists query")
		return false, fmt.Errorf("failed to check contact exists: %w", err)
	}

	logger.WithField("exists", exists).Info("Database operation completed successfully: contact existence checked")
	return exists, nil
}
