package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
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

func (r *ContactRepository) CreateContact(user_id uuid.UUID, contact_user_id uuid.UUID) error {
	const op = "ContactRepository.CreateContact"
	err := r.db.QueryRow(createContactQuery,
		user_id, contact_user_id, time.Now(), time.Now()).
		Scan(&user_id)
	if err != nil {
		// Проверяем является ли ошибка наруением уникального ограничения
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			err = errs.ErrIsDuplicateKey
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return err
		}
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return err
	}
	return nil
}

func (r *ContactRepository) GetContactsByUserID(user_id uuid.UUID) ([]*models.Contact, error) {
	const op = "ContactRepository.GetContactsByUserID"
	rows, err := r.db.Query(getContactsByUserIDQuery, user_id)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		var contact models.Contact
		err := rows.Scan(&contact.UserID, &contact.ContactUserID, &contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, wrappedErr
		}
		contacts = append(contacts, &contact)
	}
	if err = rows.Err(); err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	return contacts, nil
}
