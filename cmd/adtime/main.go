package main

import (
	"adtime-bot/internal/bot"
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize Redis client (используем pkg/redis)
	redisClient := redis.New(
		cfg.Redis.Addr, 
		cfg.Redis.Password, 
		cfg.Redis.DB, 
		cfg.Redis.TTL,
	)
	defer redisClient.Close()

	// Initialize PostgreSQL storage
	pgStorage, err := storage.NewPostgresStorage(context.Background(), *cfg, redisClient, logger)
	if err != nil {
		logger.Fatal("Failed to init PostgreSQL storage", zap.Error(err))
	}
	defer pgStorage.Close()

	// Create bot instance
	tgBot, err := bot.New(
		cfg.Telegram.Token,
		redisClient,
		pgStorage,
		logger,
		cfg,
	)
	if err != nil {
		logger.Fatal("Failed to create bot", zap.Error(err))
	}

	// Handle shutdown signals
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Start the bot
	logger.Info("Starting bot")
	if err := tgBot.Start(ctx); err != nil {
		logger.Fatal("Bot stopped with error", zap.Error(err))
	}

	logger.Info("Bot shutdown gracefully")
}
