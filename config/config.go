package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stretchr/testify/assert/yaml"
)

type Config struct {
	DBConfig         *DBConfig
	ServerConfig     *ServerConfig
	JWTConfig        *JWTConfig
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

type JWTConfig struct {
	Signature     string
	TokenLifeSpan time.Duration
}

type MigrationsConfig struct {
	Path string
}

func NewConfig() (*Config, error) {
	// Читаем конфиг из файла
	raw, err := loadYamlConfig("config.yml")
	if err != nil {
		log.Printf("CONFIG ERROR: %v\n", err)
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

	jwtConfig := &JWTConfig{
		Signature:     raw.JwtSignature,
		TokenLifeSpan: raw.JwtTokenLife,
	}

	migrationsConfig := &MigrationsConfig{
		Path: raw.MigrationsPath,
	}

	return &Config{
		DBConfig:         dbConfig,
		ServerConfig:     serverConfig,
		JWTConfig:        jwtConfig,
		MigrationsConfig: migrationsConfig,
	}, nil
}

type yamlConfig struct {
	ServerPort     string        `yaml:"SERVER_PORT"`
	JwtSignature   string        `yaml:"JWT_SIGNATURE"`
	PostgresUser   string        `yaml:"POSTGRES_USER"`
	PostgresPass   string        `yaml:"POSTGRES_PASSWORD"`
	PostgresDB     string        `yaml:"POSTGRES_DB"`
	PostgresPort   int           `yaml:"POSTGRES_PORT"`
	PostgresHost   string        `yaml:"POSTGRES_HOST"`
	MigrationsPath string        `yaml:"MIGRATIONS_PATH"`
	JwtTokenLife   time.Duration `yaml:"JWT_TOKEN_LIFESPAN"`
}

func loadYamlConfig(path string) (*yamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var cfg struct {
		ServerPort     string `yaml:"SERVER_PORT"`
		JwtSignature   string `yaml:"JWT_SIGNATURE"`
		PostgresUser   string `yaml:"POSTGRES_USER"`
		PostgresPass   string `yaml:"POSTGRES_PASSWORD"`
		PostgresDB     string `yaml:"POSTGRES_DB"`
		PostgresPort   string `yaml:"POSTGRES_PORT"`
		PostgresHost   string `yaml:"POSTGRES_HOST"`
		MigrationsPath string `yaml:"MIGRATIONS_PATH"`
		JwtTokenLife   string `yaml:"JWT_TOKEN_LIFESPAN"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	if cfg.ServerPort == "" {
		return nil, errors.New("SERVER_PORT is required")
	}
	if cfg.JwtSignature == "" {
		return nil, errors.New("JWT_SIGNATURE is required")
	}

	port, err := strconv.Atoi(cfg.PostgresPort)
	if err != nil {
		return nil, errors.New("invalid POSTGRES_PORT value")
	}

	tokenLife := 24 * time.Hour // значение по умолчанию
	if cfg.JwtTokenLife != "" {
		if tl, err := time.ParseDuration(cfg.JwtTokenLife); err == nil {
			tokenLife = tl
		}
	}

	return &yamlConfig{
		ServerPort:     cfg.ServerPort,
		JwtSignature:   cfg.JwtSignature,
		PostgresUser:   cfg.PostgresUser,
		PostgresPass:   cfg.PostgresPass,
		PostgresDB:     cfg.PostgresDB,
		PostgresPort:   port,
		PostgresHost:   cfg.PostgresHost,
		MigrationsPath: cfg.MigrationsPath,
		JwtTokenLife:   tokenLife,
	}, nil
}
