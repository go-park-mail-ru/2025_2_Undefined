package repository

import (
	"context"
	"database/sql"
	"fmt"

	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
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

func (r *AuthRepository) CreateUser(ctx context.Context, name string, phone string, password_hash string) (*models.User, error) {
	const op = "AuthRepository.CreateUser"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Starting database operation: create user")

	newUsername, err := r.generateUsername(ctx)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: generate username")
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

	logger.Debug("Executing database query: INSERT user")
	err = r.db.QueryRow(createUserQuery,
		user.ID, user.Username, user.Name, user.PhoneNumber, user.PasswordHash, user.AccountType, user.CreatedAt, user.UpdatedAt).
		Scan(&user.ID, &user.Username, &user.PhoneNumber, &user.AccountType)

	if err != nil {
		// Проверяем является ли ошибка наруением уникального ограничения
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			err = errs.ErrIsDuplicateKey
			logger.WithError(err).Error("Database operation failed: duplicate key constraint violation")
			return nil, err
		}
		logger.WithError(err).Error("Database operation failed: create user query")
		return nil, err
	}

	logger.Info("Database operation completed successfully: user created")
	return user, nil
}

func (r *AuthRepository) generateUsername(ctx context.Context) (string, error) {
	const op = "AuthRepository.generateUsername"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Starting database operation: count users")

	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM "user"`).Scan(&count)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: count users query")
		return "", err
	}

	username := fmt.Sprintf("user_%d", count)

	exists, err := r.checkUsernameExists(ctx, username)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check username exists")
		return "", err
	}
	if !exists {
		logger.Debug("Database operation completed: username generated successfully")
		return username, nil
	}

	//если не получилось создать юзернейм по умолчанию, то создаем через uuid
	logger.Debug("Database operation completed: fallback to UUID-based username")
	return "user_" + uuid.New().String()[:8], nil
}

func (r *AuthRepository) checkUsernameExists(ctx context.Context, username string) (bool, error) {
	logger := domains.GetLogger(ctx).WithField("operation", "AuthRepository.checkUsernameExists")
	logger.Debug("Starting database operation: check username exists")

	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM "user" WHERE username = $1)`, username).Scan(&exists)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check username exists query")
		return false, err
	}

	logger.Debug("Database operation completed: username existence checked")
	return exists, err
}
