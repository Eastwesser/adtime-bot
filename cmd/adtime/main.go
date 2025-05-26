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
	"time"

	"go.uber.org/zap"
	_ "github.com/lib/pq"
)

// ENTRY POINT

func main() {
	// Инициализация логгера
	zapLogger, err := logger.New()
	if err != nil {
		fmt.Println(err)
	}
	defer zapLogger.Sync()

	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация Redis клиента
	redisClient := redis.New(cfg.RedisAddr, "", 0, 24*time.Hour)

	// Инициализация PostgreSQL хранилища
	pgStorage, err := storage.NewPostgresStorage(context.Background(), storage.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to init PostgreSQL storage", zap.Error(err))
	}
	defer pgStorage.Close()

	// Инициализация API клиента
	apiClient := api.NewClient(cfg.APIBaseURL, cfg.APIKey, zapLogger)

	// Создание бота
	tgBot, err := bot.New(
		cfg.TelegramToken,
		apiClient,
		redisClient,
		pgStorage,
		zapLogger,
	)
	if err != nil {
		zapLogger.Fatal("Failed to create bot", zap.Error(err))
	}

	// Обработка сигналов завершения
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Запуск бота
	if err := tgBot.Start(ctx); err != nil {
		zapLogger.Fatal("Bot stopped with error", zap.Error(err))
	}

	zapLogger.Info("Bot shutdown gracefully")
}
