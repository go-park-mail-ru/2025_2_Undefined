package repository

import (
	"sync"
	"time"
)

type TokenRepository interface {
	AddToBlacklist(token string) error
	IsInBlacklist(token string) bool
	CleanupExpiredTokens()
}

type TokenRepo struct {
	blacklist map[string]time.Time
	mutex     sync.RWMutex
	timeDie   time.Duration
}

func NewTokenRepo() *TokenRepo {
	repo := &TokenRepo{
		blacklist: make(map[string]time.Time),
		timeDie:   24 * time.Hour,
	}
	// запуск горутины для очистки просроченных токенов
	go repo.startCleanupRoutine()
	return repo
}

func (r *TokenRepo) AddToBlacklist(token string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.blacklist[token] = time.Now()
	return nil
}

func (r *TokenRepo) IsInBlacklist(token string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.blacklist[token]
	return exists
}

func (r *TokenRepo) CleanupExpiredTokens() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()

	for token, addedAt := range r.blacklist {
		if now.Sub(addedAt) > r.timeDie {
			delete(r.blacklist, token)
		}
	}
}

func (r *TokenRepo) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		r.CleanupExpiredTokens()
	}
}
