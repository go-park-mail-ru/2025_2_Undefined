package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/stretchr/testify/assert/yaml"
)

type Config struct {
	DBConfig         *DBConfig
	ServerConfig     *ServerConfig
	SessionConfig    *SessionConfig
	MigrationsConfig *MigrationsConfig
}

type DBConfig struct {
	User            string
	Password        string
	DB              string
	Port            int
	Host            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type ServerConfig struct {
	Port string
}

type SessionConfig struct {
	Signature string
	LifeSpan  time.Duration
}

type MigrationsConfig struct {
	Path string
}

func NewConfig() (*Config, error) {
	// Читаем конфиг из файла
	raw, err := loadYamlConfig("config.yml")
	if err != nil {
		return nil, err
	}

	dbConfig := &DBConfig{
		User:            raw.PostgresUser,
		Password:        raw.PostgresPass,
		DB:              raw.PostgresDB,
		Port:            raw.PostgresPort,
		Host:            raw.PostgresHost,
		MaxOpenConns:    100,
		MaxIdleConns:    90,
		ConnMaxLifetime: 5 * time.Minute,
	}

	serverConfig := &ServerConfig{
		Port: raw.ServerPort,
	}

	sessionConfig := &SessionConfig{
		Signature: raw.Signature,
		LifeSpan:  raw.SessionTokenLife,
	}

	migrationsConfig := &MigrationsConfig{
		Path: raw.MigrationsPath,
	}

	return &Config{
		DBConfig:         dbConfig,
		ServerConfig:     serverConfig,
		SessionConfig:    sessionConfig,
		MigrationsConfig: migrationsConfig,
	}, nil
}

type yamlConfig struct {
	ServerPort       string        `yaml:"SERVER_PORT"`
	Signature        string        `yaml:"SESSION_SIGNATURE"`
	PostgresUser     string        `yaml:"POSTGRES_USER"`
	PostgresPass     string        `yaml:"POSTGRES_PASSWORD"`
	PostgresDB       string        `yaml:"POSTGRES_DB"`
	PostgresPort     int           `yaml:"POSTGRES_PORT"`
	PostgresHost     string        `yaml:"POSTGRES_HOST"`
	MigrationsPath   string        `yaml:"MIGRATIONS_PATH"`
	SessionTokenLife time.Duration `yaml:"SESSION_TOKEN_LIFESPAN"`
}

func loadYamlConfig(path string) (*yamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var cfg struct {
		ServerPort       string `yaml:"SERVER_PORT"`
		Signature        string `yaml:"SESSION_SIGNATURE"`
		PostgresUser     string `yaml:"POSTGRES_USER"`
		PostgresPass     string `yaml:"POSTGRES_PASSWORD"`
		PostgresDB       string `yaml:"POSTGRES_DB"`
		PostgresPort     string `yaml:"POSTGRES_PORT"`
		PostgresHost     string `yaml:"POSTGRES_HOST"`
		MigrationsPath   string `yaml:"MIGRATIONS_PATH"`
		SessionTokenLife string `yaml:"SESSION_TOKEN_LIFESPAN"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	if cfg.ServerPort == "" {
		return nil, errors.New("SERVER_PORT is required")
	}
	if cfg.Signature == "" {
		return nil, errors.New("SESSION_SIGNATURE is required")
	}

	port, err := strconv.Atoi(cfg.PostgresPort)
	if err != nil {
		return nil, errors.New("invalid POSTGRES_PORT value")
	}

	tokenLife := 30 * 24 * time.Hour // значение по умолчанию
	if cfg.SessionTokenLife != "" {
		if tl, err := time.ParseDuration(cfg.SessionTokenLife); err == nil {
			tokenLife = tl
		}
	}

	return &yamlConfig{
		ServerPort:       cfg.ServerPort,
		Signature:        cfg.Signature,
		PostgresUser:     cfg.PostgresUser,
		PostgresPass:     cfg.PostgresPass,
		PostgresDB:       cfg.PostgresDB,
		PostgresPort:     port,
		PostgresHost:     cfg.PostgresHost,
		MigrationsPath:   cfg.MigrationsPath,
		SessionTokenLife: tokenLife,
	}, nil
}
