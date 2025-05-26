package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// BOT KEYBOARDS

func (b *Bot) createMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🛍️ Создать заказ"),
			tgbotapi.NewKeyboardButton("ℹ️ О нас"),
		),
	)
}

func (b *Bot) createOrderKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)
}

func (b *Bot) createContactKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Отправить контакт"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад"),
		),
	)
}
