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
    
    
    // doc := nil
    // doc.Caption = fmt.Sprintf("📊 Детали заказа #%d", order.ID)
    // if _, err := b.bot.Send(doc); err != nil {}
    
    // Отправляем основное уведомление
    if b.cfg.Admin.ChatID != 0 {
        b.sendAdminNotification(ctx, b.cfg.Admin.ChatID, order)
    }
    
    // Отправляем уведомления дополнительным админам
    for _, adminID := range b.cfg.Admin.IDs {
        if adminID != 0 {
            b.sendAdminNotification(ctx, adminID, order)
        }
    }
}

func (b *Bot) sendAdminNotification(ctx context.Context, chatID int64, order storage.Order) {
    
    if chatID == 0 {
        b.logger.Warn("Skipping notification to zero chat ID")
        return
    }
    
    // Создаем Excel файл
    filepath, err := b.storage.ExportOrderToExcel(ctx, order)
    if err != nil {
        b.logger.Error("Failed to create Excel file for order",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
        return
    }

    // Отправляем сообщение с деталями заказа
    msg := tgbotapi.NewMessage(chatID, FormatOrderNotification(order))
    if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send order notification",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        return
    }
    msg.ParseMode = "Markdown"
    
    // Добавляем кнопки действий
    markup := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("✅ В обработку", fmt.Sprintf("status:%d:processing", order.ID)),
            tgbotapi.NewInlineKeyboardButtonData("❌ Отменить", fmt.Sprintf("status:%d:cancelled", order.ID)),
        ),
    )
    msg.ReplyMarkup = markup

    if _, err := b.bot.Send(msg); err != nil {
        b.logger.Error("Failed to send admin notification",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
    }

    // Отправляем Excel файл
    doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filepath))
    if _, err := b.bot.Send(doc); err != nil {
        b.logger.Error("Failed to send order document",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }
    doc.Caption = fmt.Sprintf("📊 Детали заказа #%d", order.ID)
    
    if _, err := b.bot.Send(doc); err != nil {
        b.logger.Error("Failed to send Excel file to admin",
            zap.Int64("order_id", order.ID),
            zap.Error(err))
    }
}
