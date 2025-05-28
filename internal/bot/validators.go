package bot

import (
	"adtime-bot/internal/storage"
	"fmt"
	"strings"
)

func CalculatePrice(widthCm, heightCm int, pricePerDM2 float64) float64 {
    widthDM := float64(widthCm) / 10
    heightDM := float64(heightCm) / 10
    area := widthDM * heightDM
    return area * pricePerDM2
}

func IsValidPhoneNumber(phone string) bool {
	if len(phone) < 10 {
		return false
	}

	if !strings.HasPrefix(phone, "+") {
		return false
	}

	for _, c := range phone[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func FormatOrderNotification(order storage.Order) string {
    return fmt.Sprintf(
        "📦 Новый заказ #%d\n\n"+
            "Размеры: %d x %d см\n"+
            "Текстура: %s (%.2f₽/дм²)\n"+
            "Цена: %.2f руб\n"+
            "Контакт: %s\n"+
            "Статус: %s\n"+
            "Дата: %s",
        order.ID,
        order.WidthCM,
        order.HeightCM,
        order.TextureName,
        order.PricePerDM2,
        order.TotalPrice,
        order.Contact,
        order.Status,
        order.CreatedAt.Format("02.01.2006 15:04"),
    )
}
