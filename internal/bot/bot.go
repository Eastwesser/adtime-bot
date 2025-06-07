package bot

import (
	"adtime-bot/internal/bot/calculators"
	"adtime-bot/internal/config"
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Bot struct {
	bot      *tgbotapi.BotAPI
	logger   *zap.Logger
	state    *StateStorage
	storage  *storage.PostgresStorage
	cfg      *config.Config
	mu       sync.Mutex
	handlers map[string]func(context.Context, int64, string)
    printingCalculator *calculators.PrintingCalculator
    vinylCalculator    *calculators.VinylCalculator
}

func New(
	token string,
	redisClient *redis.Client,
	pgStorage *storage.PostgresStorage,
	logger *zap.Logger,
	cfg *config.Config,
) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	botAPI.Debug = true // Enable debug for development

	logger.Info("Bot authorized",
		zap.String("username", botAPI.Self.UserName),
		zap.Int64("id", botAPI.Self.ID))

	b := &Bot{
		bot:     botAPI,
		logger:  logger,
		state:   NewStateStorage(redisClient),
		storage: pgStorage,
		cfg:     cfg,
        printingCalculator: calculators.NewPrintingCalculator(),
        vinylCalculator:    calculators.NewVinylCalculator(),
	}

	b.RegisterHandlers()
    b.RegisterPrintingHandlers()
    b.RegisterVinylHandlers()

    return b, nil
}

func (b *Bot) RegisterHandlers() {
    b.handlers = map[string]func(context.Context, int64, string){
        StepPrivacyAgreement: b.HandlePrivacyAgreement,
        StepServiceSelection: b.HandleServiceSelection,
		StepServiceType:      b.HandleServiceType,
		StepDimensions:       b.HandleDimensionsSize,
		StepDateSelection:    b.HandleDateSelection,
		StepManualDateInput:  b.HandleManualDateInput,
		StepDateConfirmation: b.HandleDateConfirmation,
		StepContactMethod:    b.HandleContactMethod,
		StepPhoneNumber:      b.HandlePhoneNumber,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting bot")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Shutting down bot")
			return nil

		case update := <-updates:
			b.mu.Lock()
			if update.Message != nil {
				b.ProcessMessage(ctx, update.Message)
			} else if update.CallbackQuery != nil {
				b.ProcessCallback(ctx, update.CallbackQuery)
			}
			b.mu.Unlock()
		}
	}
}

func (b *Bot) ProcessMessage(ctx context.Context, message *tgbotapi.Message) {
    
    chatID := message.Chat.ID

    
	// Handle contact sharing first
    if message.Contact != nil {
        // Normalize the phone number first
        normalized := NormalizePhoneNumber(message.Contact.PhoneNumber)
        if !IsValidPhoneNumber(normalized) {
            b.SendError(chatID, "Пожалуйста, предоставьте действительный номер телефона")
            return
        }
        
        // Skip phone number input step and proceed to create order
        _, err := b.CreateOrder(ctx, chatID, normalized)
        if err != nil {
            b.logger.Error("Failed to create order from contact",
                zap.Int64("chat_id", chatID),
                zap.Error(err))
            b.SendError(chatID, "Ошибка при оформлении заказа")
            return
        }

        // Just clear state without sending another message
        b.state.ClearState(ctx, chatID)
        return
	}
    
    if message.IsCommand() {
        // Split command and arguments
        cmd := message.Command()
        args := strings.Fields(message.CommandArguments())
        
        // First check if it's an admin command
        if b.IsAdmin(chatID) {
            b.HandleAdminCommand(ctx, chatID, cmd, args)
            return
        }
        
        // Handle regular user commands
        switch cmd {
        case "start":
            b.HandleStart(ctx, chatID)
        case "help":
            b.HandleHelp(ctx, chatID)
        case "new_order":
            b.HandleNewOrder(ctx, chatID)
        default:
            b.HandleUnknownCommand(ctx, chatID)
        }
        return
        
    }

    // Add handling for "New Order" button
    if message.Text == "🆕 Новый заказ" {
        b.HandleNewOrder(ctx, chatID)
        return
    }

    // Handle regular messages
    step, err := b.state.GetStep(ctx, chatID)
    if err != nil {
        b.logger.Error("Failed to get user step", zap.Error(err))
        return
    }

    // Добавляем обработку кастомной текстуры
    if step == CustomTextureInput {
        b.HandleCustomTextureInput(ctx, chatID, message.Text)
        return
    }

    // Special case for texture selection from reply keyboard
    if step == StepTextureSelection {
        b.HandleTextureSelectionMessage(ctx, chatID, message.Text)
        return
    }

    if handler, ok := b.handlers[step]; ok {
        handler(ctx, chatID, message.Text)
    } else {
        b.HandleDefault(ctx, chatID)
    }

    // Получаем текущее меню
    currentMenu, _ := b.state.GetCurrentMenu(ctx, chatID)

    // Обработка главного меню
    if message.Text == "Типография" {
        b.state.SetCurrentMenu(ctx, chatID, "printing")
        b.HandlePrintingMenu(ctx, chatID, 1)
        return
    }

    // Обработка навигации
    if message.Text == "Назад" || message.Text == "Далее" {
        if currentMenu == "printing" {
            b.HandlePrintingNavigation(ctx, chatID, message.Text)
            return
        }
    }

    // Обработка выбора продукта
    printingProducts := []string{"Визитки", "Бирки", "Листовки", "Буклеты", "Каталоги", "Календари", "Открытки"}
    if slices.Contains(printingProducts, message.Text) {
        b.HandlePrintingSelection(ctx, chatID, message.Text)
        return
    }

    // Обработка главного меню
    if message.Text == "Печать наклеек" {
        b.state.SetCurrentMenu(ctx, chatID, "vinyl")
        b.HandleVinylMenu(ctx, chatID)
        return
    }

    // Обработка выбора услуги наклеек
    vinylServices := []string{"Печать на пленке", "Резка пленки", "Ламинация", "Комплекс"}
    if slices.Contains(vinylServices, message.Text) && currentMenu == "vinyl" {
        b.HandleVinylSelection(ctx, chatID, message.Text)
        return
    }
}

