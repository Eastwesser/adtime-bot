package keyboards

import (
	"adtime-bot/internal/storage"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Аксессуары из кожи"),
			tgbotapi.NewKeyboardButton("Типография"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Печать наклеек"),
		),
	)
}

func CreateMainMenuKeyboardAgreedTPA() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🛍 Новый заказ"),
			tgbotapi.NewKeyboardButton("📋 Мои заказы"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✏️ Изменить номер"),
			tgbotapi.NewKeyboardButton("ℹ️ Помощь"),
		),
	)
}

func CreateConfirmationKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔁 Сменить дату"),
			tgbotapi.NewKeyboardButton("✅ Подтвердить заказ"),
		),
	)
}

func CreateContactRequestKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Отправить контакт"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Ввести вручную"),
		),
	)
}

func CreatePrivacyAgreementKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Продолжить"),
		),
	)
}

func CreateOrderInitKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Оформить заказ"),
		),
	)
}

func CreateDateSelectionKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Сегодня"),
			tgbotapi.NewKeyboardButton("Завтра"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выбрать дату вручную"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
}

func CreateDimensionsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("30 40"),
			tgbotapi.NewKeyboardButton("50 40"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
}

func CreateDateConfirmationKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔁 Сменить дату"),
			tgbotapi.NewKeyboardButton("✅ Подтвердить дату"),
		),
	)
}

func CreatePhoneInputKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Отправить контакт"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Ввести вручную"),
		),
	)
}

func CreateServiceTypeKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Натуральная кожа"),
			tgbotapi.NewKeyboardButton("Искусственная кожа"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Замша"),
			tgbotapi.NewKeyboardButton("Другая текстура"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
}

func CreateTextureSelectionKeyboard(textures []storage.Texture) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	const maxButtonsPerRow = 2

	if len(textures) == 0 {
		return tgbotapi.NewInlineKeyboardMarkup() // Return empty keyboard if no textures
	}

	// Group textures into rows
	for i := 0; i < len(textures); i += maxButtonsPerRow {
		end := min(i+maxButtonsPerRow, len(textures))

		var row []tgbotapi.InlineKeyboardButton
		for _, texture := range textures[i:end] {
			btn := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s (%.2f₽/дм²)", texture.Name, texture.PricePerDM2),
				fmt.Sprintf("texture:%s", texture.ID),
			)
			row = append(row, btn)
		}
		rows = append(rows, row)
	}

	// Add cancel button
	cancelBtn := tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel")
	rows = append(rows, []tgbotapi.InlineKeyboardButton{cancelBtn})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func CreatePrintingMenuKeyboard(page int) tgbotapi.ReplyKeyboardMarkup {
	products := []string{"Визитки", "Бирки", "Листовки", "Буклеты", "Каталоги", "Календари", "Открытки"}
	return CreatePagedKeyboard(products, page, 4)
}

func CreateVinylServicesKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Печать на пленке"),
			tgbotapi.NewKeyboardButton("Резка пленки"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Ламинация"),
			tgbotapi.NewKeyboardButton("Комплекс"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("В главное меню"),
		),
	)
}

func CreateOptionsKeyboard(options []string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	for i := 0; i < len(options); i += 2 {
		row := make([]tgbotapi.KeyboardButton, 0)
		if i < len(options) {
			row = append(row, tgbotapi.NewKeyboardButton(options[i]))
		}
		if i+1 < len(options) {
			row = append(row, tgbotapi.NewKeyboardButton(options[i+1]))
		}
		rows = append(rows, row)
	}

	rows = append(rows, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Назад"),
		tgbotapi.NewKeyboardButton("Рассчитать"),
	})

	return tgbotapi.NewReplyKeyboard(rows...)
}

func CreatePagedKeyboard(items []string, page, itemsPerPage int) tgbotapi.ReplyKeyboardMarkup {
	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage
	if end > len(items) {
		end = len(items)
	}

	var rows [][]tgbotapi.KeyboardButton
	currentRow := make([]tgbotapi.KeyboardButton, 0, 2)

	for _, item := range items[start:end] {
		btn := tgbotapi.NewKeyboardButton(item)
		currentRow = append(currentRow, btn)

		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = make([]tgbotapi.KeyboardButton, 0, 2)
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	navRow := make([]tgbotapi.KeyboardButton, 0)
	if page > 1 {
		navRow = append(navRow, tgbotapi.NewKeyboardButton("Назад"))
	}
	if end < len(items) {
		navRow = append(navRow, tgbotapi.NewKeyboardButton("Далее"))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	rows = append(rows, []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("В главное меню"),
	})

	return tgbotapi.NewReplyKeyboard(rows...)
}

func CreateVinylOptionsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("0.5 м²"),
			tgbotapi.NewKeyboardButton("1 м²"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("2 м²"),
			tgbotapi.NewKeyboardButton("5 м²"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Черно-белое"),
			tgbotapi.NewKeyboardButton("Цветное"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
			tgbotapi.NewKeyboardButton("Рассчитать"),
		),
	)
}
