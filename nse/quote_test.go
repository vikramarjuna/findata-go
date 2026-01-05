package nse

import (
	"testing"
)

func TestGet(t *testing.T) {
	// This is an integration test that requires network access
	// Skip in CI/CD environments if needed
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	quote, err := Get("RELIANCE")
	if err != nil {
		t.Fatalf("Failed to get quote: %v", err)
	}

	if quote.Symbol != "RELIANCE" {
		t.Errorf("Expected symbol RELIANCE, got %s", quote.Symbol)
	}

	if quote.LastPrice <= 0 {
		t.Errorf("Expected positive last price, got %.2f", quote.LastPrice)
	}

	if quote.CompanyName == "" {
		t.Error("Expected company name to be set")
	}
}

func TestGetMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	symbols := []string{"TCS", "INFY"}
	quotes, errors := GetMultiple(symbols)

	if len(errors) > 0 {
		t.Logf("Encountered %d errors: %v", len(errors), errors)
	}

	if len(quotes) == 0 {
		t.Fatal("Expected at least one quote")
	}

	for symbol, quote := range quotes {
		if quote.Symbol != symbol {
			t.Errorf("Expected symbol %s, got %s", symbol, quote.Symbol)
		}
		if quote.LastPrice <= 0 {
			t.Errorf("Expected positive price for %s, got %.2f", symbol, quote.LastPrice)
		}
	}
}

