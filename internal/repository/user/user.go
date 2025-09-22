package repository

import models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByPhone(phone string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
}
