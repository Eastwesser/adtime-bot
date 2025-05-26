package bot

import (
	"adtime-bot/internal/storage"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// MESSAGE HANDLERS

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	switch {
	case msg.IsCommand():
		b.handleCommand(ctx, chatID, msg.Command())
	case b.state.IsWaitingForPrivacyAgreement(chatID):
		b.handlePrivacyAgreement(ctx, chatID, msg.Text)
	case b.state.IsWaitingForServiceSelection(chatID):
		b.handleServiceSelection(ctx, chatID, msg.Text)
	case b.state.IsWaitingForDateSelection(chatID):
		b.handleDateSelection(ctx, chatID, msg.Text)
	case b.state.IsWaitingForPhoneNumber(chatID):
		b.handlePhoneNumber(ctx, chatID, msg.Text)
	default:
		b.handleDefault(ctx, chatID)
	}
}

func (b *Bot) handleDefault(ctx context.Context, chatID int64) {
	panic("unimplemented")
}

func (b *Bot) handleCommand(ctx context.Context, chatID int64, command string) {
	switch command {
	case "start":
		b.handleStart(ctx, chatID)
	default:
		b.handleUnknownCommand(ctx, chatID)
	}
}

func (b *Bot) handleUnknownCommand(ctx context.Context, chatID int64) {
	panic("unimplemented")
}

func (b *Bot) handleStart(ctx context.Context, chatID int64) {
	text := "Привет! 👋\n\n⚠️ Прежде чем продолжить, ознакомьтесь с нашей Политикой конфиденциальности.\n" +
		"Используя этого бота, вы соглашаетесь на обработку персональных данных.\n\n" +
		"Если всё ок — нажмите кнопку ниже 👇"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Продолжить"),
		),
	)

	b.sendMessage(msg)
	b.state.SetWaitingForPrivacyAgreement(chatID)
}

func (b *Bot) handlePrivacyAgreement(ctx context.Context, chatID int64, text string) {
	if text != "✅ Продолжить" {
		b.sendError(chatID, "Пожалуйста, нажмите кнопку \"✅ Продолжить\" чтобы согласиться с условиями")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Спасибо за согласие! Теперь вы можете оформить заказ.")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Оформить заказ"),
		),
	)

	b.sendMessage(msg)
	b.state.SetWaitingForServiceSelection(chatID)
}

func (b *Bot) handleServiceSelection(ctx context.Context, chatID int64, text string) {
	if text != "✅ Оформить заказ" {
		b.sendError(chatID, "Пожалуйста, нажмите кнопку \"✅ Оформить заказ\" чтобы продолжить")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Какую услугу вы хотите заказать?")
	b.sendMessage(msg)
	b.state.SetWaitingForServiceInput(chatID)
}

func (b *Bot) handleServiceInputSizeCheck(ctx context.Context, chatID int64, text string) {
	// Установим выбранную услугу
	b.state.SetService(chatID, text)

	// Запрашиваем размеры
	msg := tgbotapi.NewMessage(chatID, "Введите ширину и длину в сантиметрах через пробел (например: 30 40)\nМаксимальный размер: 80x50 см")
	b.sendMessage(msg)
	b.state.SetWaitingDimensions(chatID)
}

func (b *Bot) handleDimensionsSize(ctx context.Context, chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendError(chatID, "Неверный формат. Введите ширину и длину через пробел")
		return
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil || width <= 0 || width > 80 {
		b.sendError(chatID, "Некорректная ширина. Допустимый диапазон: 1-80 см")
		return
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil || height <= 0 || height > 50 {
		b.sendError(chatID, "Некорректная длина. Допустимый диапазон: 1-50 см")
		return
	}

	// Исправленный вызов с передачей context
	if err := b.state.SetDimensions(ctx, chatID, width, height); err != nil {
		b.logger.Error("Failed to set dimensions", zap.Error(err))
		b.sendError(chatID, "Ошибка при сохранении размеров")
		return
	}

	b.showTextures(ctx, chatID)
}

func (b *Bot) showTextures(ctx context.Context, chatID int64) {
	textures, err := b.api.GetTextures(ctx)
	if err != nil {
		b.logger.Error("Failed to get textures", zap.Error(err))
		b.sendError(chatID, "Не удалось загрузить варианты текстуры")
		return
	}

	width, height := b.state.GetDimensions(chatID)
	var buttons []tgbotapi.InlineKeyboardButton

	for _, texture := range textures {
		if !texture.InStock {
			continue
		}

		price := calculatePrice(width, height, texture.PricePerDM2)
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s - %.2f руб", texture.Name, price),
			fmt.Sprintf("texture:%s", texture.ID),
		)
		buttons = append(buttons, btn)
	}

	if len(buttons) == 0 {
		b.sendError(chatID, "Нет доступных текстур")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите текстуру:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)
	b.sendMessage(msg)
}

func calculatePrice(widthCm, heightCm int, pricePerDM2 float64) float64 {
	widthDM := float64(widthCm) / 10
	heightDM := float64(heightCm) / 10
	return widthDM * heightDM * pricePerDM2
}

func (b *Bot) handleServiceInput(ctx context.Context, chatID int64, text string) {
	// Здесь можно добавить проверку доступных услуг
	b.state.SetService(chatID, text)

	msg := tgbotapi.NewMessage(chatID, "Когда вам удобно выполнить заказ?")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Сегодня"),
			tgbotapi.NewKeyboardButton("Завтра"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выбрать дату вручную"),
		),
	)
	b.sendMessage(msg)
	b.state.SetWaitingForDateSelection(chatID)
}

