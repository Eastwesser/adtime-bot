package main

import (
	"adtime-bot/internal/bot"
	"adtime-bot/internal/bot/amocrm"
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"net/http"
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

	// Handle shutdown signals
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Invalid config", zap.Error(err))
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
		cfg.AmoCRM.WebhookSecret, // Add this
	)
	if err != nil {
		logger.Fatal("Failed to create bot", zap.Error(err))
	}

	// Create HTTP server for both webhooks
	mux := http.NewServeMux()
	
	// Telegram webhook
	mux.HandleFunc("/telegram", func(w http.ResponseWriter, r *http.Request) {
		// Your existing Telegram webhook handling
	})

	// amoCRM webhook
	amocrmHandler := amocrm.NewWebhookHandler(tgBot)
	mux.HandleFunc("/amocrm", amocrmHandler.HandleWebhook)

	server := &http.Server{
		Addr:    cfg.HTTP.ListenAddr,
		Handler: mux,
	}

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Start the bot with webhook
	logger.Info("Starting bot with webhook",
		zap.String("webhook_url", cfg.Telegram.WebhookURL),
		zap.String("listen_addr", cfg.Telegram.ListenAddr))

	if err := tgBot.StartWebhook(ctx, cfg.Telegram.WebhookURL, cfg.Telegram.ListenAddr); err != nil {
		logger.Fatal("Bot stopped with error", zap.Error(err))
	}

	logger.Info("Bot shutdown gracefully")
}
