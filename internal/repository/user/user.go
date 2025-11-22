package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	getUserByPhoneQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.phone_number = $1`

	getUserByUsernameQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.username = $1`

	getUserByIDQuery = `
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.id = $1`

	insertUserAvatarInAttachmentTableQuery = `
		INSERT INTO attachment (id, file_name, file_size, content_disposition)
		VALUES ($1, $2, $3, $4)`

	insertUserAvatarInUserAvatarTableQuery = `
		INSERT INTO avatar_user (user_id, attachment_id)
		VALUES ($1, $2)`

	getUserAvatarsQuery = `
		WITH latest_avatars AS (
			SELECT DISTINCT ON (au.user_id) 
				au.user_id, 
				a.id as attachment_id
			FROM avatar_user au
			JOIN attachment a ON au.attachment_id = a.id
			WHERE au.user_id = ANY($1)
			ORDER BY au.user_id, au.created_at DESC
		)
		SELECT user_id, attachment_id 
		FROM latest_avatars`
)

type UserRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *UserRepository {
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
	var bio *string
	err := r.db.QueryRow(ctx, getUserByPhoneQuery, phone).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &user.CreatedAt, &user.UpdatedAt)

	if bio != nil {
		user.Bio = bio
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
	var bio *string
	err := r.db.QueryRow(ctx, getUserByUsernameQuery, username).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &user.CreatedAt, &user.UpdatedAt)

	if bio != nil {
		user.Bio = bio
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
	var bio *string
	err := r.db.QueryRow(ctx, getUserByIDQuery, id).
		Scan(&user.ID, &user.Username, &user.Name, &user.PhoneNumber, &user.PasswordHash, &bio, &user.AccountType, &user.CreatedAt, &user.UpdatedAt)

	if bio != nil {
		user.Bio = bio
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

	rows, err := r.db.Query(ctx, querySQL, args...)
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

	tx, err := r.db.Begin(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: begin transaction: status: %s", query, queryStatus)
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, insertUserAvatarInAttachmentTableQuery, avatarID, "avatar_"+avatarID.String(), file_size, "inline")
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert attachment: status: %s", query, queryStatus)
		return err
	}

	_, err = tx.Exec(ctx, insertUserAvatarInUserAvatarTableQuery, userID, avatarID)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: insert user avatar: status: %s", query, queryStatus)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
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

	result, err := r.db.Exec(ctx, querySQL, args...)
	if err != nil {
		queryStatus = "fail"

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == errs.PostgresErrorUniqueViolationCode {
			logger.WithError(err).Errorf("db query: %s: duplicate key violation: status: %s", query, queryStatus)
			return errs.ErrIsDuplicateKey
		}

		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return errors.New("error update user")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		queryStatus = "fail"
		logger.Errorf("db query: %s: user not updated: status: %s", query, queryStatus)
		return errors.New("user not updated")
	}

	return nil
}

func (r *UserRepository) GetUserAvatars(ctx context.Context, userIDs []uuid.UUID) (map[string]uuid.UUID, error) {
	const op = "UserRepository.GetUserAvatars"
	const query = "SELECT user avatars"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_ids_count", len(userIDs))

	queryStatus := "success"
	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	if len(userIDs) == 0 {
		logger.Debugf("db query: %s: empty list: status: %s", query, queryStatus)
		return make(map[string]uuid.UUID), nil
	}

	// pgx нативно поддерживает работу с массивами PostgreSQL
	rows, err := r.db.Query(ctx, getUserAvatarsQuery, userIDs)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]uuid.UUID)
	for rows.Next() {
		var userID, avatarID uuid.UUID
		if err := rows.Scan(&userID, &avatarID); err != nil {
			queryStatus = "fail"
			logger.WithError(err).Errorf("db query: %s: scan row error: status: %s", query, queryStatus)
			return nil, err
		}
		result[userID.String()] = avatarID
	}

	if err := rows.Err(); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: rows iteration error: status: %s", query, queryStatus)
		return nil, err
	}

	logger.WithField("avatars_count", len(result)).Debugf("db query: %s: status: %s", query, queryStatus)
	return result, nil
}
