package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// BOT KEYBOARDS

func (b *Bot) CreateDateSelectionKeyboard() tgbotapi.ReplyKeyboardMarkup {
    return tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Сегодня"),
            tgbotapi.NewKeyboardButton("Завтра"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Выбрать дату вручную"),
        ),
    )
}

func (b *Bot) CreateConfirmationKeyboard() tgbotapi.ReplyKeyboardMarkup {
    return tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("🔁 Сменить дату"),
            tgbotapi.NewKeyboardButton("✅ Подтвердить заказ"),
        ),
    )
}

func (b *Bot) CreateContactRequestKeyboard() tgbotapi.ReplyKeyboardMarkup {
    return tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButtonContact("📱 Отправить контакт"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Ввести вручную"),
        ),
    )
}