func (b *Bot) handleDateSelection(ctx context.Context, chatID int64, text string) {
	switch text {
	case "Сегодня":
		today := time.Now().Format("02.01.2006")
		b.state.SetDate(chatID, today)
		b.confirmDateSelection(ctx, chatID, today)
	case "Завтра":
		tomorrow := time.Now().Add(24 * time.Hour).Format("02.01.2006")
		b.state.SetDate(chatID, tomorrow)
		b.confirmDateSelection(ctx, chatID, tomorrow)
	case "Выбрать дату вручную":
		msg := tgbotapi.NewMessage(chatID, "Введите дату вручную в формате ДД.ММ.ГГГГ")
		b.sendMessage(msg)
		b.state.SetWaitingForManualDateInput(chatID)
	case "🔁 Сменить дату":
		msg := tgbotapi.NewMessage(chatID, "Хорошо, давайте выберем новую дату")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Сегодня"),
				tgbotapi.NewKeyboardButton("Завтра"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Выбрать дату вручную"),
			),
		)
		b.sendMessage(msg)
		b.state.SetWaitingForDateSelection(chatID)
	case "✅ Подтвердить дату":
		b.handleDateConfirmation(ctx, chatID)
	default:
		if b.state.IsWaitingForManualDateInput(chatID) {
			b.handleManualDateInput(ctx, chatID, text)
		} else {
			b.sendError(chatID, "Пожалуйста, выберите один из предложенных вариантов")
		}
	}
}

func (b *Bot) handleManualDateInput(ctx context.Context, chatID int64, text string) {
	_, err := time.Parse("02.01.2006", text)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "🕒 Путешествие во времени пока не поддерживается. Пожалуйста, выбери дату в будущем ⏳")
		b.sendMessage(msg)
		return
	}

	// Проверка что дата в будущем
	inputDate, _ := time.Parse("02.01.2006", text)
	if inputDate.Before(time.Now().Truncate(24 * time.Hour)) {
		msg := tgbotapi.NewMessage(chatID, "🕒 Путешествие во времени пока не поддерживается. Пожалуйста, выбери дату в будущем ⏳")
		b.sendMessage(msg)
		return
	}

	b.state.SetDate(chatID, text)
	b.confirmDateSelection(ctx, chatID, text)
}

func (b *Bot) confirmDateSelection(ctx context.Context, chatID int64, date string) {
	days := b.calculateWorkingDays(date)

	text := fmt.Sprintf("Отлично! Вы выбрали дату: %s\nДо этой даты %d рабочих дней (без учёта выходных и праздников).", date, days)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔁 Сменить дату"),
			tgbotapi.NewKeyboardButton("✅ Подтвердить дату"),
		),
	)
	b.sendMessage(msg)
	b.state.SetWaitingForDateConfirmation(chatID)
}

func (b *Bot) calculateWorkingDays(date string) int {
	// Простая реализация - можно улучшить с учетом праздников
	targetDate, _ := time.Parse("02.01.2006", date)
	now := time.Now()

	days := 0
	for d := now; d.Before(targetDate); d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			days++
		}
	}
	return days
}

func (b *Bot) handleDateConfirmation(ctx context.Context, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Введите ваш номер телефона")
	b.sendMessage(msg)
	b.state.SetWaitingForPhoneNumber(chatID)
}

func (b *Bot) handlePhoneNumber(ctx context.Context, chatID int64, text string) {
	if !isValidPhoneNumber(text) {
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введите реальный номер телефона с кодом страны (например, +79161234567)")
		b.sendMessage(msg)
		return
	}

	// Получаем данные заказа
	state, err := b.state.Get(ctx, chatID)
	if err != nil {
		b.logger.Error("Failed to get order state", zap.Error(err))
		b.sendError(chatID, "Ошибка при обработке заказа")
		return
	}

	price, _ := strconv.ParseFloat(state.Price, 64)

	// Сохраняем заказ в PostgreSQL
	order := storage.Order{
		UserID:    chatID,
		WidthCM:   state.WidthCM,
		HeightCM:  state.HeightCM,
		TextureID: state.TextureID,
		Price:     price,
		Contact:   text,
	}

	if err := b.storage.SaveOrder(ctx, order); err != nil {
		b.logger.Error("Failed to save order to database", zap.Error(err))
		b.sendError(chatID, "Ошибка при сохранении заказа")
		return
	}

	// Формируем Excel-совместимые данные
	excelData := fmt.Sprintf(
		"SKU\tШирина (см)\tДлина (см)\tПлощадь (см²)\tЦЕНА ДЛЯ КЛ (₽)\n"+
			"%s\t%d\t%d\t%d\t%.2f",
		state.TextureID,
		state.WidthCM,
		state.HeightCM,
		state.WidthCM*state.HeightCM,
		price,
	)

	// Отправляем данные в чат администратора
	adminMsg := tgbotapi.NewMessage(chatID, "Новый заказ:\n```\n"+excelData+"\n```")
	adminMsg.ParseMode = "Markdown"
	b.bot.Send(adminMsg)

	// Отправляем подтверждение пользователю
	msg := tgbotapi.NewMessage(chatID, "Спасибо! Ваш заказ отправлен. Мы скоро свяжемся с вами.")
	b.sendMessage(msg)
	b.state.ClearState(chatID)
}

func isValidPhoneNumber(phone string) bool {
	if len(phone) < 10 {
		return false
	}
	
	// Проверяем, что номер начинается с + и содержит только цифры после +
	if !strings.HasPrefix(phone, "+") {
		return false
	}
	
	// Проверяем, что после + только цифры
	for _, c := range phone[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	
	return true
}
