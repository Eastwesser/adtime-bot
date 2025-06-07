package calculators

type PrintingCalculator struct {
    BasePrices map[string]float64
}

func NewPrintingCalculator() *PrintingCalculator {
    return &PrintingCalculator{
        BasePrices: map[string]float64{
            "Визитки":   10.00,
            "Бирки":     11.00,
            "Листовки":  12.00,
            "Буклеты":   13.00,
            "Каталоги":  14.00,
            "Календари": 15.00,
            "Открытки":  16.00,
        },
    }
}

func (pc *PrintingCalculator) Calculate(product string, quantity int, options map[string]interface{}) float64 {
    basePrice, exists := pc.BasePrices[product]
    if !exists {
        return 0.00
    }

    // Примерные расчеты
    total := basePrice * float64(quantity)
    
    if paper, ok := options["paper_type"]; ok && paper == "премиум" {
        total *= 1.2
    }
    
    if quantity > 500 {
        total *= 0.9 // скидка 10% на большие тиражи
    }
    
    return total
}
