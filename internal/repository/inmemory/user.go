package inmemory

import (
	"errors"
	"sync"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
)

type UserRepo struct {
	users map[string]*models.User //храним по phone
	mutex sync.RWMutex
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[string]*models.User),
	}
}

func (r *UserRepo) Create(user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.PhoneNumber]; exists {
		return errors.New("user with this phone already exists")
	}

	r.users[user.PhoneNumber] = user
	return nil
}

func (r *UserRepo) GetByID(id uuid.UUID) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *UserRepo) GetByPhone(phone string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[phone]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepo) GetByUsername(username string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *UserRepo) Update(updatedUser *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, exists := r.users[updatedUser.PhoneNumber]
	if !exists {
		return errors.New("user not found")
	}

	r.users[updatedUser.PhoneNumber] = updatedUser
	return nil
}
