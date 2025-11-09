package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	redisClient "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	dsn, err := repository.GetConnectionString(cfg.DBConfig)
	if err != nil {
		log.Fatalf("Can't connect to database: %v", err)
	}

	m, err := migrate.New(
		cfg.MigrationsConfig.Path,
		dsn,
	)
	if err != nil {
		log.Panicf("Error initializing migrations: %v", err)
	}

	redis, err := redisClient.NewClient(cfg.RedisConfig)
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "down":
			// Откат миграций PostgreSQL
			if err = m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("Error rolling back migrations: %v", err)
			}
			log.Println("Migrations rolled back successfully.")

			// Очистка Redis при откате
			if redis != nil {
				if err := clearRedis(redis); err != nil {
					log.Printf("Warning: Could not clear Redis: %v", err)
				} else {
					log.Println("Redis cleared successfully.")
				}
			}

		case "clear-redis":
			// Только очистка Redis
			if redis != nil {
				if err := clearRedis(redis); err != nil {
					log.Fatalf("Error clearing Redis: %v", err)
				} else {
					log.Println("Redis cleared successfully.")
				}
			} else {
				log.Fatalf("Could not connect to Redis")
			}

		default:
			// Применение миграций PostgreSQL (по умолчанию)
			if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("Error applying migrations: %v", err)
			}
			log.Println("Migrations applied successfully.")
		}
	} else {
		// Применение миграций PostgreSQL (по умолчанию)
		if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Error applying migrations: %v", err)
		}
		log.Println("Migrations applied successfully.")
	}
}

func clearRedis(redis *redisClient.Client) error {
	ctx := context.Background()
	return redis.Client.FlushDB(ctx).Err()
}
