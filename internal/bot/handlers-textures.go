package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) HandleTextureSelection(ctx context.Context, callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID
    
    // Parse texture ID from callback data
    parts := strings.Split(callback.Data, ":")
    if len(parts) != 2 {
        b.SendError(chatID, "Неверный формат выбора текстуры")
        return
    }
    textureID := parts[1]

    // Get dimensions from state
    width, height, err := b.state.GetDimensions(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get dimensions",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.SendError(chatID, "Ошибка при получении размеров")
        return
    }

    // Get texture from storage
    texture, err := b.storage.GetTextureByID(ctx, textureID)
    if err != nil {
        b.logger.Error("Failed to get texture",
            zap.String("texture_id", textureID),
            zap.Error(err))
        b.SendError(chatID, "Не удалось получить информацию о текстуре")
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
        b.SendError(chatID, "Ошибка при сохранении текстуры")
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
    b.SendMessage(msg)

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

func (b *Bot) HandleTextureSelectionMessage(ctx context.Context, chatID int64, textureName string) {
    // Get texture from storage by name
    texture, err := b.storage.GetTextureByName(ctx, textureName)
    if err != nil {
        b.logger.Error("Failed to get texture by name",
            zap.String("texture_name", textureName),
            zap.Error(err))
        b.SendError(chatID, "Не удалось найти выбранную текстуру")
        return
    }

    // Get dimensions from state
    width, height, err := b.state.GetDimensions(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get dimensions",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.SendError(chatID, "Ошибка при получении размеров")
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
        b.SendError(chatID, "Ошибка при сохранении текстуры")
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
    b.SendMessage(msg)

    // Update user step
    if err := b.state.SetStep(ctx, chatID, StepDateSelection); err != nil {
        b.logger.Error("Failed to set date selection state",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }
}
