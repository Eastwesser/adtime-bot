package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	Telegram struct {
		Token string `env:"TELEGRAM_TOKEN,required"`
	}

	API struct {
		BaseURL         string        `env:"API_BASE_URL,required"`
		Key             string        `env:"API_KEY,required"`
		RequestTimeout  time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"30s"`
	}

	Redis struct {
		Addr     string        `env:"REDIS_ADDR,required"`
		Password string        `env:"REDIS_PASSWORD"`
		DB       int           `env:"REDIS_DB" envDefault:"0"`
		TTL      time.Duration `env:"REDIS_TTL" envDefault:"24h"`
	}

	Database struct {
		Host            string        `env:"DB_HOST,required"`
		Port            int           `env:"DB_PORT,required"`
		User            string        `env:"DB_USER,required"`
		Password        string        `env:"DB_PASSWORD,required"`
		Name            string        `env:"DB_NAME,required"`
		MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
		MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
		ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
		ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"2m"`
	}

	Admin struct {
		IDs []int64 `env:"ADMIN_IDS" envSeparator:","`
	}
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if len(cfg.Admin.IDs) == 0 {
		return nil, fmt.Errorf("at least one admin ID is required")
	}

	return &cfg, nil
}

// func Load(envPrefix string) (*Config, error) {
//     var cfg Config
//     opts := []env.Options{}
//     if envPrefix != "" {
//         opts = append(opts, env.Prefix(envPrefix))
//     }
//     if err := env.ParseWithOptions(&cfg, opts...); err != nil {
//         return nil, err
//     }
//     return &cfg, nil
// }

func (c *Config) Validate() error {
    if c.Database.Host == "" {
        return errors.New("database host is required")
    }
    // Add other validations as needed
    return nil
}
