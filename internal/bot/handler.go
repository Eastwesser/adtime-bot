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

func (b *Bot) handleCancel(ctx context.Context, chatID int64) {
    // Get current step to determine where to return
    currentStep, err := b.state.GetStep(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get current step",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        currentStep = "" // Default to start if cannot get step
    }

    var msg tgbotapi.MessageConfig
    var keyboard interface{}

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
    b.sendMessage(msg)
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

func (b *Bot) handleAdminCommand(ctx context.Context, chatID int64, cmd string, args []string) {
	if !b.isAdmin(chatID) {
		return
	}

	switch cmd {
	case "export":
		if len(args) == 0 {
			b.handleExportAllOrders(ctx, chatID)
		} else {
			orderID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				b.sendError(chatID, "Неверный формат ID заказа")
				return
			}
			b.handleExportSingleOrder(ctx, chatID, orderID)
		}
	case "stats":
		b.handleOrderStats(ctx, chatID)
	case "status":
		if len(args) < 2 {
			b.sendError(chatID, "Использование: /status <ID_заказа> <новый_статус>")
			return
		}
		b.handleStatusUpdate(ctx, chatID, args[0], args[1])
	default:
		b.sendError(chatID, "Неизвестная команда администратора")
	}
}

func (b *Bot) handleStatusUpdate(ctx context.Context, chatID int64, orderIDStr string, newStatus string) {
    orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
    if err != nil {
        b.sendError(chatID, "Неверный формат ID заказа")
        return
    }

    // Validate status
    validStatuses := map[string]bool{
        "new":        true,
        "processing": true,
        "completed":  true,
        "cancelled":  true,
    }
    if !validStatuses[newStatus] {
        b.sendError(chatID, "Недопустимый статус. Допустимые значения: new, processing, completed, cancelled")
        return
    }

    // Update status in database
    err = b.storage.UpdateOrderStatus(ctx, orderID, newStatus)
    if err != nil {
        b.logger.Error("Failed to update order status",
            zap.Int64("order_id", orderID),
            zap.String("status", newStatus),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при обновлении статуса")
        return
    }

    // Notify admin
    b.sendMessage(tgbotapi.NewMessage(chatID, fmt.Sprintf(
        "✅ Статус заказа #%d изменён на: %s",
        orderID,
        map[string]string{
            "new":        "Новый",
            "processing": "В обработке",
            "completed":  "Завершён",
            "cancelled":  "Отменён",
        }[newStatus],
    )))

    // Notify user if possible
    order, err := b.storage.GetOrderByID(ctx, orderID)
    if err == nil {
        userMsg := tgbotapi.NewMessage(order.UserID, fmt.Sprintf(
            "ℹ️ Статус вашего заказа #%d изменён на: %s",
            orderID,
            map[string]string{
                "new":        "Новый",
                "processing": "В обработке",
                "completed":  "Завершён",
                "cancelled":  "Отменён",
            }[newStatus],
        ))
        if _, err := b.bot.Send(userMsg); err != nil {
            b.logger.Warn("Failed to notify user about status change",
                zap.Int64("user_id", order.UserID),
                zap.Error(err))
        }
    }
}

// handleOrderStats shows statistics about orders
func (b *Bot) handleOrderStats(ctx context.Context, chatID int64) {
    // Get statistics from storage
    stats, err := b.storage.GetOrderStatistics(ctx)
    if err != nil {
        b.logger.Error("Failed to get order statistics", zap.Error(err))
        b.sendError(chatID, "Ошибка при получении статистики")
        return
    }

    // Format message
    msgText := fmt.Sprintf(
        "📊 *Статистика заказов*\n\n"+
            "📌 Всего заказов: %d\n"+
            "💰 Общая сумма: %.2f ₽\n"+
            "📅 За сегодня: %d (%.2f ₽)\n"+
            "📅 За неделю: %d (%.2f ₽)\n"+
            "📅 За месяц: %d (%.2f ₽)\n\n"+
            "📌 По статусам:\n"+
            "🆕 Новые: %d\n"+
            "🔄 В обработке: %d\n"+
            "✅ Завершённые: %d\n"+
            "❌ Отменённые: %d",
        stats.TotalOrders,
        stats.TotalRevenue,
        stats.TodayOrders, stats.TodayRevenue,
        stats.WeekOrders, stats.WeekRevenue,
        stats.MonthOrders, stats.MonthRevenue,
        stats.StatusCounts["new"],
        stats.StatusCounts["processing"],
        stats.StatusCounts["completed"],
        stats.StatusCounts["cancelled"],
    )

    msg := tgbotapi.NewMessage(chatID, msgText)
    msg.ParseMode = "Markdown"
    b.sendMessage(msg)
}

func (b *Bot) handleExportAllOrders(ctx context.Context, chatID int64) {
	filename := fmt.Sprintf("orders_report_%s", time.Now().Format("20060102"))
	if err := b.storage.ExportAllOrdersToExcel(ctx, filename); err != nil {
		b.logger.Error("Failed to export all orders", zap.Error(err))
		b.sendError(chatID, "Failed to export orders")
		return
	}

	filepath := fmt.Sprintf("reports/%s.xlsx", filename)
	msg := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filepath))
	msg.Caption = "📊 All orders export"

	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("Failed to send Excel file", zap.Error(err))
		b.sendError(chatID, "Failed to send exported file")
	}
}

