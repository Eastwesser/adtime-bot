package bot

import (
	"adtime-bot/internal/storage"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) handleStart(ctx context.Context, chatID int64) {
	text := `Привет! 👋

⚠️ Прежде чем продолжить, ознакомьтесь с нашей Политикой конфиденциальности.
Используя этого бота, вы соглашаетесь на обработку персональных данных.

Если всё ок — нажмите кнопку ниже 👇`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = b.CreatePrivacyAgreementKeyboard()
	b.sendMessage(msg)
	
	if err := b.state.SetStep(ctx, chatID, StepPrivacyAgreement); err != nil {
		b.logger.Error("Failed to set privacy agreement state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handleCommand(ctx context.Context, chatID int64, command string) {
	switch command {
	case "start":
		b.handleStart(ctx, chatID)
	case "help":
		b.handleHelp(ctx, chatID)
	default:
		b.handleUnknownCommand(ctx, chatID)
	}
}

func (b *Bot) handleDefault(ctx context.Context, chatID int64) {
	b.sendError(chatID, "Я не понимаю эту команду. Пожалуйста, используйте меню.")
}

func (b *Bot) handleUnknownCommand(ctx context.Context, chatID int64) {
	b.sendError(chatID, "Неизвестная команда. Пожалуйста, используйте /start для начала работы.")
}

func (b *Bot) handleHelp(ctx context.Context, chatID int64) {
	helpText := `Доступные команды:
/start - Начать работу с ботом
/help - Показать эту справку

Если у вас возникли проблемы, свяжитесь с поддержкой.`
	
	msg := tgbotapi.NewMessage(chatID, helpText)
	b.sendMessage(msg)
}

func (b *Bot) handlePrivacyAgreement(ctx context.Context, chatID int64, text string) {
	if text != "✅ Продолжить" {
		b.sendError(chatID, "Пожалуйста, нажмите кнопку \"✅ Продолжить\" чтобы согласиться с условиями")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Спасибо за согласие! Теперь вы можете оформить заказ.")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Оформить заказ"),
		),
	)

	b.sendMessage(msg)
	if err := b.state.SetStep(ctx, chatID, StepServiceSelection); err != nil {
		b.logger.Error("Failed to set service selection state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handleServiceSelection(ctx context.Context, chatID int64, text string) {
	if text != "✅ Оформить заказ" {
		b.sendError(chatID, "Пожалуйста, нажмите кнопку \"✅ Оформить заказ\" чтобы продолжить")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите тип услуги:")
	msg.ReplyMarkup = b.CreateServiceTypeKeyboard()
	b.sendMessage(msg)
	
	if err := b.state.SetStep(ctx, chatID, StepServiceType); err != nil {
		b.logger.Error("Failed to set service type state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handleServiceType(ctx context.Context, chatID int64, text string) {
	validServices := map[string]bool{
		"Печать наклеек": true,
		"Другая услуга":  true,
	}
	
	if !validServices[text] {
		b.sendError(chatID, "Пожалуйста, выберите один из предложенных вариантов")
		return
	}

	if err := b.state.SetService(ctx, chatID, text); err != nil {
		b.logger.Error("Failed to set service",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.sendError(chatID, "Ошибка при сохранении услуги")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Введите ширину и длину в сантиметрах через пробел (например: 30 40)\nМаксимальный размер: 80x50 см")
	b.sendMessage(msg)
	
	if err := b.state.SetStep(ctx, chatID, StepDimensions); err != nil {
		b.logger.Error("Failed to set dimensions state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handleDimensionsSize(ctx context.Context, chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendError(chatID, "Неверный формат. Введите ширину и длину через пробел")
		return
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil || width <= 0 || width > 80 {
		b.sendError(chatID, "Некорректная ширина. Допустимый диапазон: 1-80 см")
		return
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil || height <= 0 || height > 50 {
		b.sendError(chatID, "Некорректная длина. Допустимый диапазон: 1-50 см")
		return
	}

	if err := b.state.SetDimensions(ctx, chatID, width, height); err != nil {
		b.logger.Error("Failed to set dimensions",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.sendError(chatID, "Ошибка при сохранении размеров")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Когда вам удобно выполнить заказ?")
	msg.ReplyMarkup = b.CreateDateSelectionKeyboard()
	b.sendMessage(msg)
	
	if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
		b.logger.Error("Failed to set date selection state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handleDateSelection(ctx context.Context, chatID int64, text string) {
	var selectedDate string
	
	switch text {
	case "Сегодня":
		selectedDate = time.Now().Format("02.01.2006")
	case "Завтра":
		selectedDate = time.Now().Add(24 * time.Hour).Format("02.01.2006")
	case "Выбрать дату вручную":
		msg := tgbotapi.NewMessage(chatID, "Введите дату вручную в формате ДД.ММ.ГГГГ")
		b.sendMessage(msg)
		if err := b.state.SetStep(ctx, chatID, StepManualDateInput); err != nil {
			b.logger.Error("Failed to set manual date input state",
				zap.Int64("chat_id", chatID),
				zap.Error(err))
		}
		return
	default:
		b.sendError(chatID, "Пожалуйста, выберите один из предложенных вариантов")
		return
	}

	if err := b.state.SetDate(ctx, chatID, selectedDate); err != nil {
		b.logger.Error("Failed to set date",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.sendError(chatID, "Ошибка при установке даты")
		return
	}

	b.confirmDateSelection(ctx, chatID, selectedDate)
}

func (b *Bot) handleManualDateInput(ctx context.Context, chatID int64, text string) {
	_, err := time.Parse("02.01.2006", text)
	if err != nil {
		b.sendError(chatID, "Неверный формат даты. Пожалуйста, введите дату в формате ДД.ММ.ГГГГ")
		return
	}

	inputDate, _ := time.Parse("02.01.2006", text)
	if inputDate.Before(time.Now().Truncate(24 * time.Hour)) {
		b.sendError(chatID, "Пожалуйста, выберите дату в будущем")
		return
	}

	if err := b.state.SetDate(ctx, chatID, text); err != nil {
		b.logger.Error("Failed to set manual date",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.sendError(chatID, "Ошибка при сохранении даты")
		return
	}

	b.confirmDateSelection(ctx, chatID, text)
}

func (b *Bot) confirmDateSelection(ctx context.Context, chatID int64, date string) {
	days := b.calculateWorkingDays(date)

	text := fmt.Sprintf("Отлично! Вы выбрали дату: %s\nДо этой даты %d рабочих дней (без учёта выходных и праздников).", date, days)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔁 Сменить дату"),
			tgbotapi.NewKeyboardButton("✅ Подтвердить дату"),
		),
	)
	b.sendMessage(msg)
	if err := b.state.SetStep(ctx, chatID, StepDateConfirmation); err != nil {
		b.logger.Error("Failed to set date confirmation state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) calculateWorkingDays(date string) int {
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

func (b *Bot) handleDateConfirmation(ctx context.Context, chatID int64, text string) {
	if text != "✅ Подтвердить дату" {
		b.sendError(chatID, "Пожалуйста, нажмите кнопку подтверждения")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Как вам удобно предоставить контактные данные?")
	msg.ReplyMarkup = b.CreatePhoneInputKeyboard()
	b.sendMessage(msg)
	
	if err := b.state.SetStep(ctx, chatID, StepContactMethod); err != nil {
		b.logger.Error("Failed to set contact method state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handleContactMethod(ctx context.Context, chatID int64, text string) {
	if text == "Ввести вручную" {
		msg := tgbotapi.NewMessage(chatID, "Введите ваш номер телефона в формате +79123456789 или 89123456789:")
		b.sendMessage(msg)
		
		if err := b.state.SetStep(ctx, chatID, StepPhoneNumber); err != nil {
			b.logger.Error("Failed to set phone number state",
				zap.Int64("chat_id", chatID),
				zap.Error(err))
		}
	}
}

func (b *Bot) handlePhoneNumber(ctx context.Context, chatID int64, text string) {
    normalized := NormalizePhoneNumber(text)
    
    if !IsValidPhoneNumber(normalized) {
        b.sendError(chatID, "Пожалуйста, введите реальный номер телефона с кодом страны (например, +79161234567 или 89123456789)")
        return
    }

    b.logger.Debug("Phone number validation",
        zap.String("input", text),
        zap.String("normalized", normalized),
        zap.Bool("is_valid", IsValidPhoneNumber(normalized)))

    // Create and save the order
    orderID, err := b.createOrder(ctx, chatID, normalized)
    if err != nil {
        b.logger.Error("Failed to create order",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при оформлении заказа")
        return
    }

    // Clear user state
    if err := b.state.ClearState(ctx, chatID); err != nil {
        b.logger.Error("Failed to clear user state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }

    // Send final confirmation
    b.sendMessage(tgbotapi.NewMessage(chatID,
        "✅ Ваш заказ успешно оформлен!\nНомер заказа: #"+strconv.FormatInt(orderID, 10)))
}

func (b *Bot) createOrder(ctx context.Context, chatID int64, phone string) (int64, error) {
    state, err := b.state.GetFullState(ctx, chatID)
    if err != nil {
        return 0, fmt.Errorf("failed to get order state: %w", err)
    }

    width, height, err := b.state.GetDimensions(ctx, chatID)
    if err != nil {
        return 0, fmt.Errorf("failed to get dimensions: %w", err)
    }

    // Get or set default texture
    var texture *storage.Texture
    if state.TextureID != "" {
        texture, err = b.storage.GetTextureByID(ctx, state.TextureID)
        if err != nil {
            return 0, fmt.Errorf("failed to get texture: %w", err)
        }
    } else {
        texture = &storage.Texture{
            ID:          "11111111-1111-1111-1111-111111111111",
            Name:        "Стандартная текстура",
            PricePerDM2: 10.0,
        }
    }

    // Create pricing config from bot's configuration
    pricingConfig := PricingConfig{
        LeatherPricePerDM2:    b.cfg.Pricing.LeatherPricePerDM2,
        ProcessingCostPerDM2:  b.cfg.Pricing.ProcessingCostPerDM2,
        PaymentCommissionRate: b.cfg.Pricing.PaymentCommissionRate,
        SalesTaxRate:          b.cfg.Pricing.SalesTaxRate,
        MarkupMultiplier:      b.cfg.Pricing.MarkupMultiplier,
    }

    // Calculate pricing
    priceDetails := CalculatePrice(width, height, pricingConfig)

    order := storage.Order{
        UserID:      chatID,
        WidthCM:     width,
        HeightCM:    height,
        TextureID:   texture.ID,
        TextureName: texture.Name,
        Price:       priceDetails["final_price"],
        LeatherCost: priceDetails["leather_cost"],
        ProcessCost: priceDetails["processing_cost"],
        TotalCost:   priceDetails["total_cost"],
        Commission:  priceDetails["commission"],
        Tax:         priceDetails["tax"],
        NetRevenue:  priceDetails["net_revenue"],
        Profit:      priceDetails["profit"],
        Contact:     phone,
        Status:      "new",
        CreatedAt:   time.Now(),
    }

    // Save to database
    orderID, err := b.storage.SaveOrder(ctx, order)
    if err != nil {
        return 0, fmt.Errorf("failed to save order: %w", err)
    }

    // Export to Excel
    if err := b.storage.ExportOrderToExcel(ctx, order); err != nil {
        b.logger.Error("Failed to export order to Excel",
            zap.Int64("order_id", orderID),
            zap.Error(err))
    }

    // Send order details
    msgText := fmt.Sprintf(
        "✅ Заказ #%d оформлен!\n%s\n\nКонтакт: %s",
        orderID,
        FormatPriceBreakdown(width, height, priceDetails),
        FormatPhoneNumber(phone),
    )
    b.sendMessage(tgbotapi.NewMessage(chatID, msgText))

    return orderID, nil
}

func (b *Bot) handleTextureSelection(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
    
	// Parse texture ID from callback data
	parts := strings.Split(callback.Data, ":")
	if len(parts) != 2 {
		b.sendError(chatID, "Неверный формат выбора текстуры")
		return
	}
	textureID := parts[1]

	// Get dimensions from state
	width, height, err := b.state.GetDimensions(ctx, chatID)
	if err != nil {
		b.logger.Error("Failed to get dimensions",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.sendError(chatID, "Ошибка при получении размеров")
		return
	}

	// Get texture from storage
	texture, err := b.storage.GetTextureByID(ctx, textureID)
	if err != nil {
		b.logger.Error("Failed to get texture",
			zap.String("texture_id", textureID),
			zap.Error(err))
		b.sendError(chatID, "Не удалось получить информацию о текстуре")
		return
	}

	// Calculate price
	price := CalculatePrice(width, height, texture.PricePerDM2)

	// Save texture selection to state
	if err := b.state.SetTexture(ctx, chatID, textureID, price); err != nil {
		b.logger.Error("Failed to set texture",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.sendError(chatID, "Ошибка при сохранении текстуры")
		return
	}

	// Send confirmation message with price
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
		"Вы выбрали текстуру: %s\nЦена: %.2f руб\n\nКогда вам удобно выполнить заказ?",
		texture.Name,
		price,
	))
	msg.ReplyMarkup = b.CreateDateSelectionKeyboard()
	b.sendMessage(msg)

	// Update user step
	if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
		b.logger.Error("Failed to set date selection state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}

	// Delete the original message with texture options
	delMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	if _, err := b.bot.Send(delMsg); err != nil {
		b.logger.Warn("Failed to delete message",
			zap.Int("message_id", callback.Message.MessageID),
			zap.Error(err))
	}
}
