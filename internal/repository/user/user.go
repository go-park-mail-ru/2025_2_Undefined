package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	getUserByPhoneQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.phone_number = $1`

	getUserByUsernameQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.username = $1`

	getUserByIDQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.id = $1`

	updateUserQuery = `
		UPDATE "user" 
		SET name = $2, description = $3
		WHERE id = $1
		RETURNING updated_at`

	getUsersNamesQuery = `
		SELECT name FROM "user" WHERE id = ANY($1)`

	insertUserAvatarInAttachmentTableQuery = `
		INSERT INTO attachment (id, file_name, file_size, content_disposition)
		VALUES ($1, $2, $3, $4)`

	insertUserAvatarInUserAvatarTableQuery = `
		INSERT INTO avatar_user (user_id, attachment_id)
		VALUES ($1, $2)`
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	const op = "UserRepository.GetUserByPhone"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("phone", phone)
	logger.Debug("Starting database operation: get user by phone")

	var user models.User
	err := r.pool.QueryRow(ctx, getUserByPhoneQuery, phone).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &user.AccountType, &user.AvatarID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("Database operation completed: user not found")
			return nil, errs.ErrUserNotFound
		}
		logger.WithError(err).Error("Database operation failed: get user by phone query")
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}

	logger.WithField("user_id", user.ID.String()).Info("Database operation completed successfully: user found by phone")
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const op = "UserRepository.GetUserByUsername"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("username", username)
	logger.Debug("Starting database operation: get user by username")

	var user models.User
	err := r.pool.QueryRow(ctx, getUserByUsernameQuery, username).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &user.AccountType, &user.AvatarID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("Database operation completed: user not found")
			return nil, errs.ErrUserNotFound
		}
		logger.WithError(err).Error("Database operation failed: get user by username query")
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	logger.WithField("user_id", user.ID.String()).Info("Database operation completed successfully: user found by username")
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "UserRepository.GetUserByID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", id.String())
	logger.Debug("Starting database operation: get user by ID")

	var user models.User
	err := r.pool.QueryRow(ctx, getUserByIDQuery, id).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.AccountType, &user.AvatarID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("Database operation completed: user not found")
			return nil, errs.ErrUserNotFound
		}
		logger.WithError(err).Error("Database operation failed: get user by ID query")
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	logger.Info("Database operation completed successfully: user found by ID")
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, name, description string) error {
	const op = "UserRepository.UpdateUser"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: update user")

	var updatedAt interface{}
	err := r.pool.QueryRow(ctx, updateUserQuery, userID, name, description).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("Database operation completed: user not found")
			return errs.ErrUserNotFound
		}
		logger.WithError(err).Error("Database operation failed: update user")
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("Database operation completed successfully: user updated")
	return nil
}

func (r *UserRepository) GetUsersNames(ctx context.Context, usersIds []uuid.UUID) ([]string, error) {
	const op = "UserRepository.GetUsersNames"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("users_count", len(usersIds))
	logger.Debug("Starting database operation: get users names by IDs")

	if len(usersIds) == 0 {
		logger.Debug("Database operation completed: empty users list provided")
		return []string{}, nil
	}

	rows, err := r.pool.Query(ctx, getUsersNamesQuery, usersIds)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get users names query")
		return nil, fmt.Errorf("failed to get users names: %w", err)
	}
	defer rows.Close()

	result := make([]string, 0, len(usersIds))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			logger.WithError(err).Error("Database operation failed: scan user name row")
			return nil, fmt.Errorf("failed to scan user name: %w", err)
		}
		result = append(result, name)
	}

	if err := rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	logger.WithField("names_count", len(result)).Info("Database operation completed successfully: users names retrieved")
	return result, nil
}

func (r *UserRepository) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarID uuid.UUID, fileSize int64) error {
	const op = "UserRepository.UpdateUserAvatar"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: update user avatar")

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Вставляем запись в таблицу attachment
	_, err = tx.Exec(ctx, insertUserAvatarInAttachmentTableQuery,
		avatarID, "avatar_"+avatarID.String(), fileSize, "inline")
	if err != nil {
		tx.Rollback(ctx)
		logger.WithError(err).Error("Database operation failed: insert user avatar in attachment table")
		return fmt.Errorf("failed to insert attachment: %w", err)
	}

	// Вставляем связь в таблицу avatar_user
	_, err = tx.Exec(ctx, insertUserAvatarInUserAvatarTableQuery, userID, avatarID)
	if err != nil {
		tx.Rollback(ctx)
		logger.WithError(err).Error("Database operation failed: insert user avatar in user_avatar table")
		return fmt.Errorf("failed to insert user avatar: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		logger.WithError(err).Error("Database operation failed: commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Database operation completed successfully: user avatar updated")
	return nil
}
