package bot

import (
	"adtime-bot/internal/storage"
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) NotifyPrivacyAgreement(ctx context.Context, username string) {
    if b.cfg.Admin.ChannelID == 0 {
        b.logger.Warn("Channel notifications disabled - no channel ID configured")
        return
    }
    
    text := fmt.Sprintf("🔐 Пользователь @%s подтвердил согласие на обработку персональных данных.", username)

    msg := tgbotapi.NewMessage(b.cfg.Admin.ChannelID, text)
    
    if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send privacy agreement notification to channel",
            zap.Error(err))
    }
}

// отправляет краткое уведомление в канал
func (b *Bot) NotifyNewOrderToChannel(ctx context.Context, order storage.Order, username string) {
    
    if b.cfg.Admin.ChannelID == 0 {
        b.logger.Warn("Channel notifications disabled - no channel ID configured")
        return
    }

    // Verify bot has proper permissions
    _, err := b.bot.GetChat(tgbotapi.ChatInfoConfig{
        ChatConfig: tgbotapi.ChatConfig{
            ChatID: b.cfg.Admin.ChannelID,
        },
    })
    if err != nil {
        b.logger.Error("Failed to verify channel access",
            zap.Int64("channel_id", b.cfg.Admin.ChannelID),
            zap.Error(err))
        return
    }

    // Add debug log to verify channel ID
    b.logger.Info("Preparing channel notification",
        zap.Int64("channel_id", b.cfg.Admin.ChannelID),
        zap.String("username", username))

    text := fmt.Sprintf(
        "📦 Новый заказ #%d\n"+
        "Тип: %s\n"+
        "Размер: %dx%d см\n"+
        "Цена: %.2f руб\n"+
        "Контакт: %s\n"+
        "TG: @%s",
        order.ID, order.TextureName, 
        order.WidthCM, order.HeightCM,
        order.Price,
        FormatPhoneNumber(order.Contact),
        username,
    )

    msg := tgbotapi.NewMessage(b.cfg.Admin.ChannelID, text)
    msg.ParseMode = "HTML" // Используем HTML вместо Markdown для совместимости
    
    if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send channel notification",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
    }
}

// отправляет детали заказа и Excel файл конкретному админу
func (b *Bot) NotifyAdmin(ctx context.Context, order storage.Order) {
    // Use a map to track notified admins
    notifiedAdmins := make(map[int64]bool)

    // Send to main admin if not zero
    if b.cfg.Admin.ChatID != 0 && !notifiedAdmins[b.cfg.Admin.ChatID] {
        b.sendAdminNotification(ctx, b.cfg.Admin.ChatID, order)
        notifiedAdmins[b.cfg.Admin.ChatID] = true
    }
    
    // Send to additional admins
    for _, adminID := range b.cfg.Admin.IDs {
        if adminID != 0 && !notifiedAdmins[adminID] {
            b.sendAdminNotification(ctx, adminID, order)
            notifiedAdmins[adminID] = true
        }
    }
}

func (b *Bot) sendAdminNotification(ctx context.Context, chatID int64, order storage.Order) {
    if chatID == 0 {
        b.logger.Warn("Skipping notification to zero chat ID")
        return
    }

    // Create Excel file first (so we can fail fast if there's an error)
    filepath, err := b.storage.ExportOrderToExcel(ctx, order)
    if err != nil {
        b.logger.Error("Failed to create Excel file for order",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
        return
    }

    // Prepare the notification message
    msg := tgbotapi.NewMessage(chatID, FormatOrderNotification(order))
    msg.ParseMode = "Markdown"

    // Only add buttons if we have a valid order ID
    if order.ID > 0 {
        msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
            tgbotapi.NewInlineKeyboardRow(
                tgbotapi.NewInlineKeyboardButtonData(
                    "✅ В обработку", 
                    fmt.Sprintf("status:%d:processing", order.ID)),
                tgbotapi.NewInlineKeyboardButtonData(
                    "❌ Отменить", 
                    fmt.Sprintf("status:%d:cancelled", order.ID)),
            ),
        )
    }

    if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send admin notification",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
        return
    }

    // Send the Excel file
    doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filepath))
    doc.Caption = fmt.Sprintf("📊 Детали заказа #%d", order.ID)
    
    if _, err := b.bot.Send(doc); err != nil {
        b.logger.Error("Failed to send Excel file to admin",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
    }
}
