package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) HandleStart(ctx context.Context, chatID int64) {
    msg := tgbotapi.NewMessage(chatID, `Привет! 👋

    ⚠️ Прежде чем продолжить, вы должны согласиться с:
    1. Нашей Политикой конфиденциальности
    2. Telegram Bot Privacy Policy (TPA)
    
    Используя бота, вы подтверждаете согласие на обработку данных в соответствии с этими документами.`)
    
    b.ShowPrivacyPolicy(chatID)

    msg.ReplyMarkup = b.CreatePrivacyAgreementKeyboard()
    b.SendMessage(msg)
    
    if err := b.state.SetStep(ctx, chatID, StepPrivacyAgreement); err != nil {
        b.logger.Error("Failed to set privacy agreement state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }
}

func (b *Bot) ShowPrivacyPolicy(chatID int64) {
    
    policyText := `🔐 <b>Политика конфиденциальности</b>

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

    msg := tgbotapi.NewMessage(
        chatID, 
        fmt.Sprintf(policyText, "https://telegram.org/privacy-tpa"),
    )
    msg.ParseMode = "HTML"
    b.SendMessage(msg)
}

func (b *Bot) HandleError(ctx context.Context, chatID int64, errorMsg string) {
    step, err := b.state.GetStep(ctx, chatID)
    if err != nil {
        step = ""
    }

    var keyboard tgbotapi.ReplyKeyboardMarkup

    switch step {
    case StepDimensions:
        keyboard = b.CreateDimensionsKeyboard()
    case StepDateSelection:
        keyboard = b.CreateDateSelectionKeyboard()
    case StepServiceType:
        keyboard = b.CreateServiceTypeKeyboard()
    case StepManualDateInput:
        keyboard = tgbotapi.NewReplyKeyboard(
            tgbotapi.NewKeyboardButtonRow(
                tgbotapi.NewKeyboardButton("Назад"),
            ),
        )
    default:
        keyboard = tgbotapi.NewReplyKeyboard(
            tgbotapi.NewKeyboardButtonRow(
                tgbotapi.NewKeyboardButton("/start"),
            ),
        )
    }

    msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
    msg.ReplyMarkup = keyboard
    b.SendMessage(msg)
}

func (b *Bot) HandleCancel(ctx context.Context, chatID int64) {
    // Get current step to determine where to return
    currentStep, err := b.state.GetStep(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get current step",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        currentStep = "" // Default to start if cannot get step
    }

    var msg tgbotapi.MessageConfig
    var keyboard any

    switch currentStep {
    case StepDateSelection, StepManualDateInput, StepDateConfirmation:
        // Return to dimensions input
        msg = tgbotapi.NewMessage(chatID, "❌ Выбор даты отменен. Введите размеры снова:")
        keyboard = b.CreateDimensionsKeyboard()
        b.state.SetStep(ctx, chatID, StepDimensions)

    case StepDimensions:
        // Return to service type selection
        msg = tgbotapi.NewMessage(chatID, "❌ Ввод размеров отменен. Выберите тип услуги:")
        keyboard = b.CreateServiceTypeKeyboard()
        b.state.SetStep(ctx, chatID, StepServiceType)

    case CustomTextureInput:
        msg = tgbotapi.NewMessage(chatID, "❌ Ввод текстуры отменен. Выберите тип услуги:")
        keyboard = b.CreateServiceTypeKeyboard()
        b.state.SetStep(ctx, chatID, StepServiceType)    

    default:
        // Default cancellation - clear all and return to start
        if err := b.state.ClearState(ctx, chatID); err != nil {
            b.logger.Error("Failed to clear state on cancel",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
        }
        msg = tgbotapi.NewMessage(chatID, "❌ Действие отменено. Начните заново:")
        keyboard = tgbotapi.NewReplyKeyboard(
            tgbotapi.NewKeyboardButtonRow(
                tgbotapi.NewKeyboardButton("/start"),
            ),
        )
    }

    msg.ReplyMarkup = keyboard
    b.SendMessage(msg)
}


func (b *Bot) HandleDefault(ctx context.Context, chatID int64) {
	b.SendError(chatID, "Я не понимаю эту команду. Пожалуйста, используйте меню.")
}

func (b *Bot) HandleUnknownCommand(ctx context.Context, chatID int64) {
	b.SendError(chatID, "Неизвестная команда. Пожалуйста, используйте /start для начала работы.")
}

func (b *Bot) HandleHelp(ctx context.Context, chatID int64) {
	helpText := `Доступные команды:
	/start - Начать работу с ботом
	/help - Показать эту справку

	Если у вас возникли проблемы, свяжитесь с поддержкой.`
	
	msg := tgbotapi.NewMessage(chatID, helpText)
	b.SendMessage(msg)
}
