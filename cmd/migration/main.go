package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func GetConnectionString(conf *config.DBConfig) (string, error) {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DB,
	), nil
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	dsn, err := GetConnectionString(cfg.DBConfig)
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

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Error applying migrations: %v", err)
	}

	log.Println("Migrations applied successfully.")
}
