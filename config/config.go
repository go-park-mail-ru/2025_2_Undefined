package config

import "time"

type Config struct {
	JWTSecret string
	Port      string
	JWTTTL    time.Duration
}

func NewConfig() *Config {
	return &Config{
		JWTSecret: "secret-jwt-key-v1",
		Port:      "8080",
		JWTTTL:    24 * time.Hour,
	}
}
