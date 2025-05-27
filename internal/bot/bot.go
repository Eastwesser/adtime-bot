package bot

import (
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/api"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Bot struct {
	api      *api.Client
	bot      *tgbotapi.BotAPI
	logger   *zap.Logger
	state    *StateStorage
	storage  *storage.PostgresStorage
	cfg      *config.Config
	mu       sync.Mutex
	handlers map[string]func(context.Context, int64, string)
}

func New(
	token string,
	apiClient *api.Client,
	redisClient *redis.Client,
	pgStorage *storage.PostgresStorage,
	logger *zap.Logger,
	cfg *config.Config,
) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	botAPI.Debug = false

	logger.Info("Bot authorized", 
		zap.String("username", botAPI.Self.UserName),
		zap.Int64("id", botAPI.Self.ID))

	b := &Bot{
		api:     apiClient,
		bot:     botAPI,
		logger:  logger,
		state:   NewStateStorage(redisClient),
		storage: pgStorage,
		cfg:     cfg,
	}

	b.registerHandlers()
	return b, nil
}

func (b *Bot) registerHandlers() {
	b.handlers = map[string]func(context.Context, int64, string){
		StepPrivacyAgreement:   b.handlePrivacyAgreement,
		StepServiceSelection:   b.handleServiceSelection,
		StepServiceInput:       b.handleServiceInput,
		StepDimensions:         b.handleDimensionsSize,
		StepDateSelection:      b.handleDateSelection,
		StepManualDateInput:    b.handleManualDateInput,
		StepDateConfirmation:   b.handleDateConfirmation,
		StepPhoneNumber:        b.handlePhoneNumber,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting bot")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Shutting down bot")
			return nil

		case update := <-updates:
			b.mu.Lock()
			if update.Message != nil {
				b.processMessage(ctx, update.Message)
			} else if update.CallbackQuery != nil {
				b.processCallback(ctx, update.CallbackQuery)
			}
			b.mu.Unlock()
		}
	}
}

func (b *Bot) processMessage(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	b.logger.Debug("Processing message",
		zap.Int64("chat_id", chatID),
		zap.String("text", msg.Text))

	if msg.IsCommand() {
		b.handleCommand(ctx, chatID, msg.Command())
		return
	}

	state, err := b.state.Get(ctx, chatID)
	if err != nil {
		b.logger.Error("Failed to get user state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.sendError(chatID, "Ошибка при обработке запроса")
		return
	}

	if handler, exists := b.handlers[state.Step]; exists {
		handler(ctx, chatID, msg.Text)
	} else {
		b.handleDefault(ctx, chatID)
	}
}


func (b *Bot) processCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID // This is already int64
    data := callback.Data

    b.logger.Debug("Processing callback",
        zap.Int64("chat_id", chatID),
        zap.String("data", data))

    if !strings.HasPrefix(data, "texture:") {
        b.logger.Warn("Unknown callback data format", 
            zap.String("data", data))
        b.sendError(chatID, "Неизвестный формат команды")
        return
    }
    // Handle texture selection here or call another method
    b.handleTextureSelection(ctx, callback)
}

func (b *Bot) sendMessage(msg tgbotapi.MessageConfig) {
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("Failed to send message",
			zap.Int64("chat_id", msg.ChatID),
			zap.String("text", msg.Text),
			zap.Error(err))
	}
}

func (b *Bot) sendError(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+text)
	b.sendMessage(msg)
}

func (b *Bot) sendAdminNotification(ctx context.Context, message string) {
	for _, adminID := range b.cfg.AdminIDs {
		msg := tgbotapi.NewMessage(adminID, message)
		if _, err := b.bot.Send(msg); err != nil {
			b.logger.Error("Failed to send admin notification",
				zap.Int64("admin_id", adminID),
				zap.Error(err))
		}
	}
}
