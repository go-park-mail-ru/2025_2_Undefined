package main

import (
	"context"
	"errors"
	"os"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	redisClient "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	const op = "main"
	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.WithError(err).Fatal("Error loading configuration")
	}

	dsn := repository.GetConnectionString(cfg.DBConfig)

	m, err := migrate.New(
		cfg.MigrationsConfig.Path,
		dsn,
	)
	if err != nil {
		logger.WithError(err).Panic("Error initializing migrations")
	}

	redis, err := redisClient.NewClient(cfg.RedisConfig)
	if err != nil {
		logger.WithError(err).Warn("Warning: Could not connect to Redis")
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "down":
			// Откат миграций PostgreSQL
			if err = m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				logger.WithError(err).Fatal("Error rolling back migrations")
			}
			logger.Info("Migrations rolled back successfully.")

			// Очистка Redis при откате
			if redis != nil {
				if err := clearRedis(redis); err != nil {
					logger.WithError(err).Warn("Warning: Could not clear Redis")
				} else {
					logger.Info("Redis cleared successfully.")
				}
			}

		case "clear-redis":
			// Только очистка Redis
			if redis != nil {
				if err := clearRedis(redis); err != nil {
					logger.WithError(err).Fatal("Error clearing Redis")
				} else {
					logger.Info("Redis cleared successfully.")
				}
			} else {
				logger.Fatal("Could not connect to Redis")
			}

		default:
			// Применение миграций PostgreSQL (по умолчанию)
			if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				logger.WithError(err).Fatal("Error applying migrations")
			}
			logger.Info("Migrations applied successfully.")
		}
	} else {
		// Применение миграций PostgreSQL (по умолчанию)
		if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.WithError(err).Fatal("Error applying migrations")
		}
		logger.Info("Migrations applied successfully.")
	}
}

func clearRedis(redis *redisClient.Client) error {
	ctx := context.Background()
	return redis.Client.FlushDB(ctx).Err()
}
