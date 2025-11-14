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
	const query = "SELECT user by phone"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("phone", phone)

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var user models.User
	var bio sql.NullString
	var avatarID sql.NullString
	err := r.db.QueryRow(getUserByPhoneQuery, phone).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &avatarID, &user.CreatedAt, &user.UpdatedAt)

	if bio.Valid {
		user.Bio = &bio.String
	}

	if avatarID.Valid {
		if parsedUUID, parseErr := uuid.Parse(avatarID.String); parseErr == nil {
			user.AvatarID = &parsedUUID
		}
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			queryStatus = "not found"
			logger.Debugf("db query: %s: user not found: status: %s", query, queryStatus)
			return nil, errors.New("user not found")
		}
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const op = "UserRepository.GetUserByUsername"
	const query = "SELECT user by username"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("username", username)

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var user models.User
	var bio sql.NullString
	var avatarID sql.NullString
	err := r.db.QueryRow(getUserByUsernameQuery, username).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &avatarID, &user.CreatedAt, &user.UpdatedAt)

	if bio.Valid {
		user.Bio = &bio.String
	}

	if avatarID.Valid {
		if parsedUUID, parseErr := uuid.Parse(avatarID.String); parseErr == nil {
			user.AvatarID = &parsedUUID
		}
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			queryStatus = "not found"
			logger.Debugf("db query: %s: user not found: status: %s", query, queryStatus)
			return nil, errors.New("user not found")
		}
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "UserRepository.GetUserByID"
	const query = "SELECT user by ID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", id.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	var user models.User
	var bio sql.NullString
	var avatarID sql.NullString
	err := r.db.QueryRow(getUserByIDQuery, id).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &avatarID, &user.CreatedAt, &user.UpdatedAt)

	if bio.Valid {
		user.Bio = &bio.String
	}

	if avatarID.Valid {
		if parsedUUID, parseErr := uuid.Parse(avatarID.String); parseErr == nil {
			user.AvatarID = &parsedUUID
		}
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			queryStatus = "not found"
			logger.Debugf("db query: %s: user not found: status: %s", query, queryStatus)
			return nil, errs.ErrUserNotFound
		}
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUsersNames(ctx context.Context, usersIds []uuid.UUID) ([]string, error) {
	const op = "UserRepository.GetUsersNames"
	const query = "SELECT users names"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("users_count", len(usersIds))

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	if len(usersIds) == 0 {
		logger.Debugf("db query: %s: empty list: status: %s", query, queryStatus)
		return []string{}, nil
	}

	querySQL := `SELECT name FROM "user" WHERE id IN (`
	placeholders := []string{}
	args := make([]interface{}, len(usersIds))

	for i, userID := range usersIds {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args[i] = userID
	}

	querySQL += strings.Join(placeholders, ", ")
	querySQL += ")"

	rows, err := r.db.Query(querySQL, args...)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0, len(usersIds))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}
		result = append(result, name)
	}

	return result, nil
}

func (r *UserRepository) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarID uuid.UUID, file_size int64) error {
	const op = "UserRepository.UpdateUserAvatar"
	const query = "UPDATE user avatar"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: begin transaction: status: %s", query, queryStatus)
		return err
	}

	_, err = tx.Exec(insertUserAvatarInAttachmentTableQuery, avatarID, "avatar_"+avatarID.String(), file_size, "inline")
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert attachment: status: %s", query, queryStatus)
		return err
	}

	_, err = tx.Exec(insertUserAvatarInUserAvatarTableQuery, userID, avatarID)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert user avatar: status: %s", query, queryStatus)
		return err
	}

	if err := tx.Commit(); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: commit transaction: status: %s", query, queryStatus)
		return err
	}

	return nil
}

func (r *UserRepository) UpdateUserInfo(ctx context.Context, userID uuid.UUID, name *string, username *string, bio *string) error {
	const op = "UserRepository.UpdateUserInfo"
	const query = "UPDATE user info"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

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
		queryStatus = "fail"
		logger.Errorf("db query: %s: no fields to update: status: %s", query, queryStatus)
		return fmt.Errorf("no fields to update")
	}

	args = append(args, userID)

	querySQL := fmt.Sprintf("UPDATE \"user\" SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.Exec(querySQL, args...)
	if err != nil {
		queryStatus = "fail"

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errs.PostgresErrorUniqueViolationCode {
			logger.WithError(err).Errorf("db query: %s: duplicate key violation: status: %s", query, queryStatus)
			return errs.ErrIsDuplicateKey
		}

		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return errors.New("error update user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: check rows affected: status: %s", query, queryStatus)
		return err
	}

	if rowsAffected == 0 {
		queryStatus = "fail"
		logger.Errorf("db query: %s: user not updated: status: %s", query, queryStatus)
		return errors.New("user not updated")
	}

	return nil
}
