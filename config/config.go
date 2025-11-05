package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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
	PublicHost   string
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
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	dbConfig, err := newDBConfig()
	if err != nil {
		return nil, err
	}

	serverConfig, err := newServerConfig()
	if err != nil {
		return nil, err
	}

	sessionConfig, err := newSessionConfig()
	if err != nil {
		return nil, err
	}

	csrfConfig, err := newCSRFConfig()
	if err != nil {
		return nil, err
	}

	migrationsConfig, err := newMigrationsConfig()
	if err != nil {
		return nil, err
	}

	redisConfig, err := newRedisConfig()
	if err != nil {
		return nil, err
	}

	minioConfig, err := newMinioConfig()
	if err != nil {
		return nil, err
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

func newDBConfig() (*DBConfig, error) {
	user, userExists := os.LookupEnv("POSTGRES_USER")
	password, passwordExists := os.LookupEnv("POSTGRES_PASSWORD")
	dbname, dbExists := os.LookupEnv("POSTGRES_DB")
	host, hostExists := os.LookupEnv("POSTGRES_HOST")
	portStr, portExists := os.LookupEnv("POSTGRES_PORT")

	if !userExists || !passwordExists || !dbExists || !hostExists || !portExists {
		return nil, errors.New("incomplete database configuration")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("invalid POSTGRES_PORT value")
	}

	return &DBConfig{
		User:            user,
		Password:        password,
		DB:              dbname,
		Port:            port,
		Host:            host,
		MaxOpenConns:    100,
		MaxIdleConns:    90,
		ConnMaxLifetime: 5 * time.Minute,
	}, nil
}

func newServerConfig() (*ServerConfig, error) {
	port, portExists := os.LookupEnv("SERVER_PORT")
	if !portExists {
		return nil, errors.New("SERVER_PORT is required")
	}

	return &ServerConfig{
		Port: port,
	}, nil
}

func newSessionConfig() (*SessionConfig, error) {
	signature, signatureExists := os.LookupEnv("SESSION_SIGNATURE")
	if !signatureExists {
		return nil, errors.New("SESSION_SIGNATURE is required")
	}

	lifespanStr, lifespanExists := os.LookupEnv("SESSION_TOKEN_LIFESPAN")
	if !lifespanExists {
		return nil, errors.New("SESSION_TOKEN_LIFESPAN is required")
	}

	lifespan, err := parseDurationWithDays(lifespanStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TOKEN_LIFESPAN value: %v", err)
	}

	return &SessionConfig{
		Signature: signature,
		LifeSpan:  lifespan,
	}, nil
}

// parseDurationWithDays парсит duration с поддержкой дней (d)
func parseDurationWithDays(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}
func newCSRFConfig() (*CSRFConfig, error) {
	secret, secretExists := os.LookupEnv("CSRF_SECRET")
	if !secretExists {
		return nil, errors.New("CSRF_SECRET is required")
	}

	return &CSRFConfig{
		Secret: secret,
	}, nil
}

func newMigrationsConfig() (*MigrationsConfig, error) {
	path, pathExists := os.LookupEnv("MIGRATIONS_PATH")
	if !pathExists {
		return nil, errors.New("MIGRATIONS_PATH is required")
	}

	return &MigrationsConfig{
		Path: path,
	}, nil
}

func newRedisConfig() (*RedisConfig, error) {
	host, hostExists := os.LookupEnv("AUTH_REDIS_HOST")
	port, portExists := os.LookupEnv("AUTH_REDIS_PORT")
	password, passwordExists := os.LookupEnv("AUTH_REDIS_PASSWORD")
	dbStr, dbExists := os.LookupEnv("AUTH_REDIS_DB")

	if !hostExists || !portExists || !passwordExists || !dbExists {
		return nil, errors.New("incomplete Redis configuration")
	}

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return nil, errors.New("invalid AUTH_REDIS_DB value")
	}

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}, nil
}

func newMinioConfig() (*MinioConfig, error) {
	port, portExists := os.LookupEnv("MINIO_PORT")
	host, hostExists := os.LookupEnv("MINIO_HOST")
	publicHost, publicHostExists := os.LookupEnv("MINIO_PUBLIC_HOST")
	bucketName, bucketNameExists := os.LookupEnv("MINIO_BUCKET_NAME")
	accessKey, accessKeyExists := os.LookupEnv("MINIO_ACCESS_KEY")
	secretKey, secretKeyExists := os.LookupEnv("MINIO_SECRET_KEY")
	useSSLStr, useSSLExists := os.LookupEnv("MINIO_USE_SSL")

	if !portExists || !hostExists || !publicHostExists || !bucketNameExists || !accessKeyExists || !secretKeyExists || !useSSLExists {
		return nil, errors.New("incomplete MinIO configuration")
	}

	useSSL, err := strconv.ParseBool(useSSLStr)
	if err != nil {
		return nil, errors.New("invalid MINIO_USE_SSL value")
	}

	return &MinioConfig{
		PORT:         port,
		Host:         host,
		PublicHost:   publicHost,
		BucketName:   bucketName,
		RootUser:     accessKey,
		RootPassword: secretKey,
		UseSSL:       useSSL,
	}, nil
}
