package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Host        string        `mapstructure:"http_host"`
	Port        int           `mapstructure:"http_port"`
	Timeout     time.Duration `mapstructure:"timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	LogLevel    string        `mapstructure:"log_level"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Address      string        `mapstructure:"address"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	TokenExpire  time.Duration `mapstructure:"token_ttl"`
}

type Config struct {
	App   AppConfig   `mapstructure:"app"`
	DB    DBConfig    `mapstructure:"db"`
	Redis RedisConfig `mapstructure:"redis"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")

	v.AddConfigPath("./config")

	v.SetConfigName("base")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read base.yaml: %w", err)
	}

	env := os.Getenv("APP_ENV") // "local" | "prod"
	if env == "" {
		env = "local"
	}

	v.SetConfigName(env)
	if err := v.MergeInConfig(); err != nil {
		// Для prod отсутствие файла критично.
		if env == "prod" {
			return nil, fmt.Errorf("read %s.yaml: %w", env, err)
		}
	}

	_ = v.BindEnv("db.user", "APP_DB_USER")
	_ = v.BindEnv("db.password", "APP_DB_PASSWORD")
	_ = v.BindEnv("db.host", "APP_DB_HOST")
	_ = v.BindEnv("db.port", "APP_DB_PORT")
	_ = v.BindEnv("db.name", "APP_DB_NAME")
	_ = v.BindEnv("db.sslmode", "APP_DB_SSLMODE")
	_ = v.BindEnv("redis.addr", "APP_REDIS_ADDR")
	_ = v.BindEnv("redis.db", "APP_REDIS_DB")
	_ = v.BindEnv("app.env", "APP_ENV")

	v.AutomaticEnv() // даёт приоритет переменным среды

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// ➏ Быстрая валидация (fail-fast)
	if cfg.DB.User == "" || cfg.DB.Password == "" {
		return nil, fmt.Errorf("missing DB credentials in env (APP_DB_USER / APP_DB_PASSWORD)")
	}
	if cfg.App.Port == 0 {
		return nil, fmt.Errorf("http_port must be > 0")
	}

	return &cfg, nil
}
