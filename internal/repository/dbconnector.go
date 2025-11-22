package repository

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetConnectionString(conf *config.DBConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DB,
	)
}

func NewPgxPool(ctx context.Context, conf *config.DBConfig) (*pgxpool.Pool, error) {
	connString := GetConnectionString(conf)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
