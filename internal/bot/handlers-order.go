package bot

import (
	"adtime-bot/internal/storage"
	"context"
	"fmt"
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
        b.logger.Error("Failed to get texture",
            zap.Int64("chat_id", chatID),
            zap.Any("state", state),
            zap.Error(err))
        return 0, fmt.Errorf("failed to get texture: %w", err)
    }

	priceDetails := b.CalculateOrderPrice(width, height, texture)

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
    b.SendUserConfirmation(ctx, chatID, orderID, phone, width, height, priceDetails)
    go b.NotifyAdmin(ctx, order)
    go b.NotifyNewOrderToChannel(ctx, order, username)

    return orderID, nil
}

func (b *Bot) GetOrderTexture(ctx context.Context, chatID int64, state UserState) (*storage.Texture, error) {
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

func (b *Bot) CalculateOrderPrice(
	width, 
	height int, 
	texture *storage.Texture,
	) map[string]float64 {

	pricingConfig := PricingConfig{
		LeatherPricePerDM2:    texture.PricePerDM2,
		ProcessingCostPerDM2:  b.cfg.Pricing.ProcessingCostPerDM2,
		PaymentCommissionRate: b.cfg.Pricing.PaymentCommissionRate,
		SalesTaxRate:          b.cfg.Pricing.SalesTaxRate,
		MarkupMultiplier:      b.cfg.Pricing.MarkupMultiplier,
	}

	return CalculatePrice(width, height, pricingConfig)
}

func (b *Bot) SendUserConfirmation(
	ctx context.Context, 
	chatID, 
	orderID int64, 
	phone string, 
	width, 
	height int, 
	priceDetails map[string]float64,
	) {
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
    b.SendMessage(msg)
}
