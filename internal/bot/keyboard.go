package bot

import (
	"adtime-bot/internal/storage"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) CreateMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
			tgbotapi.NewKeyboardButton("/help"),
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

func (b *Bot) CreatePrivacyAgreementKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Продолжить"),
		),
	)
}

func (b *Bot) CreateOrderInitKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Оформить заказ"),
		),
	)
}

func (b *Bot) CreateDateSelectionKeyboard() tgbotapi.ReplyKeyboardMarkup {
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

func (b *Bot) CreateDimensionsKeyboard() tgbotapi.ReplyKeyboardMarkup {
    return tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("⬅️ Назад"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Пример: 30 40"),
        ),
    )
}

func (b *Bot) CreateDateConfirmationKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔁 Сменить дату"),
			tgbotapi.NewKeyboardButton("✅ Подтвердить дату"),
		),
	)
}

func (b *Bot) CreatePhoneInputKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Отправить контакт"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Ввести вручную"),
		),
	)
}

func (b *Bot) CreateServiceTypeKeyboard() tgbotapi.ReplyKeyboardMarkup {
    return tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("🐄 Натуральная кожа"),
            tgbotapi.NewKeyboardButton("🔄 Искусственная кожа"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("🧶 Замша"),
            tgbotapi.NewKeyboardButton("🎨 Другая текстура"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("❌ Отмена"),
        ),
    )
}

func (b *Bot) CreateTextureSelectionKeyboard(textures []storage.Texture) tgbotapi.InlineKeyboardMarkup {
    var rows [][]tgbotapi.InlineKeyboardButton
    const maxButtonsPerRow = 2

	if len(textures) == 0 {
        return tgbotapi.NewInlineKeyboardMarkup() // Return empty keyboard if no textures
    }
    
    // Group textures into rows
    for i := 0; i < len(textures); i += maxButtonsPerRow {
        end := i + maxButtonsPerRow
        if end > len(textures) {
            end = len(textures)
        }
        
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

// For future pagination
// func (b *Bot) CreateTextureSelectionKeyboard(textures []storage.Texture, page int) tgbotapi.InlineKeyboardMarkup {
//     const itemsPerPage = 6
//     start := page * itemsPerPage
//     end := start + itemsPerPage
    
//     if end > len(textures) {
//         end = len(textures)
//     }

//     // ... create buttons for current page ...

//     // Add navigation buttons
//     var navButtons []tgbotapi.InlineKeyboardButton
//     if page > 0 {
//         navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Пред.", fmt.Sprintf("texture_page:%d", page-1)))
//     }
//     if end < len(textures) {
//         navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("След. ➡️", fmt.Sprintf("texture_page:%d", page+1)))
//     }
    
//     // Add to rows if we have navigation
//     if len(navButtons) > 0 {
//         rows = append(rows, navButtons)
//     }
    
//     // Add cancel button
//     rows = append(rows, []tgbotapi.InlineKeyboardButton{
//         tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
//     })
    
//     return tgbotapi.NewInlineKeyboardMarkup(rows...)
// }