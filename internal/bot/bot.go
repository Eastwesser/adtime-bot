package bot

import (
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

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
	StepTextureSelection = "texture_selection"
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

func (b *Bot) processMessage(ctx context.Context, message *tgbotapi.Message) {
    chatID := message.Chat.ID

	// Handle contact sharing first
    if message.Contact != nil {
    // Normalize the phone number first
    normalized := NormalizePhoneNumber(message.Contact.PhoneNumber)
    if !IsValidPhoneNumber(normalized) {
        b.sendError(chatID, "Пожалуйста, предоставьте действительный номер телефона")
        return
    }
    
    // Skip phone number input step and proceed to create order
    orderID, err := b.createOrder(ctx, chatID, normalized)
    if err != nil {
        b.logger.Error("Failed to create order from contact",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.sendError(chatID, "Ошибка при оформлении заказа")
        return
    }
    
    // Clear state and send confirmation
    b.state.ClearState(ctx, chatID)
    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("✅ Ваш заказ успешно оформлен!\nНомер заказа: #%d\n\nС вами свяжутся в ближайшее время.", orderID))
		b.sendMessage(msg)
		return
	}
    
    if message.IsCommand() {
        // Split command and arguments
        cmd := message.Command()
        args := strings.Fields(message.CommandArguments())
        
        // First check if it's an admin command
        if b.isAdmin(chatID) {
            b.handleAdminCommand(ctx, chatID, cmd, args)
            return
        }
        
        // Handle regular user commands
        switch cmd {
        case "start":
            b.handleStart(ctx, chatID)
        case "help":
            b.handleHelp(ctx, chatID)
        default:
            b.handleUnknownCommand(ctx, chatID)
        }
        return
    }

    // Handle regular messages
    step, err := b.state.GetStep(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get user step", zap.Error(err))
        return
    }

    // Special case for texture selection from reply keyboard
    if step == StepTextureSelection {
        b.handleTextureSelectionMessage(ctx, chatID, message.Text)
        return
    }

    if handler, ok := b.handlers[step]; ok {
        handler(ctx, chatID, message.Text)
    } else {
        b.handleDefault(ctx, chatID)
    }
}

func (b *Bot) processCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID
    
    switch {
    case strings.HasPrefix(callback.Data, "texture:"):
        b.handleTextureSelection(ctx, callback)
    case callback.Data == "cancel":
        b.handleCancel(ctx, chatID)
    default:
        b.logger.Warn("Unknown callback received",
            zap.String("callback_data", callback.Data),
            zap.Int64("chat_id", chatID),
            zap.Int("message_id", callback.Message.MessageID),
            zap.String("user", callback.From.UserName))
        b.sendError(chatID, "Неизвестная команда")
    }
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

func (b *Bot) ExportOrdersToSingleFile(ctx context.Context) error {
	filename := fmt.Sprintf("orders_report_%s", time.Now().Format("20060102_1504"))
	return b.storage.ExportAllOrdersToExcel(ctx, filename)
}

func (b *Bot) notifyAdmin(ctx context.Context, order storage.Order) {
	if b.cfg.Admin.ChatID == 0 {
        b.logger.Warn("Admin notifications disabled - no chat ID configured")
        return
    }

	msg := tgbotapi.NewMessage(b.cfg.Admin.ChatID, FormatOrderNotification(order))
	// Add action buttons
    markup := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📊 Экспорт", fmt.Sprintf("export:%d", order.ID)),
            tgbotapi.NewInlineKeyboardButtonData("✏️ Статус", fmt.Sprintf("status:%d", order.ID)),
        ),
    )
    msg.ReplyMarkup = markup

    // text := fmt.Sprintf(
    //     "📦 *Новый заказ* #%d\n"+
    //         "👤 Пользователь: %d\n"+
    //         "📏 Размер: %d×%d см (%.1f дм²)\n"+
    //         "🧵 Текстура: %s\n"+
    //         "💰 Цена: %.2f ₽\n"+
    //         "📞 Контакт: %s\n"+
    //         "⏱ Создан: %s",
    //     order.ID,
    //     order.UserID,
    //     order.WidthCM, order.HeightCM,
    //     float64(order.WidthCM*order.HeightCM)/100,
    //     order.TextureName,
    //     order.Price,
    //     FormatPhoneNumber(order.Contact),
    //     order.CreatedAt.Format("02.01.2006 15:04"),
    // )

	if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send admin notification",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
    }

    // msg := tgbotapi.NewMessage(b.cfg.Admin.ChatID, text)
    msg.ParseMode = "Markdown"

    // markup := tgbotapi.NewInlineKeyboardMarkup(
    //     tgbotapi.NewInlineKeyboardRow(
    //         tgbotapi.NewInlineKeyboardButtonData("📊 Экспорт в Excel", fmt.Sprintf("export:%d", order.ID)),
    //         tgbotapi.NewInlineKeyboardButtonData("✏️ Изменить статус", fmt.Sprintf("status:%d", order.ID)),
    //     ),
    // )
    msg.ReplyMarkup = &markup

    // if _, err := b.bot.Send(msg); err != nil {
    //     b.logger.Error("Failed to send admin notification",
    //         zap.Int64("order_id", order.ID),
    //         zap.Error(err))
    // }
	// text := fmt.Sprintf(
	// 	"📦 *Новый заказ* #%d\n"+
	// 		"👤 Пользователь: %d\n"+
	// 		"📏 Размер: %d×%d см (%.1f дм²)\n"+
	// 		"🧵 Текстура: %s\n"+
	// 		"💰 Цена: %.2f ₽\n"+
	// 		"📞 Контакт: %s\n"+
	// 		"⏱ Создан: %s",
	// 	order.ID,
	// 	order.UserID,
	// 	order.WidthCM, order.HeightCM,
	// 	float64(order.WidthCM*order.HeightCM)/100,
	// 	order.TextureName,
	// 	order.Price,
	// 	FormatPhoneNumber(order.Contact),
	// 	order.CreatedAt.Format("02.01.2006 15:04"),
	// )

	// msg := tgbotapi.NewMessage(b.cfg.Admin.ChatID, text)
	// msg.ParseMode = "Markdown"

	// markup := tgbotapi.NewInlineKeyboardMarkup(
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("📊 Экспорт в Excel", fmt.Sprintf("export:%d", order.ID)),
	// 		tgbotapi.NewInlineKeyboardButtonData("✏️ Изменить статус", fmt.Sprintf("status:%d", order.ID)),
	// 	),
	// )
	// msg.ReplyMarkup = &markup

	// if _, err := b.bot.Send(msg); err != nil {
	// 	b.logger.Error("Failed to send admin notification",
	// 		zap.Int64("order_id", order.ID),
	// 		zap.Error(err))
	// }
}

