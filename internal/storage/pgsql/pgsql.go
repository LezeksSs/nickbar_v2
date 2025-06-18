package pgsql

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxPool собирает конфиг и создает пул
func NewPgxPool(ctx context.Context) (*pgxpool.Pool, error) {
	const ep = "storage.pgsql.NewPgxPool"

	// Формируем DSN без pool-параметров
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getenv("DB_USER", "postgres"),
		getenv("DB_PASS", ""),
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_NAME", "postgres"),
		getenv("DB_SSLMODE", "disable"),
	)

	// Парсим базовый конфиг
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx parse config %s: %w", ep, err)
	}

	// Тюним параметры пула
	cfg.MaxConns = getInt32("DB_POOL_MAX_CONNS", 10)
	cfg.MinConns = getInt32("DB_POOL_MIN_CONNS", 0)
	cfg.MaxConnIdleTime = getDuration("DB_POOL_MAX_IDLE_TIME", 5*time.Minute)
	cfg.MaxConnLifetime = getDuration("DB_POOL_MAX_LIFETIME", 30*time.Minute)
	cfg.HealthCheckPeriod = getDuration("DB_POOL_HEALTH_CHECK", 1*time.Minute)

	// (Необязательно) Логирование pgx можно подключить тут:
	// cfg.ConnConfig.Tracer = otelpgx.NewTracer()  // пример с OpenTelemetry

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgx new pool %s: %w", ep, err)
	}

	// Проверяем соединение сразу, чтобы не получить ошибку поздно
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgx ping %s: %w", ep, err)
	}

	return pool, nil
}

// helper берет переменную окружения или дефолт
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// helper преобразует строку-число в int32
func getInt32(key string, def int32) int32 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return int32(i)
		}
	}
	return def
}

// helper преобразует время
func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
