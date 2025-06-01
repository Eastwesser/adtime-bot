package bot

import (
	"adtime-bot/internal/storage"
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

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
