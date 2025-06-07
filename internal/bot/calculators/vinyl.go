package calculators

type VinylCalculator struct {
    ServicePrices map[string]float64
}

func NewVinylCalculator() *VinylCalculator {
    return &VinylCalculator{
        ServicePrices: map[string]float64{
            "Печать на пленке": 100.00, // за м²
            "Резка пленки":     50.00,
            "Ламинация":        70.00,
            "Комплекс":         200.00,
        },
    }
}

func (vc *VinylCalculator) Calculate(service string, area float64, options map[string]interface{}) float64 {
    basePrice, exists := vc.ServicePrices[service]
    if !exists {
        return 0.00
    }

    total := basePrice * area
    
    if colors, ok := options["colors"]; ok && colors == "full" {
        total *= 1.15 // наценка 15% за цветную печать
    }
    
    if area > 5 {
        total *= 0.95 // скидка 5% на большие объемы
    }
    
    return total
}
