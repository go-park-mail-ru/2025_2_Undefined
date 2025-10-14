package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
