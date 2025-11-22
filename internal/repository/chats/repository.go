package repository

import "github.com/jackc/pgx/v5/pgxpool"

type ChatsRepository struct {
	db *pgxpool.Pool
}

func NewChatsRepository(db *pgxpool.Pool) *ChatsRepository {
	return &ChatsRepository{
		db: db,
	}
}
