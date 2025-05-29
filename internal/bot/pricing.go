package bot

type PricingConfig struct {
    LeatherPricePerDM2    float64
    ProcessingCostPerDM2  float64
    PaymentCommissionRate float64 // 3% for Yookassa
    SalesTaxRate          float64 // 6% for СЗ
    MarkupMultiplier      float64
}

func NewDefaultPricing() PricingConfig {
    return PricingConfig{
        LeatherPricePerDM2:    25.0,
        ProcessingCostPerDM2:  31.25, // 1000₽/3200cm² = 0.3125₽/cm² → 31.25₽/dm²
        PaymentCommissionRate: 0.03,
        SalesTaxRate:          0.06,
        MarkupMultiplier:      2.5, // Empirical from your table
    }
}

func CalculatePrice(widthCm, heightCm int, cfg PricingConfig) (priceDetails map[string]float64) {
    areaCm2 := float64(widthCm * heightCm)
    areaDm2 := areaCm2 / 100
    
    priceDetails = make(map[string]float64)
    
    // Base costs
    priceDetails["leather_cost"] = areaDm2 * cfg.LeatherPricePerDM2
    priceDetails["processing_cost"] = areaDm2 * cfg.ProcessingCostPerDM2
    priceDetails["total_cost"] = priceDetails["leather_cost"] + priceDetails["processing_cost"]
    
    // Final price with markup
    priceDetails["final_price"] = priceDetails["total_cost"] * cfg.MarkupMultiplier
    
    // Revenue calculations
    priceDetails["commission"] = priceDetails["final_price"] * cfg.PaymentCommissionRate
    priceDetails["tax"] = priceDetails["final_price"] * cfg.SalesTaxRate
    priceDetails["net_revenue"] = priceDetails["final_price"] - priceDetails["commission"] - priceDetails["tax"]
    priceDetails["profit"] = priceDetails["net_revenue"] - priceDetails["total_cost"]
    
    return priceDetails
}
