package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DBConfig         *DBConfig
	ServerConfig     *ServerConfig
	SessionConfig    *SessionConfig
	CSRFConfig       *CSRFConfig
	MigrationsConfig *MigrationsConfig
	RedisConfig      *RedisConfig
	MinioConfig      *MinioConfig
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

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type MinioConfig struct {
	PORT         string
	Host         string
	PublicHost   string // Хост для публичных URLs (например localhost)
	BucketName   string
	RootUser     string
	RootPassword string
	UseSSL       bool
}

type ServerConfig struct {
	Port string
}

type SessionConfig struct {
	Signature string
	LifeSpan  time.Duration
}

type CSRFConfig struct {
	Secret string
}

type MigrationsConfig struct {
	Path string
}

func NewConfig() (*Config, error) {
	raw, err := loadYamlConfig(".env/config.yml")
	if err != nil {
		// Если нет .env/config.yml, пробуем обычный config.yml
		raw, err = loadYamlConfig("config.yml")
		if err != nil {
			return nil, err
		}
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

	redisConfig := &RedisConfig{
		Port:     raw.AuthPort,
		Host:     raw.AuthHost,
		Password: raw.AuthPass,
		DB:       raw.AuthDB,
	}

	minioConfig := &MinioConfig{
		PORT:         raw.MinioPort,
		Host:         raw.MinioHost,
		PublicHost:   raw.MinioPublicHost,
		BucketName:   raw.MinioBucketName,
		RootUser:     raw.MinioAccessKey,
		RootPassword: raw.MinioSecretKey,
		UseSSL:       raw.MinioUseSSL,
	}

	serverConfig := &ServerConfig{
		Port: raw.ServerPort,
	}

	sessionConfig := &SessionConfig{
		Signature: raw.Signature,
		LifeSpan:  raw.SessionTokenLife,
	}

	csrfConfig := &CSRFConfig{
		Secret: raw.CSRFSecret,
	}

	migrationsConfig := &MigrationsConfig{
		Path: raw.MigrationsPath,
	}

	return &Config{
		DBConfig:         dbConfig,
		ServerConfig:     serverConfig,
		SessionConfig:    sessionConfig,
		CSRFConfig:       csrfConfig,
		MigrationsConfig: migrationsConfig,
		RedisConfig:      redisConfig,
		MinioConfig:      minioConfig,
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
	AuthPass         string        `yaml:"AUTH_REDIS_PASSWORD"`
	AuthDB           int           `yaml:"AUTH_REDIS_DB"`
	AuthPort         string        `yaml:"AUTH_REDIS_PORT"`
	AuthHost         string        `yaml:"AUTH_REDIS_HOST"`
	CSRFSecret       string        `yaml:"CSRF_SECRET"`
	MinioPort        string        `yaml:"MINIO_PORT"`
	MinioHost        string        `yaml:"MINIO_HOST"`
	MinioPublicHost  string        `yaml:"MINIO_PUBLIC_HOST"`
	MinioAccessKey   string        `yaml:"MINIO_ACCESS_KEY"`
	MinioSecretKey   string        `yaml:"MINIO_SECRET_KEY"`
	MinioBucketName  string        `yaml:"MINIO_BUCKET_NAME"`
	MinioUseSSL      bool          `yaml:"MINIO_USE_SSL"`
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
		AuthPass         string `yaml:"AUTH_REDIS_PASSWORD"`
		AuthDB           string `yaml:"AUTH_REDIS_DB"`
		AuthPort         string `yaml:"AUTH_REDIS_PORT"`
		AuthHost         string `yaml:"AUTH_REDIS_HOST"`
		CSRFSecret       string `yaml:"CSRF_SECRET"`
		MinioPort        string `yaml:"MINIO_PORT"`
		MinioHost        string `yaml:"MINIO_HOST"`
		MinioAccessKey   string `yaml:"MINIO_ACCESS_KEY"`
		MinioSecretKey   string `yaml:"MINIO_SECRET_KEY"`
		MinioBucketName  string `yaml:"MINIO_BUCKET_NAME"`
		MinioUseSSL      bool   `yaml:"MINIO_USE_SSL"`
		MinioPublicHost  string `yaml:"MINIO_PUBLIC_HOST"`
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

	authDb, err := strconv.Atoi(cfg.AuthDB)
	if err != nil {
		return nil, errors.New("invalid POSTGRES_PORT value")
	}

	if cfg.AuthPort == "" {
		return nil, errors.New("AUTH_Redis_PORT is required")
	}

	if cfg.AuthHost == "" {
		return nil, errors.New("AUTH_Redis_Host is required")
	}

	if cfg.MinioPort == "" {
		return nil, errors.New("MINIO_PORT is required")
	}

	if cfg.MinioHost == "" {
		return nil, errors.New("MINIO_HOST is required")
	}

	if cfg.MinioPublicHost == "" {
		return nil, errors.New("MINIO_PUBLIC_HOST is required")
	}

	if cfg.MinioBucketName == "" {
		return nil, errors.New("MINIO_BUCKET_NAME is required")
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
		AuthPass:         cfg.AuthPass,
		AuthDB:           authDb,
		AuthPort:         cfg.AuthPort,
		AuthHost:         cfg.AuthHost,
		CSRFSecret:       cfg.CSRFSecret,
		MinioPort:        cfg.MinioPort,
		MinioHost:        cfg.MinioHost,
		MinioAccessKey:   cfg.MinioAccessKey,
		MinioSecretKey:   cfg.MinioSecretKey,
		MinioBucketName:  cfg.MinioBucketName,
		MinioUseSSL:      cfg.MinioUseSSL,
		MinioPublicHost:  cfg.MinioPublicHost,
	}, nil
}
