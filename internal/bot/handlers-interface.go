package bot

import (
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type BotHandler interface {
	// Command handlers
	HandleStart(ctx context.Context, chatID int64)
	HandleCancel(ctx context.Context, chatID int64)
	HandleHelp(ctx context.Context, chatID int64)
	HandleUnknownCommand(ctx context.Context, chatID int64)
	HandleDefault(ctx context.Context, chatID int64)
	HandleAdminCommand(ctx context.Context, chatID int64, cmd string, args []string)
	
	// Step handlers
	HandlePrivacyAgreement(ctx context.Context, chatID int64, text string)
	HandleServiceSelection(ctx context.Context, chatID int64, text string)
	HandleServiceType(ctx context.Context, chatID int64, text string)
	HandleDimensionsSize(ctx context.Context, chatID int64, text string)
	HandleDateSelection(ctx context.Context, chatID int64, text string)
	HandleManualDateInput(ctx context.Context, chatID int64, text string)
	HandleDateConfirmation(ctx context.Context, chatID int64, text string)
	HandleContactMethod(ctx context.Context, chatID int64, text string)
	HandlePhoneNumber(ctx context.Context, chatID int64, text string)
	
	// Texture handlers
	HandleTextureSelection(ctx context.Context, callback *tgbotapi.CallbackQuery)
	HandleTextureSelectionMessage(ctx context.Context, chatID int64, textureName string)
	
	// Admin handlers
	HandleStatusUpdate(ctx context.Context, chatID int64, orderIDStr string, newStatus string)
	HandleOrderStats(ctx context.Context, chatID int64)
	HandleExportAllOrders(ctx context.Context, chatID int64)
	HandleExportSingleOrder(ctx context.Context, chatID int64, orderID int64)
	
	// Utility methods
	CreateOrder(ctx context.Context, chatID int64, phone string) (int64, error)
	GetOrderTexture(ctx context.Context, chatID int64, state UserState) (*storage.Texture, error)
	CalculateOrderPrice(width, height int, texture *storage.Texture) map[string]float64
	SendUserConfirmation(ctx context.Context, chatID, orderID int64, phone string, width, height int, priceDetails map[string]float64)
	IsAdmin(chatID int64) bool
}

type HandlerDependencies struct {
	BotAPI  *tgbotapi.BotAPI
	Logger  *zap.Logger
	State   *StateStorage
	Storage *storage.PostgresStorage
	Cfg     *config.Config
}
