package bot

import (
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/api"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"strings"
)

// BOT CORE LOGIC

type Bot struct {
	api     *api.Client
	bot     *tgbotapi.BotAPI
	logger  *zap.Logger
	state   *StateStorage
	storage *storage.PostgresStorage
}

func New(
	token string,
	apiClient *api.Client,
	redisClient *redis.Client,
	pgStorage *storage.PostgresStorage,
	logger *zap.Logger,
) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return &Bot{
		api:     apiClient,
		bot:     botAPI,
		logger:  logger,
		state:   NewStateStorage(redisClient),
		storage: pgStorage,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting bot", zap.String("username", b.bot.Self.UserName))

	updates := b.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			if update.Message != nil {
				b.handleMessage(ctx, update.Message)
			} else if update.CallbackQuery != nil {
				b.handleCallback(ctx, update.CallbackQuery)
			}
		}
	}
}

func (b *Bot) handleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	if strings.HasPrefix(data, "texture:") {
		textureID := strings.TrimPrefix(data, "texture:")
		width, height := b.state.GetDimensions(chatID)

		// Получаем цену за дм² для выбранной текстуры
		texturePrice, err := b.api.GetTexturePrice(ctx, textureID)
		if err != nil {
			b.logger.Error("Failed to get texture price", zap.Error(err))
			b.sendError(chatID, "Не удалось получить цену текстуры")
			return
		}

		price := calculatePrice(width, height, texturePrice)
		b.state.SetTexture(chatID, textureID, price)

		// Показываем подтверждение выбора
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"Вы выбрали текстуру. Итоговая цена: %.2f руб\n\nКогда вам удобно выполнить заказ?",
			price,
		))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Сегодня"),
				tgbotapi.NewKeyboardButton("Завтра"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Выбрать дату вручную"),
			),
		)
		b.sendMessage(msg)
		b.state.SetWaitingForDateSelection(chatID)

		// Удаляем inline-клавиатуру
		delMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
		b.bot.Send(delMsg)
	}

}


func (b *Bot) sendMessage(msg tgbotapi.MessageConfig) {
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("Failed to send message", zap.Error(err))
	}
}

func (b *Bot) sendError(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+text)
	b.sendMessage(msg)
}
