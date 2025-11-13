package repository

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	authRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/auth"
	chatsRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/chats"
	contactRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/contact"
	messageRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/message"
	userRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
)

// Repositories содержит все репозитории приложения
type Repositories struct {
	Auth    *authRepo.AuthRepository
	User    *userRepo.UserRepository
	Chats   *chatsRepo.ChatsRepository
	Message *messageRepo.MessageRepository
	Contact *contactRepo.ContactRepository
	DB      *DBConnector
}

// NewRepositories создает новые экземпляры всех репозиториев
func NewRepositories(cfg *Config) (*Repositories, error) {
	// Создаем подключение к базе данных
	dbConnector, err := NewDBConnector(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	pool := dbConnector.GetPool()

	// Инициализируем все репозитории
	repos := &Repositories{
		Auth:    authRepo.New(pool),
		User:    userRepo.New(pool),
		Chats:   chatsRepo.NewChatsRepository(pool),
		Message: messageRepo.NewMessageRepository(pool),
		Contact: contactRepo.New(pool),
		DB:      dbConnector,
	}

	logrus.Info("All repositories initialized successfully")
	return repos, nil
}

// Close закрывает все соединения с базой данных
func (r *Repositories) Close() {
	if r.DB != nil {
		r.DB.Close()
	}
}

// HealthCheck проверяет состояние подключения к базе данных
func (r *Repositories) HealthCheck(ctx context.Context) error {
	return r.DB.GetPool().Ping(ctx)
}
