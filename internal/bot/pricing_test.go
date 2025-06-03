package bot

import "testing"

func TestCalculatePrice(t *testing.T) {
    cfg := PricingConfig{
        LeatherPricePerDM2:    25.0,
        ProcessingCostPerDM2:  31.25,
        PaymentCommissionRate: 0.03,
        SalesTaxRate:          0.06,
        MarkupMultiplier:      2.5,
    }
    
    prices, err := CalculatePrice(80, 20, cfg)
    if err != nil {
        t.Fatalf("CalculatePrice failed: %v", err)
    }
    
    expectedLeatherCost := 400.0 // 80*20/100*25
    if prices["leather_cost"] != expectedLeatherCost {
        t.Errorf("Incorrect leather cost, got %.2f, want %.2f", 
            prices["leather_cost"], expectedLeatherCost)
    }
    
    // // Add more test cases for other calculations
    // expectedFinalPrice := 1406.25 // (400 + 500) * 2.5
    // if prices["final_price"] != expectedFinalPrice {
    //     t.Errorf("Incorrect final price, got %.2f, want %.2f",
    //         prices["final_price"], expectedFinalPrice)
    // }
}

func TestCalculatePrice_InvalidInput(t *testing.T) {
    cfg := PricingConfig{
        LeatherPricePerDM2:    -1.0, // Invalid price
        ProcessingCostPerDM2:  31.25,
        PaymentCommissionRate: 0.03,
        SalesTaxRate:          0.06,
        MarkupMultiplier:      2.5,
    }
    
    _, err := CalculatePrice(80, 20, cfg)
    if err == nil {
        t.Error("Expected error for invalid leather price, got nil")
    }
}