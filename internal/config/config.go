package config

import "github.com/caarlos0/env/v9"

// BOT CONFIGS

type Config struct {
	TelegramToken string `env:"TELEGRAM_TOKEN,required"`
	APIBaseURL    string `env:"API_BASE_URL,required"`
	APIKey        string `env:"API_KEY,required"`
	RedisAddr     string `env:"REDIS_ADDR,required"`
	DBHost        string `env:"DB_HOST,required"`
	DBPort        int    `env:"DB_PORT,required"`
	DBUser        string `env:"DB_USER,required"`
	DBPassword    string `env:"DB_PASSWORD,required"`
	DBName        string `env:"DB_NAME,required"`
}

func Load() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
