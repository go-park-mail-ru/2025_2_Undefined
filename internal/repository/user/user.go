package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	getUserByPhoneQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.phone_number = $1`

	getUserByUsernameQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.username = $1`

	getUserByIDQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.id = $1`

	insertUserAvatarInAttachmentTableQuery = `
		INSERT INTO attachment (id, file_name, file_size, content_disposition)
		VALUES ($1, $2, $3, $4)`

	insertUserAvatarInUserAvatarTableQuery = `
		INSERT INTO avatar_user (user_id, attachment_id)
		VALUES ($1, $2)`
)

type UserRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	const op = "UserRepository.GetUserByPhone"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("phone", phone)
	logger.Debug("Starting database operation: get user by phone")

	var user models.User
	var bio sql.NullString
	err := r.db.QueryRow(getUserByPhoneQuery, phone).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &user.AvatarID, &user.CreatedAt, &user.UpdatedAt)

	if bio.Valid {
		user.Bio = &bio.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("Database operation completed: user not found")
			err = errors.New("user not found")
			return nil, err
		}
		logger.WithError(err).Error("Database operation failed: get user by phone query")
		return nil, err
	}

	logger.WithField("user_id", user.ID.String()).Info("Database operation completed successfully: user found by phone")
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const op = "UserRepository.GetUserByUsername"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("username", username)
	logger.Debug("Starting database operation: get user by username")

	var user models.User
	var bio sql.NullString
	err := r.db.QueryRow(getUserByUsernameQuery, username).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &user.AvatarID, &user.CreatedAt, &user.UpdatedAt)

	if bio.Valid {
		user.Bio = &bio.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("Database operation completed: user not found")
			err = errors.New("user not found")
			return nil, err
		}
		logger.WithError(err).Error("Database operation failed: get user by username query")
		return nil, err
	}

	logger.WithField("user_id", user.ID.String()).Info("Database operation completed successfully: user found by username")
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "UserRepository.GetUserByID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", id.String())
	logger.Debug("Starting database operation: get user by ID")

	var user models.User
	var bio sql.NullString
	err := r.db.QueryRow(getUserByIDQuery, id).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &user.AvatarID, &user.CreatedAt, &user.UpdatedAt)

	if bio.Valid {
		user.Bio = &bio.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("Database operation completed: user not found")
			err = errs.ErrUserNotFound
			return nil, err
		}
		logger.WithError(err).Error("Database operation failed: get user by ID query")
		return nil, err
	}

	logger.Info("Database operation completed successfully: user found by ID")
	return &user, nil
}

func (r *UserRepository) GetUsersNames(ctx context.Context, usersIds []uuid.UUID) ([]string, error) {
	const op = "UserRepository.GetUsersNames"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("users_count", len(usersIds))
	logger.Debug("Starting database operation: get users names by IDs")

	if len(usersIds) == 0 {
		logger.Debug("Database operation completed: empty users list provided")
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
		logger.WithError(err).Error("Database operation failed: get users names query")
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0, len(usersIds))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			logger.WithError(err).Error("Database operation failed: scan user name row")
			return nil, err
		}
		result = append(result, name)
	}

	logger.WithField("names_count", len(result)).Info("Database operation completed successfully: users names retrieved")
	return result, nil
}

func (r *UserRepository) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarID uuid.UUID, file_size int64) error {
	const op = "UserRepository.UpdateUserAvatar"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: update user avatar")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: begin transaction")
		return err
	}

	_, err = tx.Exec(insertUserAvatarInAttachmentTableQuery, avatarID, "avatar_"+avatarID.String(), file_size, "inline")
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert user avatar in attachment table")
		return err
	}

	_, err = tx.Exec(insertUserAvatarInUserAvatarTableQuery, userID, avatarID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: insert user avatar in user_avatar table")
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.WithError(err).Error("Database operation failed: commit transaction")
		return err
	}

	logger.Info("Database operation completed successfully: user avatar updated")
	return nil
}

func (r *UserRepository) UpdateUserInfo(ctx context.Context, userID uuid.UUID, name *string, username *string, bio *string) error {
	const op = "UserRepository.UpdateUserInfo"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: update user info")

	var setParts []string
	var args []interface{}
	argIndex := 1

	if name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *name)
		argIndex++
	}

	if username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *username)
		argIndex++
	}

	if bio != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *bio)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	args = append(args, userID)

	query := fmt.Sprintf("UPDATE \"user\" SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update user info")

		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "user_username_key" {
					return errs.ErrIsDuplicateKey
				}
			}
		}

		return errors.New("error update user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: check rows affected")
		return err
	}

	if rowsAffected == 0 {
		err := errors.New("user not updated")
		logger.WithError(err).Error("Database operation failed: user not updated")
		return err
	}

	logger.Info("Database operation completed successfully: user into update")
	return nil
}
