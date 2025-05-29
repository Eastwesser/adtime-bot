package bot

import (
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Bot struct {
	bot      *tgbotapi.BotAPI
	logger   *zap.Logger
	state    *StateStorage
	storage  *storage.PostgresStorage
	cfg      *config.Config
	mu       sync.Mutex
	handlers map[string]func(context.Context, int64, string)
}

const (
	StepPrivacyAgreement = "privacy_agreement"
	StepServiceSelection = "service_selection"
	StepServiceType      = "service_type"
	StepDimensions       = "dimensions"
	StepDateSelection    = "date_selection"
	StepManualDateInput  = "manual_date_input"
	StepDateConfirmation = "date_confirmation"
	StepContactMethod    = "contact_method"
	StepPhoneNumber      = "phone_number"
)

func New(
	token string,
	redisClient *redis.Client,
	pgStorage *storage.PostgresStorage,
	logger *zap.Logger,
	cfg *config.Config,
) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	botAPI.Debug = true // Enable debug for development

	logger.Info("Bot authorized",
		zap.String("username", botAPI.Self.UserName),
		zap.Int64("id", botAPI.Self.ID))

	b := &Bot{
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
		StepPrivacyAgreement: b.handlePrivacyAgreement,
		StepServiceSelection: b.handleServiceSelection,
		StepServiceType:      b.handleServiceType,
		StepDimensions:       b.handleDimensionsSize,
		StepDateSelection:    b.handleDateSelection,
		StepManualDateInput:  b.handleManualDateInput,
		StepDateConfirmation: b.handleDateConfirmation,
		StepContactMethod:    b.handleContactMethod,
		StepPhoneNumber:      b.handlePhoneNumber,
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

	if msg.Contact != nil {
		if state, err := b.state.Get(ctx, chatID); err == nil && state.Step == StepContactMethod {
			// Normalize the contact phone number before processing
			normalized := NormalizePhoneNumber(msg.Contact.PhoneNumber)
			b.handlePhoneNumber(ctx, chatID, normalized)
			return
		}
	}

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
	chatID := callback.Message.Chat.ID
	data := callback.Data

	b.logger.Debug("Processing callback",
		zap.Int64("chat_id", chatID),
		zap.String("data", data))

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