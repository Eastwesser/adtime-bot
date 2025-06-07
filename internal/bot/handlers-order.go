package bot

import (
	"adtime-bot/internal/storage"
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) CreateOrder(ctx context.Context, chatID int64, phone string) (int64, error) {

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
    texture, err := b.GetOrderTexture(ctx, chatID, state)
    if err != nil {
        b.logger.Error("Failed to get texture for order",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        return 0, fmt.Errorf("texture selection required: %w", err)
    }

    // Add debug logging to verify prices
    b.logger.Debug("Using texture for pricing",
        zap.String("texture_id", texture.ID),
        zap.String("texture_name", texture.Name),
        zap.Float64("price_per_dm2", texture.PricePerDM2))

	priceDetails, err := b.CalculateOrderPrice(width, height, texture)
    if err != nil {
        b.logger.Error("Failed to calculate price",
            zap.Int("width", width),
            zap.Int("height", height),
            zap.Any("texture", texture),
            zap.Error(err))
        return 0, fmt.Errorf("price calculation failed: %w", err)
    }

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

    // After order creation
    b.logger.Info("Testing admin notification",
        zap.Int64("admin_chat", b.cfg.Admin.ChatID),
        zap.Int64s("admin_ids", b.cfg.Admin.IDs))

	// Get username for notification
    username := ""
    chat, err := b.bot.GetChat(tgbotapi.ChatInfoConfig{
        ChatConfig: tgbotapi.ChatConfig{
            ChatID: chatID,
        },
    })
    if err == nil {
        username = chat.UserName
        if username == "" {
            username = fmt.Sprintf("id%d", chatID)
        }
    } else {
        username = fmt.Sprintf("id%d", chatID)
    }
    
	// Update the order with the actual ID before notifications
    order.ID = orderID

    // Send notifications with the updated order
    b.SendUserConfirmation(ctx, chatID, orderID, phone, width, height, priceDetails)
    
    go func() {
        b.NotifyAdmin(ctx, order)
        b.NotifyNewOrderToChannel(ctx, order, username)
    }()

    return orderID, nil
}

func (b *Bot) HandleMainMenu(ctx context.Context, chatID int64) {
    msg := tgbotapi.NewMessage(chatID, "Главное меню:")
    msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("/new_order"),
            tgbotapi.NewKeyboardButton("/order_history"),
        ),
    )
    b.SendMessage(msg)
}


func (b *Bot) HandleOrderHistory(ctx context.Context, chatID int64) {
    orders, err := b.storage.GetUserOrders(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get user orders", zap.Error(err))
        b.SendError(chatID, "Ошибка при получении истории заказов")
        return
    }

    if len(orders) == 0 {
        b.SendMessage(tgbotapi.NewMessage(chatID, "У вас пока нет заказов"))
        return
    }

    var sb strings.Builder
    sb.WriteString("📋 Ваши заказы:\n\n")
    for _, order := range orders {
        sb.WriteString(fmt.Sprintf(
            "🆔 #%d\n📅 %s\n📏 %dx%d см\n💵 %.2f ₽\n🔄 %s\n\n",
            order.ID,
            order.CreatedAt.Format("02.01.2006"),
            order.WidthCM,
            order.HeightCM,
            order.Price,
            order.Status,
        ))
    }

    msg := tgbotapi.NewMessage(chatID, sb.String())
    b.SendMessage(msg)
}

func (b *Bot) GetOrderTexture(ctx context.Context, chatID int64, state UserState) (*storage.Texture, error) {
    // First try by texture ID
    if state.TextureID != "" {
        texture, err := b.storage.GetTextureByID(ctx, state.TextureID)
        if err == nil {
            return texture, nil
        }
        b.logger.Warn("Failed to get texture by ID, falling back to service name",
            zap.String("texture_id", state.TextureID),
            zap.Error(err))
    }

    // Fall back to service name if texture ID not set
    if state.Service != "" {
        texture, err := b.storage.GetTextureByName(ctx, state.Service)
        if err == nil {
            return texture, nil
        }
        b.logger.Warn("Failed to get texture by service name",
            zap.String("service", state.Service),
            zap.Error(err))
    }

    return nil, fmt.Errorf("no texture selected")
}

func (b *Bot) CalculateOrderPrice(width, height int, texture *storage.Texture) (map[string]float64, error) {
    pricingConfig := PricingConfig{
        LeatherPricePerDM2:    texture.PricePerDM2,
        ProcessingCostPerDM2:  b.cfg.Pricing.ProcessingCostPerDM2,
        PaymentCommissionRate: b.cfg.Pricing.PaymentCommissionRate,
        SalesTaxRate:          b.cfg.Pricing.SalesTaxRate,
        MarkupMultiplier:      b.cfg.Pricing.MarkupMultiplier,
    }

    return CalculatePrice(width, height, pricingConfig)
}

func (b *Bot) SendUserConfirmation(ctx context.Context, chatID, orderID int64, phone string, width, height int, priceDetails map[string]float64) {
    // Сохраняем согласие и телефон пользователя
    err := b.storage.SaveUserAgreement(ctx, chatID, phone)
    if err != nil {
        b.logger.Error("Failed to save user agreement", zap.Error(err))
    }
    
    // This will be the ONLY confirmation message for user
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
    msg.ReplyMarkup = b.CreateMainMenuKeyboard() // Добавляем главное меню
    b.SendMessage(msg)
    // show keyboard for another order
    b.ShowMainMenu(ctx, chatID, phone)
}

func (b *Bot) HandleNewOrder(ctx context.Context, chatID int64) {
    // Check saved data
    _, phone, err := b.storage.GetUserAgreement(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get user agreement", zap.Error(err))
        b.SendError(chatID, "Произошла ошибка, попробуйте позже")
        return
    }

    // Clear order state but keep phone number
    if err := b.state.ResetOrderState(ctx, chatID); err != nil {
        b.logger.Error("Failed to clear order state", 
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }

    if phone != "" {
        // User has saved phone - skip to service selection
        msg := tgbotapi.NewMessage(chatID, "Начнём новый заказ! Выберите тип услуги:")
        msg.ReplyMarkup = b.CreateServiceTypeKeyboard()
        b.SendMessage(msg)
        b.state.SetStep(ctx, chatID, StepServiceType)
    } else {
        // No saved phone - start standard process
        b.HandleStart(ctx, chatID)
    }
    // if phone != "" {
    //     // User has saved phone - skip agreement and phone steps
    //     b.state.SetPhoneNumber(ctx, chatID, phone)
        
    //     msg := tgbotapi.NewMessage(chatID, "Начнём новый заказ! Выберите тип услуги:")
    //     msg.ReplyMarkup = b.CreateServiceTypeKeyboard()
    //     b.SendMessage(msg)
    //     b.state.SetStep(ctx, chatID, StepServiceType)
    // } else {
    //     // No saved phone - start standard process
    //     b.HandleStart(ctx, chatID)
    // }
}