func (b *Bot) isAdmin(chatID int64) bool {
	for _, id := range b.cfg.Admin.IDs {
		if chatID == id {
			return true
		}
	}
	return chatID == b.cfg.Admin.ChatID
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

func (b *Bot) createOrder(ctx context.Context, chatID int64, phone string) (int64, error) {
    
	state, err := b.state.GetFullState(ctx, chatID)
    if err != nil {
        return 0, fmt.Errorf("failed to get order state: %w", err)
    }

	width, height, err := b.state.GetDimensions(ctx, chatID)
    if err != nil {
        return 0, fmt.Errorf("failed to get dimensions: %w", err)
    }

	// Validate dimensions
    if width <= 0 || height <= 0 {
        return 0, fmt.Errorf("invalid dimensions: width=%d height=%d", width, height)
    }

    // Use the helper function to get texture
    texture, err := b.getOrderTexture(ctx, chatID, state)
    if err != nil {
        b.logger.Error("Failed to get texture",
            zap.Int64("chat_id", chatID),
            zap.Any("state", state),
            zap.Error(err))
        return 0, fmt.Errorf("failed to get texture: %w", err)
    }

	priceDetails := b.calculateOrderPrice(width, height, texture)

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

	orderID, err := b.storage.SaveOrder(ctx, order)
    if err != nil {
        b.logger.Error("Failed to save order",
            zap.Int64("chat_id", chatID),
            zap.Any("order", order),
            zap.Error(err))
        return 0, fmt.Errorf("failed to save order: %w", err)
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
    }

	// Send notifications
    b.sendUserConfirmation(ctx, chatID, orderID, phone, width, height, priceDetails)
    go b.notifyAdmin(ctx, order)
    go b.notifyNewOrderToChannel(ctx, order, username)

    return orderID, nil
}

// Helper functions for createOrder
func (b *Bot) getOrderTexture(ctx context.Context, chatID int64, state UserState) (*storage.Texture, error) {
    // If texture ID is set in state, try to get it from storage
    if state.TextureID != "" {
        texture, err := b.storage.GetTextureByID(ctx, state.TextureID)
        if err != nil {
            b.logger.Warn("Failed to get texture by ID, using default",
                zap.String("texture_id", state.TextureID),
                zap.Error(err))
        } else {
            return texture, nil
        }
    }

    // Fall back to default texture
    b.logger.Info("Using default texture",
        zap.Int64("chat_id", chatID))
        
    return &storage.Texture{
        ID:          "11111111-1111-1111-1111-111111111111",
        Name:        "Стандартная текстура",
        PricePerDM2: 10.0,
    }, nil
}

func (b *Bot) calculateOrderPrice(width, height int, texture *storage.Texture) map[string]float64 {
	pricingConfig := PricingConfig{
		LeatherPricePerDM2:    texture.PricePerDM2,
		ProcessingCostPerDM2:  b.cfg.Pricing.ProcessingCostPerDM2,
		PaymentCommissionRate: b.cfg.Pricing.PaymentCommissionRate,
		SalesTaxRate:          b.cfg.Pricing.SalesTaxRate,
		MarkupMultiplier:      b.cfg.Pricing.MarkupMultiplier,
	}
	return CalculatePrice(width, height, pricingConfig)
}

func (b *Bot) sendUserConfirmation(ctx context.Context, chatID, orderID int64, phone string, width, height int, priceDetails map[string]float64) {
	// Simple confirmation for user
    msgText := fmt.Sprintf(
        "✅ Ваш заказ #%d оформлен!\n"+
            "Размер: %d×%d см\n"+
            "Итоговая цена: %.2f ₽\n\n"+
            "С вами свяжутся в ближайшее время.",
        orderID,
        width, height,
        priceDetails["final_price"],
    )
	msg := tgbotapi.NewMessage(chatID, msgText)
    b.sendMessage(msg)
}

// func (b *Bot) deleteMessage(chatID int64, messageID int) {
//     delMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
//     if _, err := b.bot.Send(delMsg); err != nil {
//         b.logger.Warn("Failed to delete message",
//             zap.Int("message_id", messageID),
//             zap.Error(err))
//     }
// }

func (b *Bot) notifyPrivacyAgreement(ctx context.Context, username string) {
    if b.cfg.Admin.ChannelID == 0 {
        return
    }

    text := fmt.Sprintf("🔐 Пользователь @%s подтвердил согласие на обработку персональных данных.", username)
    msg := tgbotapi.NewMessage(b.cfg.Admin.ChannelID, text)
    msg.ParseMode = "HTML"
    
    if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send privacy agreement notification to channel",
            zap.Error(err))
    }
}

func (b *Bot) notifyNewOrderToChannel(ctx context.Context, order storage.Order, username string) {
    if b.cfg.Admin.ChannelID == 0 {
        return
    }

    text := fmt.Sprintf(
        "📦 Новый заказ\n"+
            "Услуга: %s\n"+
            "Дата: %s\n"+
            "Контакт: %s\n"+
            "Telegram: @%s",
        order.TextureName,
        order.CreatedAt.Format("02.01.2006"),
        FormatPhoneNumber(order.Contact),
        username,
    )

    msg := tgbotapi.NewMessage(b.cfg.Admin.ChannelID, text)
    msg.ParseMode = "HTML"
    
    if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send order notification to channel",
            zap.Error(err))
    }
}
