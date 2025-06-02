package bot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)


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
