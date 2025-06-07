package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// RegisterVinylHandlers registers all vinyl-related handlers
func (b *Bot) RegisterVinylHandlers() {
    b.handlers[StepVinylSelection] = b.HandleVinylSelection
    // Add more vinyl-related steps if needed
}

// HandleVinylMenu shows the vinyl services menu
func (b *Bot) HandleVinylMenu(ctx context.Context, chatID int64) {
    msg := `🎨 Выберите услугу для пленки ORACAL:
1. Печать на пленке
2. Резка пленки
3. Ламинация + резка
4. Комплексная обработка`
    
    b.SendMessage(tgbotapi.NewMessage(chatID, msg), b.CreateVinylServicesKeyboard())
}

// HandleVinylSelection handles selection of a specific vinyl service
func (b *Bot) HandleVinylSelection(ctx context.Context, chatID int64, service string) {
    const operation = "bot.HandleVinylSelection"
    
    if err := b.state.SetService(ctx, chatID, service); err != nil {
        b.logger.Error("Failed to set vinyl service",
            zap.String("operation", operation),
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        b.SendError(chatID, "Ошибка выбора услуги")
        return
    }

    msg := fmt.Sprintf("🔧 Вы выбрали: %s\nУкажите параметры:\n1. Размер в м²\n2. Цветность", service)
    options := []string{"0.5 м²", "1 м²", "2 м²", "5 м²"}

    if err := b.state.SetStep(ctx, chatID, StepVinylSelection); err != nil {
        b.logger.Error("Failed to set vinyl selection step",
            zap.String("operation", operation),
            zap.Int64("chat_id", chatID),
            zap.Error(err))
    }

    b.SendMessage(tgbotapi.NewMessage(chatID, msg), b.CreateOptionsKeyboard(options))
}

// HandleVinylOption handles selection of vinyl options (size, color etc)
func (b *Bot) HandleVinylOption(ctx context.Context, chatID int64, option string) {
    // Implement option handling logic here
    // This would be called from ProcessMessage when in StepVinylSelection
}