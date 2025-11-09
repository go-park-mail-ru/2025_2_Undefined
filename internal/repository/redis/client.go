package redis

import (
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/gomodule/redigo/redis"
)

type Client struct {
	Pool *redis.Pool
}

func NewClient(cfg *config.RedisConfig) (*Client, error) {
	pool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
			if err != nil {
				return nil, err
			}

			if cfg.Password != "" {
				if _, err := conn.Do("AUTH", cfg.Password); err != nil {
					conn.Close()
					return nil, err
				}
			}

			if _, err := conn.Do("SELECT", cfg.DB); err != nil {
				conn.Close()
				return nil, err
			}

			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	// Тестируем соединение
	conn := pool.Get()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	return &Client{Pool: pool}, nil
}
