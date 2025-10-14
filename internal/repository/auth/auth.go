package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	createUserQuery = `
		INSERT INTO "user" (id, username, name, phone_number, password_hash, user_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::user_type_enum, $7, $8)
		RETURNING id, username, phone_number, user_type`
)

type AuthRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateUser(name string, phone string, password_hash string) (*models.User, error) {
	const op = "AuthRepository.CreateUser"
	newUsername, err := r.generateUsername()
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	user := &models.User{
		ID:           uuid.New(),
		PhoneNumber:  phone,
		PasswordHash: password_hash,
		Name:         name,
		Username:     newUsername,
		AccountType:  models.UserAccount,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = r.db.QueryRow(createUserQuery,
		user.ID, user.Username, user.Name, user.PhoneNumber, user.PasswordHash, user.AccountType, user.CreatedAt, user.UpdatedAt).
		Scan(&user.ID, &user.Username, &user.PhoneNumber, &user.AccountType)

	if err != nil {
		// Проверяем является ли ошибка наруением уникального ограничения
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			err = errs.ErrIsDuplicateKey
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

func (r *AuthRepository) generateUsername() (string, error) {
	const op = "AuthRepository.generateUsername"
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM "user"`).Scan(&count)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return "", err
	}

	username := fmt.Sprintf("user_%d", count)

	exists, err := r.checkUsernameExists(username)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return "", err
	}
	if !exists {
		return username, nil
	}

	//если не получилось создать юзернейм по умолчанию, то создаем через uuid
	return "user_" + uuid.New().String()[:8], nil
}

func (r *AuthRepository) checkUsernameExists(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM "user" WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}
