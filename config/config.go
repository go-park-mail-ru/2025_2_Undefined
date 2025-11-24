package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/joho/godotenv"
)

type Config struct {
	DBConfig            *DBConfig
	ServerConfig        *ServerConfig
	SessionConfig       *SessionConfig
	CSRFConfig          *CSRFConfig
	MigrationsConfig    *MigrationsConfig
	RedisConfig         *RedisConfig
	MinioConfig         *MinioConfig
	GRPCConfig          *GRPCConfig
	ElasticsearchConfig *ElasticsearchConfig
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
	Secret  string
	Timeout time.Duration
}

type MigrationsConfig struct {
	Path string
}

type GRPCConfig struct {
	AuthServiceAddr  string
	AuthServicePort  string
	UserServiceAddr  string
	UserServicePort  string
	ChatsServiceAddr string
	ChatsServicePort string
}

type ElasticsearchConfig struct {
	URL           string
	ContactsIndex string
	Username      string
	Password      string
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

	grpcConfig, err := newGRPCConfig()
	if err != nil {
		return nil, err
	}

	elasticsearchConfig, err := newElasticsearchConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		DBConfig:            dbConfig,
		ServerConfig:        serverConfig,
		SessionConfig:       sessionConfig,
		CSRFConfig:          csrfConfig,
		MigrationsConfig:    migrationsConfig,
		RedisConfig:         redisConfig,
		MinioConfig:         minioConfig,
		GRPCConfig:          grpcConfig,
		ElasticsearchConfig: elasticsearchConfig,
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

	csrftime, csrfExists := os.LookupEnv("CSRF_TIMEOUT")
	if !csrfExists {
		return nil, errors.New("CSRF_TIMEOUT is required")
	}

	csrftimeout, err := parseDurationWithDays(csrftime)
	if err != nil {
		return nil, fmt.Errorf("invalid CSRF_TIMEOUT value: %v", err)
	}
	return &CSRFConfig{
		Secret:  secret,
		Timeout: csrftimeout,
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

	cfg := &MinioConfig{
		PORT:         port,
		Host:         host,
		PublicHost:   publicHost,
		BucketName:   bucketName,
		RootUser:     accessKey,
		RootPassword: secretKey,
		UseSSL:       useSSL,
	}

	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", "getMinioConfig")
	logger.Debugf("minio config is %v", cfg)

	return cfg, nil
}

func newGRPCConfig() (*GRPCConfig, error) {
	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "localhost:50051" // default
	}

	authServicePort := os.Getenv("AUTH_GRPC_PORT")
	if authServicePort == "" {
		authServicePort = "50051" // default порт
	}

	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50052" // default
	}

	userServicePort := os.Getenv("USER_GRPC_PORT")
	if userServicePort == "" {
		userServicePort = "50052" // default порт
	}

	chatsServiceAddr := os.Getenv("CHATS_SERVICE_ADDR")
	if chatsServiceAddr == "" {
		chatsServiceAddr = "localhost:50053" // default
	}

	chatsServicePort := os.Getenv("CHATS_GRPC_PORT")
	if chatsServicePort == "" {
		chatsServicePort = "50053" // default порт
	}

	return &GRPCConfig{
		AuthServiceAddr:  authServiceAddr,
		AuthServicePort:  authServicePort,
		UserServiceAddr:  userServiceAddr,
		UserServicePort:  userServicePort,
		ChatsServiceAddr: chatsServiceAddr,
		ChatsServicePort: chatsServicePort,
	}, nil
}

func newElasticsearchConfig() (*ElasticsearchConfig, error) {
	url := os.Getenv("ELASTICSEARCH_URL")
	if url == "" {
		url = "http://localhost:9200" // default
	}

	contactsIndex := os.Getenv("ELASTICSEARCH_CONTACTS_INDEX")
	if contactsIndex == "" {
		contactsIndex = "contacts" // default
	}

	username := os.Getenv("ELASTICSEARCH_USERNAME")
	if username == "" {
		username = "admin" // default
	}

	password := os.Getenv("ELASTICSEARCH_PASSWORD")
	if password == "" {
		password = "" // default - no auth
	}

	return &ElasticsearchConfig{
		URL:           url,
		ContactsIndex: contactsIndex,
		Username:      username,
		Password:      password,
	}, nil
}
