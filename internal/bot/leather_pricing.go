package bot

import (
	"adtime-bot/internal/config"
	"fmt"
)

type PricingConfig struct {
    LeatherPricePerDM2    float64
    ProcessingCostPerDM2  float64
    PaymentCommissionRate float64 // 3% for Yookassa
    SalesTaxRate          float64 // 6% for СЗ
    MarkupMultiplier      float64
}

func NewPricingConfig(texturePrice float64, cfg *config.Config) PricingConfig {
    return PricingConfig{
        LeatherPricePerDM2:    texturePrice,
        ProcessingCostPerDM2:  cfg.Pricing.ProcessingCostPerDM2,
        PaymentCommissionRate: cfg.Pricing.PaymentCommissionRate,
        SalesTaxRate:          cfg.Pricing.SalesTaxRate,
        MarkupMultiplier:      cfg.Pricing.MarkupMultiplier,
    }
}

func CalculatePrice(widthCm, heightCm int, cfg PricingConfig) (map[string]float64, error) {
    
    if cfg.LeatherPricePerDM2 <= 0 {
        return nil, fmt.Errorf("invalid leather price: %.2f", cfg.LeatherPricePerDM2)
    }
    if cfg.ProcessingCostPerDM2 < 0 {
        return nil, fmt.Errorf("invalid processing cost: %.2f", cfg.ProcessingCostPerDM2)
    }
    if cfg.MarkupMultiplier < 1 {
        return nil, fmt.Errorf("invalid markup multiplier: %.2f", cfg.MarkupMultiplier)
    }

    areaCm2 := float64(widthCm * heightCm)
    areaDm2 := areaCm2 / 100
    
    priceDetails := make(map[string]float64)
    
    // Base costs. Use texture price from database
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
    
    return priceDetails, nil
}
