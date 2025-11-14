package repository

import (
	"context"
	"database/sql"
	"errors"
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
	const query = "INSERT contact"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("user_id", user_id.String()).
		WithField("contact_user_id", contact_user_id.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	err := r.db.QueryRow(createContactQuery,
		user_id, contact_user_id, time.Now(), time.Now()).
		Scan(&user_id)
	if err != nil {
		queryStatus = "fail"

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case errs.PostgresErrorUniqueViolationCode:
				logger.WithError(err).Errorf("db query: %s: duplicate key violation: status: %s", query, queryStatus)
				return errs.ErrIsDuplicateKey
			case errs.PostgresErrorForeignKeyViolationCode:
				logger.WithError(err).Errorf("db query: %s: foreign key violation: status: %s", query, queryStatus)
				return errs.ErrUserNotFound
			}
		}
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return err
	}

	return nil
}

func (r *ContactRepository) GetContactsByUserID(ctx context.Context, user_id uuid.UUID) ([]*models.Contact, error) {
	const op = "ContactRepository.GetContactsByUserID"
	const query = "SELECT contacts"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", user_id.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	rows, err := r.db.Query(getContactsByUserIDQuery, user_id)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		var contact models.Contact
		err := rows.Scan(&contact.UserID, &contact.ContactUserID, &contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}
		contacts = append(contacts, &contact)
	}

	if err = rows.Err(); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: rows iteration error: status: %s", query, queryStatus)
		return nil, err
	}

	return contacts, nil
}
