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

	Redis struct {
		Addr     string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
		Password string        `env:"REDIS_PASSWORD" envDefault:""`
		DB       int           `env:"REDIS_DB" envDefault:"0"`
		TTL      time.Duration `env:"REDIS_TTL" envDefault:"24h"`
	}

	Database struct {
		Host            string        `env:"DB_HOST" envDefault:"localhost"`
		Port            int           `env:"DB_PORT" envDefault:"5432"`
		User            string        `env:"DB_USER" envDefault:"postgres"`
		Password        string        `env:"DB_PASSWORD" envDefault:"postgres"`
		Name            string        `env:"DB_NAME" envDefault:"adtime"`
		MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
		MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
		ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
		ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"2m"`
	}

	Admin struct {
		ChatID    int64   `env:"ADMIN_CHAT_ID"`
		ChannelID int64   `env:"CHANNEL_ID"`
		IDs       []int64 `env:"ADMIN_IDS"`
	}

	Pricing struct {
        LeatherPricePerDM2    float64 `env:"LEATHER_PRICE_PER_DM2" envDefault:"25.0"`
        ProcessingCostPerDM2  float64 `env:"PROCESSING_COST_PER_DM2" envDefault:"31.25"`
        PaymentCommissionRate float64 `env:"PAYMENT_COMMISSION_RATE" envDefault:"0.03"`
        SalesTaxRate          float64 `env:"SALES_TAX_RATE" envDefault:"0.06"`
        MarkupMultiplier      float64 `env:"MARKUP_MULTIPLIER" envDefault:"2.5"`
    }

	MaxDimensions struct {
        Width  int `env:"MAX_WIDTH" envDefault:"80"`
        Height int `env:"MAX_HEIGHT" envDefault:"50"`
    }

}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Telegram.Token == "" {
		return errors.New("telegram token is required")
	}

	if c.Database.Host == "" {
		return errors.New("database host is required")
	}

	if c.Database.Name == "" {
		return errors.New("database name is required")
	}

	return nil
}
