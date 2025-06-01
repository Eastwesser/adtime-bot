package bot

import (
	"adtime-bot/internal/storage"
	"fmt"
	"strings"
	"unicode"
)

func NormalizePhoneNumber(phone string) string {
    // Remove all non-digit characters
    cleaned := strings.Map(func(r rune) rune {
        if unicode.IsDigit(r) {
            return r
        }
        return -1
    }, phone)

    // Add +7 for Russian numbers if no country code exists
    if strings.HasPrefix(cleaned, "7") && len(cleaned) == 11 {
        return "+" + cleaned
    }
    if strings.HasPrefix(cleaned, "8") && len(cleaned) == 11 {
        return "+7" + cleaned[1:]
    }
    if strings.HasPrefix(cleaned, "9") && len(cleaned) == 10 {
        return "+7" + cleaned
    }
    
    // For international numbers, preserve the + if it was there
    if strings.HasPrefix(phone, "+") {
        return "+" + cleaned
    }
    
    
    return cleaned
}

func IsValidPhoneNumber(phone string) bool {

    normalized := NormalizePhoneNumber(phone)
    
    // Russian numbers
    if strings.HasPrefix(normalized, "+7") && len(normalized) == 12 {
        return true
    }
    
    // International numbers
    if strings.HasPrefix(normalized, "+") && len(normalized) >= 10 {
        return true
    }

    return false
}

func FormatPhoneNumber(phone string) string {
    // Format as +7 (XXX) XXX-XX-XX for Russian numbers
    if strings.HasPrefix(phone, "+7") && len(phone) == 12 {
        return fmt.Sprintf("%s (%s) %s-%s-%s", 
            phone[:2],
            phone[2:5],
            phone[5:8],
            phone[8:10],
            phone[10:12])
    }
    return phone
}

func FormatOrderNotification(order storage.Order) string {
    return fmt.Sprintf(
        "📦 Новый заказ #%d\n\n"+
            "Размеры: %d x %d см\n"+
            "Текстура: %s\n"+
            "Итоговая цена: %.2f руб\n"+
            "──────────────────\n"+
            "Детали расчета:\n"+
            "- Стоимость кожи: %.2f руб\n"+
            "- Обработка: %.2f руб\n"+
            "- Комиссия: %.2f руб\n"+
            "- Налог: %.2f руб\n"+
            "Чистая выручка: %.2f руб\n"+
            "Прибыль: %.2f руб\n"+
            "──────────────────\n"+
            "Контакт: %s\n"+
            "Статус: %s\n"+
            "Дата: %s",
        order.ID,
        order.WidthCM,
        order.HeightCM,
        order.TextureName,
        order.Price,
        order.LeatherCost,
        order.ProcessCost,
        order.Commission,
        order.Tax,
        order.NetRevenue,
        order.Profit,
        order.Contact,
        order.Status,
        order.CreatedAt.Format("02.01.2006 15:04"),
    )
}

func FormatPriceBreakdown(width, height int, prices map[string]float64) string {
    return fmt.Sprintf(
        `
            📏 Размер: %d×%d см
            💵 Итоговая цена: %.2f₽

            📊 Детали расчета:
            - Стоимость кожи: %.2f₽
            - Обработка: %.2f₽
            - Комиссия платежа (3%%): %.2f₽
            - Налог (6%%): %.2f₽
            ────────────────────
            Чистая выручка: %.2f₽
            Прибыль: %.2f₽
        `,
        width, height,
        prices["final_price"],
        prices["leather_cost"],
        prices["processing_cost"],
        prices["commission"],
        prices["tax"],
        prices["net_revenue"],
        prices["profit"],
    )
}

func FormatSimplePriceBreakdown(width, height int, finalPrice float64) string {
    return fmt.Sprintf(
        `📏 Размер: %d×%d см
        💰 Итоговая цена: %.2f₽

        Нажмите "Подтвердить" для оформления заказа`,
        width, height,
        finalPrice,
    )
}
