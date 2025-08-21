package pgsql

import (
	"context"
	"fmt"
	"log/slog"
	"nickbar_v2/internal/config"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxPool собирает конфиг и создает пул
func NewPgxPool(ctx context.Context, log *slog.Logger, config *config.Config) (*pgxpool.Pool, error) {
	const ep = "storage.pgsql.NewPgxPool"

	// Формируем DSN без pool-параметров
	// dsn := fmt.Sprintf(
	// 	"postgres://%s:%s@%s:%s/%s?sslmode=%s",
	// 	config.DB.User,
	// 	config.DB.Password,
	// 	config.DB.Host,
	// 	config.DB.Port,
	// 	config.DB.Name,
	// 	config.DB.SSLMode,
	// )

	// Парсим базовый конфиг
	cfg, err := pgxpool.ParseConfig(getenv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/nickbar?sslmode=disable"))
	if err != nil {
		log.Error("error in parseconfig", slog.Any("err", err))
		return nil, fmt.Errorf("pgx parse config %s: %w", ep, err)
	}
	log.Info("cfg", slog.Any("cfg", cfg))
	// Тюним параметры пула
	// cfg.MaxConns = getInt32("DB_POOL_MAX_CONNS", 10)
	// cfg.MinConns = getInt32("DB_POOL_MIN_CONNS", 0)
	// cfg.MaxConnIdleTime = getDuration("DB_POOL_MAX_IDLE_TIME", 5*time.Minute)
	// cfg.MaxConnLifetime = getDuration("DB_POOL_MAX_LIFETIME", 30*time.Minute)
	// cfg.HealthCheckPeriod = getDuration("DB_POOL_HEALTH_CHECK", 1*time.Minute)

	// (Необязательно) Логирование pgx можно подключить тут:
	// cfg.ConnConfig.Tracer = otelpgx.NewTracer()  // пример с OpenTelemetry

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Error("error in pool creating", slog.Any("err", err))
		return nil, fmt.Errorf("pgx new pool %s: %w", ep, err)
	}

	// Проверяем соединение сразу, чтобы не получить ошибку поздно
	if err = pool.Ping(ctx); err != nil {
		log.Error("error in ping", slog.Any("err", err))
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
