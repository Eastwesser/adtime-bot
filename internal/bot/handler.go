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
	b.sendMessage(tgbotapi.NewMessage(chatID, helpText))
}

func (b *Bot) handleStart(ctx context.Context, chatID int64) {
	text := `Привет! 👋

	⚠️ Прежде чем продолжить, ознакомьтесь с нашей Политикой конфиденциальности.
	Используя этого бота, вы соглашаетесь на обработку персональных данных.

	Если всё ок — нажмите кнопку ниже 👇`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Продолжить"),
		),
	)

	b.sendMessage(msg)
	if err := b.state.SetStep(ctx, chatID, StepPrivacyAgreement); err != nil {
		b.logger.Error("Failed to set privacy agreement state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
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

	msg := tgbotapi.NewMessage(chatID, "Какую услугу вы хотите заказать?")
	b.sendMessage(msg)
	if err := b.state.SetStep(ctx, chatID, StepServiceInput); err != nil {
		b.logger.Error("Failed to set service input state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handleServiceInput(ctx context.Context, chatID int64, text string) {
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

	b.showTextures(ctx, chatID)
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

    // Get texture price from API
    texturePrice, err := b.api.GetTexturePrice(ctx, textureID)
    if err != nil {
        b.logger.Error("Failed to get texture price",
            zap.String("texture_id", textureID),
            zap.Error(err))
        b.sendError(chatID, "Не удалось получить цену текстуры")
        return
    }

    // Calculate total price
    price := CalculatePrice(width, height, texturePrice)

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
        "Вы выбрали текстуру. Итоговая цена: %.2f руб\n\nКогда вам удобно выполнить заказ?",
        price,
    ))
    msg.ReplyMarkup = b.CreateDateSelectionKeyboard()
    b.sendMessage(msg)

    // Update user step
    if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
        b.logger.Error("Failed to set waiting state",
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

func (b *Bot) handleDateSelection(ctx context.Context, chatID int64, text string) {
	switch text {
	case "Сегодня":
		today := time.Now().Format("02.01.2006")
		if err := b.state.SetDate(ctx, chatID, today); err != nil {
			b.logger.Error("Failed to set today as date",
				zap.Int64("chat_id", chatID),
				zap.Error(err))
			b.sendError(chatID, "Ошибка при установке даты")
			return
		}
		b.confirmDateSelection(ctx, chatID, today)
	case "Завтра":
		tomorrow := time.Now().Add(24 * time.Hour).Format("02.01.2006")
		if err := b.state.SetDate(ctx, chatID, tomorrow); err != nil {
			b.logger.Error("Failed to set tomorrow as date",
				zap.Int64("chat_id", chatID),
				zap.Error(err))
			b.sendError(chatID, "Ошибка при установке даты")
			return
		}
		b.confirmDateSelection(ctx, chatID, tomorrow)
	case "Выбрать дату вручную":
		msg := tgbotapi.NewMessage(chatID, "Введите дату вручную в формате ДД.ММ.ГГГГ")
		b.sendMessage(msg)
		if err := b.state.SetStep(ctx, chatID, StepManualDateInput); err != nil {
			b.logger.Error("Failed to set manual date input state",
				zap.Int64("chat_id", chatID),
				zap.Error(err))
		}
	default:
		b.sendError(chatID, "Пожалуйста, выберите один из предложенных вариантов")
	}
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

	msg := tgbotapi.NewMessage(chatID, "Введите ваш номер телефона для связи:")
	b.sendMessage(msg)
	if err := b.state.SetStep(ctx, chatID, StepPhoneNumber); err != nil {
		b.logger.Error("Failed to set phone number state",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}
}

func (b *Bot) handlePhoneNumber(ctx context.Context, chatID int64, text string) {
    // Validate phone number format
    if !IsValidPhoneNumber(text) {
        b.sendError(chatID, "Пожалуйста, введите реальный номер телефона с кодом страны (например, +79161234567)")
        return
    }

    // Get order details from state
    _, err := b.state.GetFullState(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get order state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при обработке заказа")
        return
    }

    // Get dimensions
    width, height, err := b.state.GetDimensions(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get dimensions",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при получении размеров")
        return
    }

    // Get selected texture
    textureID, err := b.state.GetTextureID(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get texture ID from state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при получении текстуры")
        return
    }

    texture, err := b.storage.GetTextureByID(ctx, textureID)
    if err != nil {
        b.logger.Error("Failed to get texture",
            zap.String("texture_id", textureID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при получении текстуры")
        return
    }

    // Calculate price and create order
    price := CalculatePrice(width, height, texture.PricePerDM2)
    order := storage.Order{
        UserID:      chatID,
        WidthCM:     width,
        HeightCM:    height,
        TextureID:   texture.ID,
        TextureName: texture.Name,
        PricePerDM2: texture.PricePerDM2,
        TotalPrice:  price,
        Contact:     text,
        Status:      "new",
        CreatedAt:   time.Now(),
    }

    // Save order to database
    orderID, err := b.storage.SaveOrder(ctx, order)
    if err != nil {
        b.logger.Error("Failed to save order",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при сохранении заказа")
        return
    }
    order.ID = orderID

    // Send confirmation to user
    b.sendMessage(tgbotapi.NewMessage(chatID,
        "✅ Ваш заказ успешно оформлен!\n\nМы свяжемся с вами в ближайшее время."))

    // Notify admin
    adminMsg := FormatOrderNotification(order)
    b.sendAdminNotification(ctx, adminMsg)

    // Export to Excel (with error handling)
    if err := b.storage.ExportOrderToExcel(ctx, order); err != nil {
        b.logger.Error("Failed to export order to Excel",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendAdminNotification(ctx, 
            fmt.Sprintf("⚠️ Failed to export order #%d to Excel: %v", order.ID, err))
    }

    // Clear user state
    if err := b.state.ClearState(ctx, chatID); err != nil {
        b.logger.Error("Failed to clear user state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }
}

func (b *Bot) showTextures(ctx context.Context, chatID int64) {
    width, height, err := b.state.GetDimensions(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get dimensions",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при получении размеров")
        return
    }

    // Validate max dimensions
    if width > 80 || height > 50 {
        b.sendError(chatID, "Максимальный размер: 80x50 см")
        return
    }

    textures, err := b.storage.GetAvailableTextures(ctx)
    if err != nil {
        b.logger.Error("Failed to get textures",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Не удалось загрузить варианты текстуры")
        return
    }

    var buttons []tgbotapi.InlineKeyboardButton
    for _, texture := range textures {
        price := CalculatePrice(width, height, texture.PricePerDM2)
        btn := tgbotapi.NewInlineKeyboardButtonData(
            fmt.Sprintf("%s - %.2f руб (%.2f₽/дм²)", texture.Name, price, texture.PricePerDM2),
            fmt.Sprintf("texture:%s", texture.ID),
        )
        buttons = append(buttons, btn)
    }

    if len(buttons) == 0 {
        b.sendError(chatID, "Нет доступных текстур")
        return
    }

    // Send texture image if available
    if textures[0].ImageURL != "" {
        photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(textures[0].ImageURL))
        photo.Caption = "Образцы доступных текстур:"
        if _, err := b.bot.Send(photo); err != nil {
            b.logger.Error("Failed to send photo",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
        }
    }

    msg := tgbotapi.NewMessage(chatID, "Выберите текстуру:")
    msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(buttons...),
    )
    b.sendMessage(msg)
}
