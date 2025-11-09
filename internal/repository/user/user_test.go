package repository
/*
import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_GetUserByPhone_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userID := uuid.New()
	phone := "+79998887766"
	username := "test_user"
	name := "Test User"
	passwordHash := "hashed_password"
	accountType := UserModels.UserAccount
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "username", "name", "phone_number", "password_hash", "user_type", "attachment_id", "created_at", "updated_at"}).
		AddRow(userID, username, name, phone, passwordHash, accountType, nil, createdAt, updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.phone_number = $1`)).
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
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	phone := "+79998887766"

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.phone_number = $1`)).
		WithArgs(phone).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByPhone(ctx, phone)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByPhone_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	phone := "+79998887766"

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.phone_number = $1`)).
		WithArgs(phone).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.GetUserByPhone(ctx, phone)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByUsername_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userID := uuid.New()
	phone := "+79998887766"
	username := "test_user"
	name := "Test User"
	passwordHash := "hashed_password"
	accountType := UserModels.UserAccount
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "username", "name", "phone_number", "password_hash", "user_type", "attachment_id", "created_at", "updated_at"}).
		AddRow(userID, username, name, phone, passwordHash, accountType, nil, createdAt, updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.username = $1`)).
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
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	username := "nonexistent_user"

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.username = $1`)).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByUsername_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	username := "test_user"

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.password_hash, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.username = $1`)).
		WithArgs(username).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userID := uuid.New()
	phone := "+79998887766"
	username := "test_user"
	name := "Test User"
	accountType := UserModels.UserAccount
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "username", "name", "phone_number", "user_type", "attachment_id", "created_at", "updated_at"}).
		AddRow(userID, username, name, phone, accountType, nil, createdAt, updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.id = $1`)).
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
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.id = $1`)).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, errs.ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT u.id, u.username, u.name, u.phone_number, u.user_type, 
               latest_avatar.attachment_id, u.created_at, u.updated_at
        FROM "user" u
        LEFT JOIN (
            SELECT DISTINCT ON (user_id) user_id, attachment_id
            FROM avatar_user
            ORDER BY user_id, updated_at DESC
        ) latest_avatar ON latest_avatar.user_id = u.id
        WHERE u.id = $1`)).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsersNames_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	names := []string{"User1", "User2", "User3"}

	rows := sqlmock.NewRows([]string{"name"}).
		AddRow(names[0]).
		AddRow(names[1]).
		AddRow(names[2])

	mock.ExpectQuery(`SELECT name FROM "user" WHERE id IN \(\$1, \$2, \$3\)`).
		WithArgs(userIDs[0], userIDs[1], userIDs[2]).
		WillReturnRows(rows)

	result, err := repo.GetUsersNames(ctx, userIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, names, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsersNames_EmptyList(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	result, err := repo.GetUsersNames(ctx, []uuid.UUID{})

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestUserRepository_GetUsersNames_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userIDs := []uuid.UUID{uuid.New()}

	mock.ExpectQuery(`SELECT name FROM "user" WHERE id IN \(\$1\)`).
		WithArgs(userIDs[0]).
		WillReturnError(sql.ErrConnDone)

	result, err := repo.GetUsersNames(ctx, userIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsersNames_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	userIDs := []uuid.UUID{uuid.New()}

	rows := sqlmock.NewRows([]string{"name"}).
		AddRow(nil)

	mock.ExpectQuery(`SELECT name FROM "user" WHERE id IN \(\$1\)`).
		WithArgs(userIDs[0]).
		WillReturnRows(rows)

	result, err := repo.GetUsersNames(ctx, userIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
*/