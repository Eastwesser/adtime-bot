package bot

import (
	"context"

	"go.uber.org/zap"
)

// RegisterPrintingHandlers registers all printing-related handlers
func (b *Bot) RegisterPrintingHandlers() {
	b.handlers[StepPrintingSelection] = b.HandlePrintingSelection
	// Add more printing-related steps if needed
}

// HandlePrintingMenu shows the printing products menu with pagination
func (b *Bot) HandlePrintingMenu(ctx context.Context, chatID int64, page int) {
	if err := b.state.SetPrintingPage(ctx, chatID, page); err != nil {
		b.logger.Error("Failed to set printing page",
			zap.Int64("chat_id", chatID),
			zap.Error(err))
	}

	//msg := "📋 Выберите продукт:"
	//b.SendMessage(tgbotapi.NewMessage(chatID, msg), b.CreatePrintingMenuKeyboard(page))
}

// HandlePrintingNavigation handles pagination for printing products
func (b *Bot) HandlePrintingNavigation(ctx context.Context, chatID int64, action string) {
	currentPage, _ := b.state.GetPrintingPage(ctx, chatID)

	switch action {
	case "Далее":
		currentPage++
	case "Назад":
		if currentPage > 1 {
			currentPage--
		}
	}

	b.HandlePrintingMenu(ctx, chatID, currentPage)
}

// HandlePrintingSelection handles selection of a specific printing product
func (b *Bot) HandlePrintingSelection(ctx context.Context, chatID int64, productName string) {
	const operation = "bot.HandlePrintingSelection"

	if err := b.state.SetSelectedProduct(ctx, chatID, productName); err != nil {
		b.logger.Error("Failed to set selected product",
			zap.String("operation", operation),
			zap.Int64("chat_id", chatID),
			zap.Error(err))
		b.SendError(chatID, "Ошибка выбора продукта")
		return
	}

	/*
	   var msg string
	   var options []string

	   switch productName {
	   case "Визитки":
	       msg = "🖨 Параметры визиток:\n- Размер: 90x50 мм\n- Материал: мелованная бумага\n- Цветность: 4+0 или 4+4"
	       options = []string{"100 шт", "200 шт", "500 шт"}
	   case "Листовки":
	       msg = "📄 Параметры листовок:\n- Формат: А4, А5, А6\n- Плотность: 130-300 г/м²\n- Цветность: полноцвет"
	       options = []string{"А4", "А5", "А6"}
	   // ... other products
	   default:
	       b.SendError(chatID, "Продукт временно недоступен")
	       return
	   }

	   if err := b.state.SetStep(ctx, chatID, StepPrintingSelection); err != nil {
	       b.logger.Error("Failed to set printing selection step",
	           zap.String("operation", operation),
	           zap.Int64("chat_id", chatID),
	           zap.Error(err))
	   }

	   b.SendMessage(tgbotapi.NewMessage(chatID, msg), b.CreateOptionsKeyboard(options))
	*/
}

// HandlePrintingOption handles selection of printing options (quantity, size etc)
func (b *Bot) HandlePrintingOption(ctx context.Context, chatID int64, option string) {
	// Implement option handling logic here
	// This would be called from ProcessMessage when in StepPrintingSelection
}
