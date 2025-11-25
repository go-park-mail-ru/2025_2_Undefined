package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_GetUserByPhone_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	phone := "+79998887766"
	username := "test_user"
	name := "Test User"
	passwordHash := "hashed_password"
	accountType := UserModels.UserAccount
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := pgxmock.NewRows([]string{"id", "username", "name", "phone_number", "password_hash", "description", "user_type", "created_at", "updated_at"}).
		AddRow(userID, username, name, phone, passwordHash, nil, accountType, createdAt, updatedAt)

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.phone_number = $1`).
		WithArgs(phone).
		WillReturnRows(rows)

	user, err := repo.GetUserByPhone(ctx, phone)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, phone, user.PhoneNumber)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.Equal(t, accountType, user.AccountType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByPhone_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	phone := "+79998887766"

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.phone_number = $1`).
		WithArgs(phone).
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetUserByPhone(ctx, phone)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByPhone_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	phone := "+79998887766"

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.phone_number = $1`).
		WithArgs(phone).
		WillReturnError(fmt.Errorf("connection error"))

	user, err := repo.GetUserByPhone(ctx, phone)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, fmt.Errorf("connection error"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByUsername_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	phone := "+79998887766"
	username := "test_user"
	name := "Test User"
	passwordHash := "hashed_password"
	accountType := UserModels.UserAccount
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := pgxmock.NewRows([]string{"id", "username", "name", "phone_number", "password_hash", "description", "user_type", "created_at", "updated_at"}).
		AddRow(userID, username, name, phone, passwordHash, nil, accountType, createdAt, updatedAt)

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.username = $1`).
		WithArgs(username).
		WillReturnRows(rows)

	user, err := repo.GetUserByUsername(ctx, username)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, phone, user.PhoneNumber)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.Equal(t, accountType, user.AccountType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByUsername_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	username := "nonexistent_user"

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.username = $1`).
		WithArgs(username).
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByUsername_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	username := "test_user"

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.username = $1`).
		WithArgs(username).
		WillReturnError(fmt.Errorf("connection error"))

	user, err := repo.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, fmt.Errorf("connection error"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	phone := "+79998887766"
	username := "test_user"
	name := "Test User"
	accountType := UserModels.UserAccount
	createdAt := time.Now()
	updatedAt := time.Now()

	passwordHash := "hashed_password"

	rows := pgxmock.NewRows([]string{"id", "username", "name", "phone_number", "password_hash", "description", "user_type", "created_at", "updated_at"}).
		AddRow(userID, username, name, phone, passwordHash, nil, accountType, createdAt, updatedAt)

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.id = $1`).
		WithArgs(userID).
		WillReturnRows(rows)

	user, err := repo.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, phone, user.PhoneNumber)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, accountType, user.AccountType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.id = $1`).
		WithArgs(userID).
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, errs.ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.description, u.user_type, 
               u.created_at, u.updated_at
        FROM "user" u
        WHERE u.id = $1`).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("connection error"))

	user, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, fmt.Errorf("connection error"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsersNames_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	names := []string{"User1", "User2", "User3"}

	rows := pgxmock.NewRows([]string{"name"}).
		AddRow(names[0]).
		AddRow(names[1]).
		AddRow(names[2])

	mock.ExpectQuery(`SELECT name FROM "user" WHERE id IN ($1, $2, $3)`).
		WithArgs(userIDs[0], userIDs[1], userIDs[2]).
		WillReturnRows(rows)

	result, err := repo.GetUsersNames(ctx, userIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, names, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsersNames_EmptyList(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	result, err := repo.GetUsersNames(ctx, []uuid.UUID{})

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestUserRepository_GetUsersNames_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userIDs := []uuid.UUID{uuid.New()}

	mock.ExpectQuery(`SELECT name FROM "user" WHERE id IN ($1)`).
		WithArgs(userIDs[0]).
		WillReturnError(fmt.Errorf("connection error"))

	result, err := repo.GetUsersNames(ctx, userIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, fmt.Errorf("connection error"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsersNames_ScanError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userIDs := []uuid.UUID{uuid.New()}

	rows := pgxmock.NewRows([]string{"name"}).
		AddRow("test_name").
		RowError(0, fmt.Errorf("scan error"))

	mock.ExpectQuery(`SELECT name FROM "user" WHERE id IN ($1)`).
		WithArgs(userIDs[0]).
		WillReturnRows(rows)

	result, err := repo.GetUsersNames(ctx, userIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateUserInfo_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	name := "Updated Name"
	username := "updated_user"
	bio := "Updated bio"

	mock.ExpectExec(`UPDATE "user" SET name = \$1, username = \$2, description = \$3 WHERE id = \$4`).
		WithArgs(name, username, bio, userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateUserInfo(ctx, userID, &name, &username, &bio)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateUserInfo_NoFieldsToUpdate(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()

	err = repo.UpdateUserInfo(ctx, userID, nil, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "no fields to update", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateUserInfo_DuplicateKey(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	username := "existing_user"

	pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate key value"}
	mock.ExpectExec(`UPDATE "user" SET username = \$1 WHERE id = \$2`).
		WithArgs(username, userID).
		WillReturnError(pgErr)

	err = repo.UpdateUserInfo(ctx, userID, nil, &username, nil)

	assert.Error(t, err)
	assert.Equal(t, errs.ErrIsDuplicateKey, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateUserInfo_UserNotUpdated(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	name := "New Name"

	mock.ExpectExec(`UPDATE "user" SET name = \$1 WHERE id = \$2`).
		WithArgs(name, userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = repo.UpdateUserInfo(ctx, userID, &name, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "user not updated", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserAvatars_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	avatarIDs := []uuid.UUID{uuid.New(), uuid.New()}

	rows := pgxmock.NewRows([]string{"user_id", "attachment_id"}).
		AddRow(userIDs[0], avatarIDs[0]).
		AddRow(userIDs[1], avatarIDs[1])

	mock.ExpectQuery(`WITH latest_avatars AS`).
		WithArgs(userIDs).
		WillReturnRows(rows)

	result, err := repo.GetUserAvatars(ctx, userIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, avatarIDs[0], result[userIDs[0].String()])
	assert.Equal(t, avatarIDs[1], result[userIDs[1].String()])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserAvatars_EmptyList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	result, err := repo.GetUserAvatars(ctx, []uuid.UUID{})

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
