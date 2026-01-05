package equity

import (
	"testing"

	"github.com/Vikramarjuna/findata-go/config"
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

	if quote.Exchange != config.ExchangeNSE {
		t.Errorf("Expected exchange NSE, got %s", quote.Exchange)
	}

	if quote.Currency != "INR" {
		t.Errorf("Expected currency INR, got %s", quote.Currency)
	}

	if quote.LastPrice <= 0 {
		t.Errorf("Expected positive last price, got %.2f", quote.LastPrice)
	}

	if quote.CompanyName == "" {
		t.Error("Expected company name to be set")
	}
}

func TestGetWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with explicit exchange
	quote, err := Get("TCS", WithExchange(config.ExchangeNSE))
	if err != nil {
		t.Fatalf("Failed to get quote: %v", err)
	}

	if quote.Symbol != "TCS" {
		t.Errorf("Expected symbol TCS, got %s", quote.Symbol)
	}

	if quote.Exchange != config.ExchangeNSE {
		t.Errorf("Expected exchange NSE, got %s", quote.Exchange)
	}

	// Test with explicit market
	quote2, err := Get("INFY", WithMarket(config.MarketIndia))
	if err != nil {
		t.Fatalf("Failed to get quote: %v", err)
	}

	if quote2.Symbol != "INFY" {
		t.Errorf("Expected symbol INFY, got %s", quote2.Symbol)
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
		if quote.Exchange != config.ExchangeNSE {
			t.Errorf("Expected exchange NSE for %s, got %s", symbol, quote.Exchange)
		}
		if quote.Currency != "INR" {
			t.Errorf("Expected currency INR for %s, got %s", symbol, quote.Currency)
		}
	}
}

func TestDefaultMarket(t *testing.T) {
	// Save original default
	original := config.GetDefaultMarket()
	defer config.SetDefaultMarket(original)

	// Set to India
	config.SetDefaultMarket(config.MarketIndia)
	if config.GetDefaultMarket() != config.MarketIndia {
		t.Error("Failed to set default market to India")
	}

	// Set to US
	config.SetDefaultMarket(config.MarketUS)
	if config.GetDefaultMarket() != config.MarketUS {
		t.Error("Failed to set default market to US")
	}
}

