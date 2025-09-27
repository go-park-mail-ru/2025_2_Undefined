package inmemory

import (
	"fmt"
	"testing"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Success(t *testing.T) {
	repo := NewUserRepo()
	user := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "testuser",
		Email:       "test@mail.ru",
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.Equal(t, user, repo.users[user.PhoneNumber])
}

func TestCreateUser_Error_UserAlreadyExists(t *testing.T) {
	repo := NewUserRepo()
	user := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "testuser",
		Email:       "test@mail.ru",
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	duplicateUser := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "anotheruser",
		Email:       "another@mail.ru",
	}

	err = repo.Create(duplicateUser)
	assert.Error(t, err)
	assert.Equal(t, "user with this phone already exists", err.Error())
}

func TestGetByID_Success(t *testing.T) {
	repo := NewUserRepo()
	user := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "testuser",
		Email:       "test@mail.ru",
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	foundUser, err := repo.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user, foundUser)
}

func TestGetByID_Error_UserNotFound(t *testing.T) {
	repo := NewUserRepo()
	nonExistentID := uuid.New()

	user, err := repo.GetByID(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, user)
}

func TestGetByPhone_Success(t *testing.T) {
	repo := NewUserRepo()
	user := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "testuser",
		Email:       "test@mail.ru",
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	foundUser, err := repo.GetByPhone(user.PhoneNumber)
	assert.NoError(t, err)
	assert.Equal(t, user, foundUser)
}

func TestGetByPhone_Error_UserNotFound(t *testing.T) {
	repo := NewUserRepo()

	user, err := repo.GetByPhone("+79990001122")
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, user)
}

func TestGetByUsername_Success(t *testing.T) {
	repo := NewUserRepo()
	user := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "testuser",
		Email:       "test@mail.ru",
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	foundUser, err := repo.GetByUsername(user.Username)
	assert.NoError(t, err)
	assert.Equal(t, user, foundUser)
}

func TestGetByUsername_Error_UserNotFound(t *testing.T) {
	repo := NewUserRepo()

	user, err := repo.GetByUsername("user222")
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, user)
}

func TestGetByEmail_Success(t *testing.T) {
	repo := NewUserRepo()
	user := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "testuser",
		Email:       "test@mail.ru",
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	foundUser, err := repo.GetByEmail(user.Email)
	assert.NoError(t, err)
	assert.Equal(t, user, foundUser)
}

func TestGetByEmail_Error_UserNotFound(t *testing.T) {
	repo := NewUserRepo()

	user, err := repo.GetByEmail("nonexistent@mail.ru")
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, user)
}

func TestUpdate_Success(t *testing.T) {
	repo := NewUserRepo()
	user := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "testuser",
		Email:       "test@mail.ru",
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	updatedUser := &models.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Username:    "updateduser",
		Email:       "updated@mail.ru",
	}

	err = repo.Update(updatedUser)
	assert.NoError(t, err)

	foundUser, err := repo.GetByPhone(user.PhoneNumber)
	assert.NoError(t, err)
	assert.Equal(t, updatedUser.Username, foundUser.Username)
	assert.Equal(t, updatedUser.Email, foundUser.Email)
	assert.Equal(t, user.ID, foundUser.ID)
}

func TestUpdate_Error_UserNotFound(t *testing.T) {
	repo := NewUserRepo()
	nonExistentUser := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79990001122",
		Username:    "nonexistent",
		Email:       "nonexistent@mail.ru",
	}

	err := repo.Update(nonExistentUser)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestGetByID_AfterMultipleCreations(t *testing.T) {
	repo := NewUserRepo()

	users := make([]*models.User, 3)
	for i := 0; i < 3; i++ {
		user := &models.User{
			ID:          uuid.New(),
			PhoneNumber: fmt.Sprintf("+7999888776%d", i),
			Username:    fmt.Sprintf("user%d", i),
			Email:       fmt.Sprintf("user%d@mail.ru", i),
		}
		users[i] = user
		err := repo.Create(user)
		assert.NoError(t, err)
	}

	for _, expectedUser := range users {
		foundUser, err := repo.GetByID(expectedUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, foundUser)
	}
}

func TestUpdate_ChangesAllFields(t *testing.T) {
	repo := NewUserRepo()
	originalUser := &models.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Username:    "original",
		Email:       "original@mail.ru",
	}

	err := repo.Create(originalUser)
	assert.NoError(t, err)

	updatedUser := &models.User{
		ID:          originalUser.ID,
		PhoneNumber: originalUser.PhoneNumber,
		Username:    "updated",
		Email:       "updated@mail.ru",
	}

	err = repo.Update(updatedUser)
	assert.NoError(t, err)

	foundUser, err := repo.GetByPhone(originalUser.PhoneNumber)
	assert.NoError(t, err)
	assert.Equal(t, updatedUser.ID, foundUser.ID)
	assert.Equal(t, updatedUser.PhoneNumber, foundUser.PhoneNumber)
	assert.Equal(t, updatedUser.Username, foundUser.Username)
	assert.Equal(t, updatedUser.Email, foundUser.Email)
}
