package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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

	// Ожидаем готовности базы данных
	if err := waitForDatabase(cfg); err != nil {
		log.Fatalf("Database is not ready: %v", err)
	}

	// Формируем строку подключения для migrate
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.DBName,
		cfg.DBConfig.SSLMode,
	)

	m, err := migrate.New(
		cfg.MigrationsConfig.Path,
		dsn,
	)
	if err != nil {
		log.Fatalf("Error initializing migrations: %v", err)
	}
	defer m.Close()

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

func waitForDatabase(cfg *config.Config) error {
	dbConfig := &repository.Config{
		Host:     cfg.DBConfig.Host,
		Port:     cfg.DBConfig.Port,
		User:     cfg.DBConfig.User,
		Password: cfg.DBConfig.Password,
		DBName:   cfg.DBConfig.DBName,
		SSLMode:  cfg.DBConfig.SSLMode,
	}

	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", i+1, maxRetries)

		dbConnector, err := repository.NewDBConnector(dbConfig)
		if err != nil {
			log.Printf("Failed to connect: %v", err)
			if i == maxRetries-1 {
				return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(retryInterval)
			continue
		}

		// Проверяем подключение
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = dbConnector.GetPool().Ping(ctx)
		cancel()

		if err != nil {
			log.Printf("Ping failed: %v", err)
			dbConnector.Close()
			if i == maxRetries-1 {
				return fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(retryInterval)
			continue
		}

		log.Println("Successfully connected to database!")
		dbConnector.Close()
		return nil
	}

	return fmt.Errorf("database connection timeout after %d attempts", maxRetries)
}

func clearRedis(redis *redisClient.Client) error {
	conn := redis.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("FLUSHALL")
	return err
}