func (b *Bot) ProcessCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID
    
    switch {
    case strings.HasPrefix(callback.Data, "texture:"):
        b.HandleTextureSelection(ctx, callback)
    case callback.Data == "cancel":
        b.HandleCancel(ctx, chatID)
    case strings.HasPrefix(callback.Data, "status:"):
        parts := strings.Split(callback.Data, ":")
        b.HandleStatusUpdate(ctx, callback.Message.Chat.ID, parts[1], parts[2])    
    default:
        b.logger.Warn("Unknown callback received",
            zap.String("callback_data", callback.Data),
            zap.Int64("chat_id", chatID),
            zap.Int("message_id", callback.Message.MessageID),
            zap.String("user", callback.From.UserName))
        b.SendError(chatID, "Неизвестная команда")
    }
}

func (b *Bot) HandleAdminStatusUpdate(ctx context.Context, chatID int64, orderIDStr, action string) {
    if !b.IsAdmin(chatID) {
        b.SendError(chatID, "❌ У вас нет прав для этого действия")
        return
    }

    orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
    if err != nil {
        b.SendError(chatID, "❌ Неверный ID заказа")
        return
    }

    var newStatus string
    switch action {
    case "processing":
        newStatus = "processing"
    case "cancelled":
        newStatus = "cancelled"
    default:
        b.SendError(chatID, "❌ Неизвестное действие")
        return
    }

    err = b.storage.UpdateOrderStatus(ctx, orderID, newStatus)
    if err != nil {
        b.logger.Error("Failed to update order status", zap.Error(err))
        b.SendError(chatID, "❌ Ошибка при обновлении статуса")
        return
    }

    // Отправляем подтверждение админу
    b.SendMessage(tgbotapi.NewMessage(chatID, fmt.Sprintf(
        "✅ Статус заказа #%d изменён на: %s",
        orderID,
        map[string]string{
            "processing": "В обработке",
            "cancelled": "Отменён",
        }[newStatus],
    )))

    // Уведомляем пользователя
    order, err := b.storage.GetOrderByID(ctx, orderID)
    if err == nil {
        userMsg := tgbotapi.NewMessage(order.UserID, fmt.Sprintf(
            "ℹ️ Статус вашего заказа #%d изменён на: %s",
            orderID,
            map[string]string{
                "processing": "В обработке",
                "cancelled": "Отменён",
            }[newStatus],
        ))
        b.SendMessage(userMsg)
    }
}

func (b *Bot) SendError(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+text)
	b.SendMessage(msg)
}

func (b *Bot) IsAdmin(chatID int64) bool {
    // Add debug logging
    b.logger.Debug("Admin check",
        zap.Int64("chatID", chatID),
        zap.Int64("configAdminID", b.cfg.Admin.ChatID),
        zap.Any("adminIDs", b.cfg.Admin.IDs))
    
    // Check both the main admin and additional admins
    return chatID == b.cfg.Admin.ChatID || slices.Contains(b.cfg.Admin.IDs, chatID)
}

func (b *Bot) ExportOrdersToSingleFile(ctx context.Context) error {
	filename := fmt.Sprintf("orders_report_%s", time.Now().Format("20060102_1504"))
	return b.storage.ExportAllOrdersToExcel(ctx, filename)
}

func (b *Bot) SendMessage(msg tgbotapi.MessageConfig) {

    // Send new message
    sentMsg, err := b.bot.Send(msg)
    if err != nil {
        b.logger.Error("Failed to send message",
            zap.Int64("chatID", msg.ChatID),
            zap.String("text", msg.Text),
            zap.Error(err))
        return
    }
    
    // Store the new message ID (no error check needed if function doesn't return error)
    b.state.SetLastBotMessageID(context.Background(), msg.ChatID, sentMsg.MessageID)
}

func (b *Bot) DeletePreviousBotMessage(chatID int64) {
    msgID, err := b.state.GetLastBotMessageID(context.Background(), chatID)
    if err != nil {
        b.logger.Warn("Failed to get last message ID",
            zap.Int64("chat_id", chatID),
            zap.Error(err))
        return
    }
    
    if msgID > 0 {
        delMsg := tgbotapi.NewDeleteMessage(chatID, msgID)
        if _, err := b.bot.Send(delMsg); err != nil {
            b.logger.Warn("Failed to delete previous message",
                zap.Int64("chat_id", chatID),
                zap.Int("message_id", msgID),
                zap.Error(err))
        }
    }
}
