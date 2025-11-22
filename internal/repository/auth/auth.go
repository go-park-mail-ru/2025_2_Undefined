package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createUserQuery = `
		INSERT INTO "user" (id, username, name, phone_number, password_hash, user_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::user_type_enum, $7, $8)
		RETURNING id, username, phone_number, user_type`
)

const (
	maxUsernameRetries = 5
	usernamePrefix     = "user_"
	maxUsernameLength  = 20
	uuidPartLength     = 15
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateUser(ctx context.Context, name string, phone string, password_hash string) (*models.User, error) {
	const op = "AuthRepository.CreateUser"
	const query = "INSERT user"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	queryStatus := "success"

	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: begin transaction: status: %s", query, queryStatus)
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	newUsername, err := r.generateUniqueUsername(ctx, tx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: generate username: status: %s", query, queryStatus)
		return nil, fmt.Errorf("generate username: %w", err)
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

	logger.Debugf("executing: %s with username: %s (length: %d)", query, newUsername, len(newUsername))
	err = tx.QueryRow(ctx, createUserQuery,
		user.ID, user.Username, user.Name, user.PhoneNumber, user.PasswordHash, user.AccountType, user.CreatedAt, user.UpdatedAt).
		Scan(&user.ID, &user.Username, &user.PhoneNumber, &user.AccountType)

	if err != nil {
		queryStatus = "fail"

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case errs.PostgresErrorUniqueViolationCode:
				logger.WithError(err).Errorf("db query: %s: duplicate key violation: status: %s", query, queryStatus)
				return nil, errs.ErrIsDuplicateKey
			case errs.PostgresErrorForeignKeyViolationCode:
				logger.WithError(err).Errorf("db query: %s: foreign key violation: status: %s", query, queryStatus)
				return nil, errs.ErrUserNotFound
			}
		}

		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: commit transaction: status: %s", query, queryStatus)
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

func (r *AuthRepository) generateUniqueUsername(ctx context.Context, tx pgx.Tx) (string, error) {
	const op = "AuthRepository.generateUniqueUsername"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	for i := 0; i < maxUsernameRetries; i++ {
		var username string

		username = r.generateTimestampUsername()

		if len(username) > maxUsernameLength {
			username = username[:maxUsernameLength]
		}

		exists, err := r.checkUsernameExists(ctx, tx, username)
		if err != nil {
			return "", fmt.Errorf("check username exists: %w", err)
		}

		if !exists {
			logger.Debugf("generated unique username: %s (length: %d, attempt %d)", username, len(username), i+1)
			return username, nil
		}

		logger.Debugf("username collision detected: %s (attempt %d)", username, i+1)

		time.Sleep(time.Millisecond * time.Duration(i*5))
	}

	return "", fmt.Errorf("failed to generate unique username after %d attempts", maxUsernameRetries)
}

func (r *AuthRepository) generateTimestampUsername() string {
	timestamp := time.Now().UnixNano()
	timestampStr := fmt.Sprintf("%d", timestamp)

	if len(timestampStr) > uuidPartLength {
		timestampStr = timestampStr[len(timestampStr)-uuidPartLength:]
	} else {
		timestampStr = fmt.Sprintf("%015s", timestampStr)
	}

	return usernamePrefix + timestampStr
}

func (r *AuthRepository) checkUsernameExists(ctx context.Context, tx pgx.Tx, username string) (bool, error) {
	const op = "AuthRepository.checkUsernameExists"
	const query = "CHECK username exists"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	queryStatus := "success"

	defer func() {
		logger.Debugf("db query: %s: status: %s", query, queryStatus)
	}()

	var exists bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM "user" WHERE username = $1)`, username).Scan(&exists)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("db query: %s: execution error: status: %s", query, queryStatus)
		return false, err
	}

	return exists, nil
}
