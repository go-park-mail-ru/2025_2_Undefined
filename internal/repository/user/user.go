package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
)

const (
	getUserByPhoneQuery = `
		SELECT id, username, name, phone_number, password_hash, user_type, created_at, updated_at
		FROM "user"
		WHERE phone_number = $1`

	getUserByUsernameQuery = `
		SELECT id, username, name, phone_number, password_hash, user_type, created_at, updated_at
		FROM "user"
		WHERE username = $1`

	getUserByIDQuery = `
		SELECT id, username, name, phone_number, user_type, created_at, updated_at
		FROM "user"
		WHERE id = $1`
)

type UserRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUserByPhone(phone string) (*models.User, error) {
	const op = "UserRepository.GetUserByPhone"
	var user models.User
	err := r.db.QueryRow(getUserByPhoneQuery, phone).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &user.AccountType, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("user not found")
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	const op = "UserRepository.GetUserByUsername"
	var user models.User
	err := r.db.QueryRow(getUserByUsernameQuery, username).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &user.AccountType, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("user not found")
			return nil, err
		}
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	const op = "UserRepository.GetUserByID"
	var user models.User
	err := r.db.QueryRow(getUserByIDQuery, id).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.AccountType, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("user not found")
			return nil, err
		}
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUsersNames(usersIds []uuid.UUID) ([]string, error) {
	const op = "UserRepository.GetUsersNames"

	if len(usersIds) == 0 {
		return []string{}, nil
	}

	query := `SELECT name FROM "user" WHERE id IN (`
	placeholders := []string{}
	args := make([]interface{}, len(usersIds))

	for i, userID := range usersIds {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args[i] = userID
	}

	query += strings.Join(placeholders, ", ")
	query += ")"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0, len(usersIds))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, err
		}
		result = append(result, name)
	}

	return result, nil
}
