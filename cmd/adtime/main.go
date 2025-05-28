package main

import (
	"adtime-bot/internal/bot"
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/api"
	"adtime-bot/pkg/logger"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger
	zapLogger, err := logger.New()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer zapLogger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize Redis client
	redisClient := redis.New(
		cfg.Redis.Addr, 
		cfg.Redis.Password, 
		cfg.Redis.DB, 
		cfg.Redis.TTL,
	)

	// Initialize PostgreSQL storage
	pgStorage, err := storage.NewPostgresStorage(context.Background(), *cfg, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to init PostgreSQL storage", zap.Error(err))
	}
	defer pgStorage.Close()

	// Initialize API client
	apiClient := api.NewClient(
		cfg.API.BaseURL, 
		cfg.API.Key, 
		zapLogger, 
		cfg.API.RequestTimeout,
	)

	// Create bot instance
	tgBot, err := bot.New(
		cfg.Telegram.Token,
		apiClient,
		redisClient,
		pgStorage,
		zapLogger,
		cfg,
	)
	if err != nil {
		zapLogger.Fatal("Failed to create bot", zap.Error(err))
	}

	// Handle shutdown signals
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Start the bot
	zapLogger.Info("Starting bot")
	if err := tgBot.Start(ctx); err != nil {
		zapLogger.Fatal("Bot stopped with error", zap.Error(err))
	}

	zapLogger.Info("Bot shutdown gracefully")
}
