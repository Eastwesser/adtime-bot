package start

import (
	"adtime-bot/internal/bot"
	"adtime-bot/internal/bot/keyboards"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	startMsg = `Привет! 👋

    ⚠️ Прежде чем продолжить, вы должны согласиться с:
    1. Нашей Политикой конфиденциальности
    2. Telegram Bot Privacy Policy (TPA)
    
    Используя бота, вы подтверждаете согласие на обработку данных в соответствии с этими документами.`

	privacyPolicyMsg = `🔐 <b>Политика конфиденциальности</b>

    Ваши данные обрабатываются в соответствии с:
    1. Нашей политикой конфиденциальности
    2. Telegram Bot Privacy Policy (TPA)

    <u>Мы собираем и храним:</u>
    - Контактные данные (номер телефона)
    - Параметры заказа (размеры, текстура)
    - Даты выполнения заказа

    <u>Telegram TPA требует:</u>
    - Явного согласия на обработку данных
    - Возможности удаления данных по запросу
    - Соблюдения GDPR и других регуляций

    📜 Полный текст: %s`
)

type StartHandler struct {
	logger       *zap.Logger
	botapi       *tgbotapi.BotAPI
	stateManager StateManager
	pgStorage    PgStorage
}

func New(
	logger *zap.Logger,
	botapi *tgbotapi.BotAPI,
	stateManager StateManager,
	pgStorage PgStorage,
) *StartHandler {
	return &StartHandler{
		logger:       logger,
		botapi:       botapi,
		stateManager: stateManager,
		pgStorage:    pgStorage,
	}
}

func (s *StartHandler) Handle(ctx context.Context, msg *tgbotapi.Message) {
	chatId := msg.Chat.ID
	agreed, phone, err := s.pgStorage.GetUserAgreement(ctx, chatId)
	if err != nil {
		s.logger.Error("Failed to check user agreement", zap.Error(err))
	}

	if agreed && phone != "" {
		// If user has signed TPA, show menu
		s.showMainMenu(ctx, chatId, phone)
		return
	}

	// send start msg
	s.sendMessage(tgbotapi.NewMessage(chatId, startMsg))
	s.showPrivacyPolicy(chatId)

	if err := s.stateManager.SetStep(ctx, chatId, bot.StepPrivacyAgreement); err != nil {
		s.logger.Error("Failed to set privacy agreement state",
			zap.Int64("chat_id", chatId),
			zap.Error(err))
	}
}

func (s *StartHandler) sendMessage(msg tgbotapi.MessageConfig) {
	// Send new message
	_, err := s.botapi.Send(msg)
	if err != nil {
		s.logger.Error("Failed to send message",
			zap.Int64("chatID", msg.ChatID),
			zap.String("text", msg.Text),
			zap.Error(err))
		return
	}
}

func (s *StartHandler) showPrivacyPolicy(chatID int64) {
	msg := tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf(privacyPolicyMsg, "https://telegram.org/privacy-tpa"),
	)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboards.CreatePrivacyAgreementKeyboard()
	s.sendMessage(msg)
}

// showMainMenu displays the main menu with options for authenticated users
func (s *StartHandler) showMainMenu(ctx context.Context, chatID int64, phone string) {
	// Clear any previous order state while keeping essential user data
	if err := s.stateManager.ResetDialogState(ctx, chatID); err != nil {
		s.logger.Error("Failed to reset dialog state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		// Continue despite the error as it's not critical for showing the menu
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
		"🏠 *Главное меню*\n\n"+
			"Ваш контактный номер: %s\n"+
			"Выберите действие:",
		phone))

	msg.ReplyMarkup = keyboards.CreateMainMenuKeyboard()
	msg.ParseMode = "Markdown"

	s.sendMessage(msg)
}
