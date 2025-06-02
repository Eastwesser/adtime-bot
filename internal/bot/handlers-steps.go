package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) HandlePrivacyAgreement(ctx context.Context, chatID int64, text string) {
	if text != "✅ Продолжить" {
        b.SendError(chatID, "Пожалуйста, нажмите кнопку \"✅ Продолжить\" чтобы согласиться с условиями")
        return
    }

    // Get username for notification
    chat, err := b.bot.GetChat(tgbotapi.ChatInfoConfig{
        ChatConfig: tgbotapi.ChatConfig{
            ChatID: chatID,
        },
    })
    username := ""
    if err == nil && chat.UserName != "" {
        username = chat.UserName
        // Send notification to channel
        b.NotifyPrivacyAgreement(ctx, username)
    }

	msg := tgbotapi.NewMessage(chatID, "Спасибо за согласие! Теперь вы можете оформить заказ.")
    msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("✅ Оформить заказ"),
        ),
    )

	b.SendMessage(msg)
    if err := b.state.SetStep(ctx, chatID, StepServiceSelection); err != nil {
        b.logger.Error("Failed to set service selection state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }
}

func (b *Bot) HandleServiceSelection(ctx context.Context, chatID int64, text string) {
	if text != "✅ Оформить заказ" {
		b.SendError(chatID, "Пожалуйста, нажмите кнопку \"✅ Оформить заказ\" чтобы продолжить")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите тип услуги:")
	msg.ReplyMarkup = b.CreateServiceTypeKeyboard()
	b.SendMessage(msg)
	
	if err := b.state.SetStep(ctx, chatID, StepServiceType); err != nil {
		b.logger.Error("Failed to set service type state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) HandleServiceType(ctx context.Context, chatID int64, text string) {
    validServices := map[string]bool{
        "Натуральная кожа": true,
        "Искусственная кожа": true,
        "Замша": true,
        "Другая текстура": true,
    }

	if !validServices[text] {
        if text == "❌ Отмена" {
            b.HandleCancel(ctx, chatID)
            return
        }
        b.SendError(chatID, "Пожалуйста, выберите один из предложенных вариантов")
        return
    }

	if err := b.state.SetService(ctx, chatID, text); err != nil {
        b.logger.Error("Failed to set service",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.SendError(chatID, "Ошибка при сохранении услуги")
        return
    }

    // Особый случай для "Другой текстуры"
    if text == "Другая текстура" {
        msg := tgbotapi.NewMessage(chatID, "Введите желаемую текстуру:")
        msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
            tgbotapi.NewKeyboardButtonRow(
                tgbotapi.NewKeyboardButton("Назад"),
            ),
        )

        b.SendMessage(msg)
        if err := b.state.SetStep(ctx, chatID, CustomTextureInput); err != nil {
            b.logger.Error("Failed to set custom texture input state",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
        }
        return
    }

    // Для остальных случаев сразу запрашиваем размеры
    msg := tgbotapi.NewMessage(chatID, "Введите ширину и длину в сантиметрах через пробел (например: 30 40)\nМаксимальный размер: 80x50 см")
    msg.ReplyMarkup = b.CreateDimensionsKeyboard()
    b.SendMessage(msg)
    
    if err := b.state.SetStep(ctx, chatID, StepDimensions); err != nil {
        b.logger.Error("Failed to set dimensions state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }
}

func (b *Bot) HandleCustomTextureInput(ctx context.Context, chatID int64, text string) {
    if text == "Назад" {
        msg := tgbotapi.NewMessage(chatID, "Выберите тип услуги:")
        msg.ReplyMarkup = b.CreateServiceTypeKeyboard()
        b.SendMessage(msg)
        b.state.SetStep(ctx, chatID, StepServiceType)
        return
    }

    if len(text) < 3 {
        b.HandleError(ctx, chatID, "Пожалуйста, введите название текстуры (минимум 3 символа)")
        return
    }

    // Сохраняем кастомную текстуру
    if err := b.state.SetServiceType(ctx, chatID, text); err != nil {
        b.logger.Error("Failed to set custom texture",
            zap.Int64("chatID", chatID),
            zap.Error(err))
        b.HandleError(ctx, chatID, "Ошибка при сохранении текстуры")
        return
    }

    // Переходим к вводу размеров
    msg := tgbotapi.NewMessage(chatID, "Введите ширину и длину в сантиметрах через пробел (например: 30 40)\nМаксимальный размер: 80x50 см")
    msg.ReplyMarkup = b.CreateDimensionsKeyboard()
    b.SendMessage(msg)
    
    if err := b.state.SetStep(ctx, chatID, StepDimensions); err != nil {
        b.logger.Error("Failed to set dimensions state",
            zap.Int64("chatID", chatID),
            zap.Error(err))
    }
}

func (b *Bot) HandleDimensionsSize(ctx context.Context, chatID int64, text string) {
	// Handle back button first
    if text == "Назад" {
        b.HandleCancel(ctx, chatID)
        return
    }
    
    // Process the dimensions input
    parts := strings.Split(text, " ")
    if len(parts) != 2 {
        b.HandleError(ctx, chatID, "Неверный формат. Введите ширину и длину через пробел (например: 30 40)")
        return
    }

	width, err := strconv.Atoi(parts[0])
    if err != nil || width <= 0 || width > 80 {
        b.HandleError(ctx, chatID, "Некорректная ширина. Допустимый диапазон: 1-80 см")
        return
    }

	height, err := strconv.Atoi(parts[1])
    if err != nil || height <= 0 || height > 50 {
        b.HandleError(ctx, chatID, "Некорректная длина. Допустимый диапазон: 1-50 см")
        return
    }

	if err := b.state.SetDimensions(ctx, chatID, width, height); err != nil {
        b.logger.Error("Failed to set dimensions",
            zap.Int64("chatID", chatID),
            zap.Error(err))
        b.HandleError(ctx, chatID, "Ошибка при сохранении размеров")
        return
    }

	msg := tgbotapi.NewMessage(chatID, "Когда вам удобно выполнить заказ?")
    msg.ReplyMarkup = b.CreateDateSelectionKeyboard()
    b.SendMessage(msg)
	
	if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
        b.logger.Error("Failed to set date selection state",
            zap.Int64("chatID", chatID),
            zap.Error(err))
    }
}

func (b *Bot) HandleDateSelection(ctx context.Context, chatID int64, text string) {
	var selectedDate string
	
	switch text {
	case "Сегодня":
		selectedDate = time.Now().Format("02.01.2006")
	case "Завтра":
		selectedDate = time.Now().Add(24 * time.Hour).Format("02.01.2006")
	case "Выбрать дату вручную":
		msg := tgbotapi.NewMessage(chatID, "Введите дату вручную в формате ДД.ММ.ГГГГ")
		b.SendMessage(msg)
		if err := b.state.SetStep(ctx, chatID, StepManualDateInput); err != nil {
			b.logger.Error("Failed to set manual date input state",
				zap.Int64("chat_id", chatID),
				zap.Error(err))
		}
		return
	default:
		b.SendError(chatID, "Пожалуйста, выберите один из предложенных вариантов")
		return
	}

	if err := b.state.SetDate(ctx, chatID, selectedDate); err != nil {
		b.logger.Error("Failed to set date",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.SendError(chatID, "Ошибка при установке даты")
		return
	}

	b.ConfirmDateSelection(ctx, chatID, selectedDate)
}

func (b *Bot) HandleManualDateInput(ctx context.Context, chatID int64, text string) {
	// Автокоррекция года
    if len(text) == 8 { // формат ДД.ММ.ГГ
        text = text[:6] + "20" + text[6:]
    }
    
    _, err := time.Parse("02.01.2006", text)
    if err != nil {
        b.SendError(chatID, "Неверный формат даты. Пожалуйста, введите дату в формате ДД.ММ.ГГГГ")
        return
    }

	inputDate, _ := time.Parse("02.01.2006", text)
	if inputDate.Before(time.Now().Truncate(24 * time.Hour)) {
		b.SendError(chatID, "Пожалуйста, выберите дату в будущем")
		return
	}

	if err := b.state.SetDate(ctx, chatID, text); err != nil {
		b.logger.Error("Failed to set manual date",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.SendError(chatID, "Ошибка при сохранении даты")
		return
	}

	b.ConfirmDateSelection(ctx, chatID, text)
}

func (b *Bot) ConfirmDateSelection(ctx context.Context, chatID int64, date string) {
	days := b.CalculateWorkingDays(date)

	text := fmt.Sprintf("Отлично! Вы выбрали дату: %s\nДо этой даты %d рабочих дней (без учёта выходных и праздников).", date, days)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔁 Сменить дату"),
			tgbotapi.NewKeyboardButton("✅ Подтвердить дату"),
		),
	)
	b.SendMessage(msg)
	if err := b.state.SetStep(ctx, chatID, StepDateConfirmation); err != nil {
		b.logger.Error("Failed to set date confirmation state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) CalculateWorkingDays(date string) int {
	targetDate, _ := time.Parse("02.01.2006", date)
	now := time.Now()

	days := 0
	for d := now; d.Before(targetDate); d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			days++
		}
	}
	return days
}

func (b *Bot) HandleDateConfirmation(ctx context.Context, chatID int64, text string) {
    switch text {
    case "✅ Подтвердить дату":
        msg := tgbotapi.NewMessage(chatID, "Как вам удобно предоставить контактные данные?")
        msg.ReplyMarkup = b.CreatePhoneInputKeyboard()
        b.SendMessage(msg)
        
        if err := b.state.SetStep(ctx, chatID, StepContactMethod); err != nil {
            b.logger.Error("Failed to set contact method state",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
        }
        
    case "🔁 Сменить дату":
        // Go back to date selection
        msg := tgbotapi.NewMessage(chatID, "Когда вам удобно выполнить заказ?")
        msg.ReplyMarkup = b.CreateDateSelectionKeyboard()
        b.SendMessage(msg)
        
        if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
            b.logger.Error("Failed to set date selection state",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
        }
        
    default:
        b.SendError(chatID, "Пожалуйста, используйте кнопки для продолжения")
    }
}

func (b *Bot) HandleContactMethod(ctx context.Context, chatID int64, text string) {
    switch text {
    case "📱 Отправить контакт":
        msg := tgbotapi.NewMessage(chatID, "Пожалуйста, нажмите кнопку 'Поделиться контактом'")
        msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
            tgbotapi.NewKeyboardButtonRow(
                tgbotapi.NewKeyboardButtonContact("📱 Поделиться контактом"),
            ),
            tgbotapi.NewKeyboardButtonRow(
                tgbotapi.NewKeyboardButton("Назад"),
            ),
        )
        b.SendMessage(msg)
        
    case "Ввести вручную":
        msg := tgbotapi.NewMessage(chatID, "Введите ваш номер телефона в формате +79123456789 или 89123456789:")
        b.SendMessage(msg)
        b.state.SetStep(ctx, chatID, StepPhoneNumber)
        
    case "Назад":
        // Go back to previous step
        b.HandleDateConfirmation(ctx, chatID, "🔁 Сменить дату")
        
    default:
        b.SendError(chatID, "Пожалуйста, выберите один из вариантов")
    }
}

func (b *Bot) HandlePhoneNumber(ctx context.Context, chatID int64, text string) {

    normalized := NormalizePhoneNumber(text)
    
    if !IsValidPhoneNumber(normalized) {
        b.SendError(chatID, "Пожалуйста, введите реальный номер телефона с кодом страны (например, +79161234567 или 89123456789)")
        return
    }

    b.logger.Debug("Phone number validation",
        zap.String("input", text),
        zap.String("normalized", normalized),
        zap.Bool("is_valid", IsValidPhoneNumber(normalized)))

    // Create and save the order
    orderID, err := b.CreateOrder(ctx, chatID, normalized)
    if err != nil {
        b.logger.Error("Failed to create order",
            zap.Int64("chat_id", chatID),
            zap.String("phone", normalized),
            zap.Error(err))
        b.SendError(chatID, "Ошибка при оформлении заказа. Пожалуйста, попробуйте позже.")
        return
    }

    // Clear user state
    if err := b.state.ClearState(ctx, chatID); err != nil {
        b.logger.Error("Failed to clear user state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }

    // Send final confirmation
    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("✅ Ваш заказ успешно оформлен!\nНомер заказа: #%d\n\nС вами свяжутся в ближайшее время.", orderID))
    b.SendMessage(msg)
}