func (b *Bot) handleExportSingleOrder(ctx context.Context, chatID int64, orderID int64) {
	order, err := b.storage.GetOrderByID(ctx, orderID)
	if err != nil {
		b.logger.Error("Failed to get order",
			zap.Int64("order_id", orderID),
			zap.Error(err))
		b.sendError(chatID, "Order not found")
		return
	}

	filepath, err := b.storage.ExportOrderToExcel(ctx, *order)
	if err != nil {
		b.logger.Error("Failed to export order",
			zap.Int64("order_id", orderID),
			zap.Error(err))
		b.sendError(chatID, "Failed to export order")
		return
	}

	msg := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filepath))
	msg.Caption = fmt.Sprintf("📊 Order #%d export", orderID)

	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("Failed to send Excel file", zap.Error(err))
		b.sendError(chatID, "Failed to send exported file")
	}
}

func (b *Bot) handlePrivacyAgreement(ctx context.Context, chatID int64, text string) {
	if text != "✅ Продолжить" {
        b.sendError(chatID, "Пожалуйста, нажмите кнопку \"✅ Продолжить\" чтобы согласиться с условиями")
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
        b.notifyPrivacyAgreement(ctx, username)
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
        "Натуральная кожа": true,
        "Искусственная кожа": true,
        "Замша": true,
        "Другая текстура": true,
    }

	if !validServices[text] {
        if text == "❌ Отмена" {
            b.handleCancel(ctx, chatID)
            return
        }
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

	// textures, err := b.storage.GetAvailableTextures(ctx)
    // if err != nil {
    //     b.logger.Error("Failed to get textures",
    //         zap.Int64("chat_id", chatID),
    //         zap.Error(err))
    //     b.sendError(chatID, "Ошибка при получении текстур")
    //     return
    // }

	// // For all service types except "Другая текстура", ask for dimensions first
    // if text != "Другая текстура" {
    //     msg := tgbotapi.NewMessage(chatID, "Введите ширину и длину в сантиметрах через пробел (например: 30 40)\nМаксимальный размер: 80x50 см")
    //     b.sendMessage(msg)
        
    //     if err := b.state.SetStep(ctx, chatID, StepDimensions); err != nil {
    //         b.logger.Error("Failed to set dimensions state",
    //             zap.Int64("chat_id", chatID),
    //             zap.Error(err))
    //     }
    //     return
    // }

    // // Show texture selection keyboard
    // msg := tgbotapi.NewMessage(chatID, "Выберите текстуру:")
    // msg.ReplyMarkup = b.CreateTextureSelectionKeyboard(textures)
    // b.sendMessage(msg)
    
    // if err := b.state.SetStep(ctx, chatID, StepTextureSelection); err != nil {
    //     b.logger.Error("Failed to set texture selection state",
    //         zap.Int64("chat_id", chatID),
    //         zap.Error(err))
    // }
    // Always ask for dimensions first, regardless of service type
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
    switch text {
    case "✅ Подтвердить дату":
        msg := tgbotapi.NewMessage(chatID, "Как вам удобно предоставить контактные данные?")
        msg.ReplyMarkup = b.CreatePhoneInputKeyboard()
        b.sendMessage(msg)
        
        if err := b.state.SetStep(ctx, chatID, StepContactMethod); err != nil {
            b.logger.Error("Failed to set contact method state",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
        }
        
    case "🔁 Сменить дату":
        // Go back to date selection
        msg := tgbotapi.NewMessage(chatID, "Когда вам удобно выполнить заказ?")
        msg.ReplyMarkup = b.CreateDateSelectionKeyboard()
        b.sendMessage(msg)
        
        if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
            b.logger.Error("Failed to set date selection state",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
        }
        
    default:
        b.sendError(chatID, "Пожалуйста, используйте кнопки для продолжения")
    }
}

func (b *Bot) handleContactMethod(ctx context.Context, chatID int64, text string) {
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
        b.sendMessage(msg)
        
    case "Ввести вручную":
        msg := tgbotapi.NewMessage(chatID, "Введите ваш номер телефона в формате +79123456789 или 89123456789:")
        b.sendMessage(msg)
        b.state.SetStep(ctx, chatID, StepPhoneNumber)
        
    case "Назад":
        // Go back to previous step
        b.handleDateConfirmation(ctx, chatID, "🔁 Сменить дату")
        
    default:
        b.sendError(chatID, "Пожалуйста, выберите один из вариантов")
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
            zap.String("phone", normalized),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при оформлении заказа. Пожалуйста, попробуйте позже.")
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
    b.sendMessage(msg)
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

    // Create temporary pricing config using texture price
    pricingConfig := PricingConfig{
        LeatherPricePerDM2:    texture.PricePerDM2,
        ProcessingCostPerDM2:  b.cfg.Pricing.ProcessingCostPerDM2,
        PaymentCommissionRate: b.cfg.Pricing.PaymentCommissionRate,
        SalesTaxRate:          b.cfg.Pricing.SalesTaxRate,
        MarkupMultiplier:      b.cfg.Pricing.MarkupMultiplier,
    }

    // Calculate full price details
    priceDetails := CalculatePrice(width, height, pricingConfig)

    // Save texture selection to state
    if err := b.state.SetTexture(ctx, chatID, textureID, priceDetails["final_price"]); err != nil {
        b.logger.Error("Failed to set texture",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при сохранении текстуры")
        return
    }

    // Send confirmation message with price breakdown
    msg := tgbotapi.NewMessage(
        chatID, 
        fmt.Sprintf(
            "Вы выбрали текстуру: %s\n%s\n\nКогда вам удобно выполнить заказ?",
            texture.Name,
            FormatSimplePriceBreakdown(width, height, priceDetails["final_price"]),
        ),
    )
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

func (b *Bot) handleTextureSelectionMessage(ctx context.Context, chatID int64, textureName string) {
    // Get texture from storage by name
    texture, err := b.storage.GetTextureByName(ctx, textureName)
    if err != nil {
        b.logger.Error("Failed to get texture by name",
            zap.String("texture_name", textureName),
            zap.Error(err))
        b.sendError(chatID, "Не удалось найти выбранную текстуру")
        return
    }

    // Get dimensions from state
    width, height, err := b.state.GetDimensions(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get dimensions",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при получении размеров")
        return
    }

    // Calculate price
    pricingConfig := PricingConfig{
        LeatherPricePerDM2:    texture.PricePerDM2,
        ProcessingCostPerDM2:  b.cfg.Pricing.ProcessingCostPerDM2,
        PaymentCommissionRate: b.cfg.Pricing.PaymentCommissionRate,
        SalesTaxRate:          b.cfg.Pricing.SalesTaxRate,
        MarkupMultiplier:      b.cfg.Pricing.MarkupMultiplier,
    }
    priceDetails := CalculatePrice(width, height, pricingConfig)

    // Save selection
    if err := b.state.SetTexture(ctx, chatID, texture.ID, priceDetails["final_price"]); err != nil {
        b.logger.Error("Failed to set texture",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при сохранении текстуры")
        return
    }

    // Send confirmation message with price breakdown
    msg := tgbotapi.NewMessage(
        chatID, 
        fmt.Sprintf(
            "Вы выбрали текстуру: %s\n%s\n\nКогда вам удобно выполнить заказ?",
            texture.Name,
            FormatSimplePriceBreakdown(width, height, priceDetails["final_price"]),
        ),
    )
    msg.ReplyMarkup = b.CreateDateSelectionKeyboard()
    b.sendMessage(msg)

    // Update user step
    if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
        b.logger.Error("Failed to set date selection state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }
}
