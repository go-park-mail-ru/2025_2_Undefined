package repository

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetConnectionString(conf *config.DBConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DB,
	)
}

func NewPgxPool(ctx context.Context, conf *config.DBConfig) (*pgxpool.Pool, error) {
	connString := GetConnectionString(conf)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Настройка параметров connection pool для оптимальной работы
	// MaxConns - максимальное количество открытых соединений в пуле
	// Значение из config.DBConfig.MaxOpenConns (по умолчанию 100)
	// ОБОСНОВАНИЕ: микросервисная архитектура с несколькими сервисами требует
	// достаточного количества соединений. 100 покрывает пиковые нагрузки.
	poolConfig.MaxConns = int32(conf.MaxOpenConns)

	// MinConns - минимальное количество поддерживаемых соединений
	// Значение из config.DBConfig.MaxIdleConns (по умолчанию 90)
	// ОБОСНОВАНИЕ: держим большую часть соединений открытыми для быстрого ответа
	// на запросы без overhead на установку нового соединения
	poolConfig.MinConns = int32(conf.MaxIdleConns)

	// MaxConnLifetime - максимальное время жизни соединения
	// Значение из config.DBConfig.ConnMaxLifetime (по умолчанию 5 минут)
	// ОБОСНОВАНИЕ: PostgreSQL может закрыть старые соединения, поэтому мы
	// закрываем их чуть раньше для предотвращения ошибок "broken pipe"
	poolConfig.MaxConnLifetime = conf.ConnMaxLifetime

	// MaxConnIdleTime - максимальное время неактивного соединения
	// Устанавливаем в половину от MaxConnLifetime (2.5 минуты)
	// ОБОСНОВАНИЕ: закрываем неиспользуемые соединения для освобождения ресурсов
	// на стороне БД, но не слишком агрессивно чтобы не пересоздавать часто
	poolConfig.MaxConnIdleTime = conf.ConnMaxLifetime / 2

	// HealthCheckPeriod - периодичность проверки здоровья соединений
	// Устанавливаем 1 минуту
	// ОБОСНОВАНИЕ: регулярная проверка позволяет обнаружить "мертвые" соединения
	poolConfig.HealthCheckPeriod = 1 * conf.ConnMaxLifetime / 5

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
