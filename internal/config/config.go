package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	TelegramToken      string        `env:"TELEGRAM_TOKEN,required"`
	APIBaseURL         string        `env:"API_BASE_URL,required"`
	APIKey             string        `env:"API_KEY,required"`
	RedisAddr          string        `env:"REDIS_ADDR,required"`
	RedisPassword      string        `env:"REDIS_PASSWORD"`
	RedisDB            int           `env:"REDIS_DB" envDefault:"0"`
	RedisTTL           time.Duration `env:"REDIS_TTL" envDefault:"24h"`
	DBHost             string        `env:"DB_HOST,required"`
	DBPort             int           `env:"DB_PORT,required"`
	DBUser             string        `env:"DB_USER,required"`
	DBPassword         string        `env:"DB_PASSWORD,required"`
	DBName             string        `env:"DB_NAME,required"`
	DBMaxOpenConns     int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns     int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	DBConnMaxLifetime  time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	DBConnMaxIdleTime  time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"2m"`
	AdminIDs           []int64       `env:"ADMIN_IDS" envSeparator:","`
	HTTPRequestTimeout time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"30s"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    // Validate required fields
    if len(cfg.AdminIDs) == 0 {
        return nil, fmt.Errorf("at least one admin ID is required")
    }
    
    return &cfg, nil
}