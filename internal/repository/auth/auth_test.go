package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRepository_CreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_5").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	userID := uuid.New()
	mock.ExpectQuery(`INSERT INTO "user" \(id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at\)`).
		WithArgs(sqlmock.AnyArg(), "user_5", "Test User", "+79998887766", "hashed_password", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "phone_number", "user_type_id"}).
			AddRow(userID, "user_5", "+79998887766", 0))

	user, err := repo.CreateUser("Test User", "+79998887766", "hashed_password")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user_5", user.Username)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "+79998887766", user.PhoneNumber)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, 0, user.AccountType)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_CreateUser_DuplicateKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_0").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	pqErr := &pq.Error{Code: "23505"}
	mock.ExpectQuery(`INSERT INTO "user" \(id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at\)`).
		WithArgs(sqlmock.AnyArg(), "user_0", "Test User", "+79998887766", "hashed_password", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(pqErr)

	user, err := repo.CreateUser("Test User", "+79998887766", "hashed_password")

	assert.Error(t, err)
	assert.Equal(t, errs.ErrIsDuplicateKey, err)
	assert.Nil(t, user)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_CreateUser_GenerateUsernameError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnError(errors.New("database error"))

	user, err := repo.CreateUser("Test User", "+79998887766", "hashed_password")

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, user)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_CreateUser_UsernameExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_3").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	userID := uuid.New()
	mock.ExpectQuery(`INSERT INTO "user" \(id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at\)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Test User", "+79998887766", "hashed_password", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "phone_number", "user_type_id"}).
			AddRow(userID, "user_12345678", "+79998887766", 0))

	user, err := repo.CreateUser("Test User", "+79998887766", "hashed_password")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Contains(t, user.Username, "user_")
	assert.Equal(t, "Test User", user.Name)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_GetUserByPhone_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery(`SELECT id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at FROM "user" WHERE phone_number = \$1`).
		WithArgs("+79998887766").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "name", "phone_number", "password_hash", "user_type_id", "created_at", "updated_at"}).
			AddRow(userID, "testuser", "Test User", "+79998887766", "hashed_password", 0, createdAt, updatedAt))

	user, err := repo.GetUserByPhone("+79998887766")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "+79998887766", user.PhoneNumber)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, 0, user.AccountType)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_GetUserByPhone_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at FROM "user" WHERE phone_number = \$1`).
		WithArgs("+79998887766").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByPhone("+79998887766")

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, user)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_GetUserByPhone_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at FROM "user" WHERE phone_number = \$1`).
		WithArgs("+79998887766").
		WillReturnError(errors.New("database connection error"))

	user, err := repo.GetUserByPhone("+79998887766")

	assert.Error(t, err)
	assert.Equal(t, "database connection error", err.Error())
	assert.Nil(t, user)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_GetUserByUsername_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery(`SELECT id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at FROM "user" WHERE username = \$1`).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "name", "phone_number", "password_hash", "user_type_id", "created_at", "updated_at"}).
			AddRow(userID, "testuser", "Test User", "+79998887766", "hashed_password", 0, createdAt, updatedAt))

	user, err := repo.GetUserByUsername("testuser")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "+79998887766", user.PhoneNumber)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, 0, user.AccountType)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_GetUserByUsername_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at FROM "user" WHERE username = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByUsername("nonexistent")

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, user)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_GetUserByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	mock.ExpectQuery(`SELECT id, username, name, phone_number, user_type_id, created_at, updated_at FROM "user" WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "name", "phone_number", "user_type_id", "created_at", "updated_at"}).
			AddRow(userID, "testuser", "Test User", "+79998887766", 0, createdAt, updatedAt))

	user, err := repo.GetUserByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "+79998887766", user.PhoneNumber)
	assert.Equal(t, 0, user.AccountType)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_GetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, username, name, phone_number, user_type_id, created_at, updated_at FROM "user" WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByID(userID)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, user)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_checkUsernameExists_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("existinguser").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.checkUsernameExists("existinguser")

	assert.NoError(t, err)
	assert.True(t, exists)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_checkUsernameExists_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("newuser").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.checkUsernameExists("newuser")

	assert.NoError(t, err)
	assert.False(t, exists)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_checkUsernameExists_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("testuser").
		WillReturnError(errors.New("database error"))

	exists, err := repo.checkUsernameExists("testuser")

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.False(t, exists)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_generateUsername_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_10").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	username, err := repo.generateUsername()

	assert.NoError(t, err)
	assert.Equal(t, "user_10", username)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_generateUsername_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnError(errors.New("count error"))

	username, err := repo.generateUsername()

	assert.Error(t, err)
	assert.Equal(t, "count error", err.Error())
	assert.Empty(t, username)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_generateUsername_UsernameExistsCheckError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_5").
		WillReturnError(errors.New("check error"))

	username, err := repo.generateUsername()

	assert.Error(t, err)
	assert.Equal(t, "check error", err.Error())
	assert.Empty(t, username)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_generateUsername_FallbackToUUID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_7").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	username, err := repo.generateUsername()

	assert.NoError(t, err)
	assert.Contains(t, username, "user_")
	assert.True(t, len(username) > len("user_7"))

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepository_CreateUser_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO "user" \(id, username, name, phone_number, password_hash, user_type_id, created_at, updated_at\)`).
		WithArgs(sqlmock.AnyArg(), "user_1", "Test User", "+79998887766", "hashed_password", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("general database error"))

	user, err := repo.CreateUser("Test User", "+79998887766", "hashed_password")

	assert.Error(t, err)
	assert.Equal(t, "general database error", err.Error())
	assert.Nil(t, user)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
