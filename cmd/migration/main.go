package main

import (
	"errors"
	"log"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
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

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Error applying migrations: %v", err)
	}

	log.Println("Migrations applied successfully.")
}
