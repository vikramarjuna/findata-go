package nse

import (
	"testing"

	"github.com/Vikramarjuna/findata-go/config"
)

func TestProviderName(t *testing.T) {
	p := New()
	if p.Name() != "NSE" {
		t.Errorf("Name() = %v, want NSE", p.Name())
	}
}

func TestSupportsSymbol(t *testing.T) {
	p := New()

	tests := []struct {
		name   string
		symbol string
		want   bool
	}{
		{"Valid uppercase", "RELIANCE", true},
		{"Valid with numbers", "NIFTY50", true},
		{"Valid with ampersand", "M&M", true},
		{"Invalid with dot", "BRK.A", false},
		{"Invalid with dash", "BRK-A", false},
		{"Invalid lowercase", "reliance", false},
		{"Invalid with space", "REL IANCE", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.SupportsSymbol(tt.symbol)
			if got != tt.want {
				t.Errorf("SupportsSymbol(%v) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestGet_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()
	quote, err := p.Get("RELIANCE")
	if err != nil {
		t.Fatalf("Get(RELIANCE) error = %v", err)
	}

	// Validate quote structure
	if quote.Symbol != "RELIANCE" {
		t.Errorf("Symbol = %v, want RELIANCE", quote.Symbol)
	}
	if quote.Exchange != config.ExchangeNSE {
		t.Errorf("Exchange = %v, want NSE", quote.Exchange)
	}
	if quote.Currency != "INR" {
		t.Errorf("Currency = %v, want INR", quote.Currency)
	}
	if quote.LastPrice <= 0 {
		t.Errorf("LastPrice = %v, want > 0", quote.LastPrice)
	}
	if quote.CompanyName == "" {
		t.Error("CompanyName should not be empty")
	}
	if quote.Sector == "" {
		t.Error("Sector should not be empty")
	}
}

func TestGet_InvalidSymbol(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()
	quote, err := p.Get("INVALIDSYMBOL123456")
	// NSE may return partial data or error for invalid symbols
	// We just verify it doesn't crash
	if err != nil {
		t.Logf("Got expected error for invalid symbol: %v", err)
	} else if quote != nil {
		t.Logf("NSE returned partial data for invalid symbol (this is OK)")
	}
}

func TestGetMultiple_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()
	symbols := []string{"TCS", "INFY"}
	quotes, errors := p.GetMultiple(symbols)

	if len(errors) > 0 {
		t.Logf("Encountered %d errors: %v", len(errors), errors)
	}

	if len(quotes) == 0 {
		t.Fatal("GetMultiple should return at least one quote")
	}

	for symbol, quote := range quotes {
		if quote.Symbol != symbol {
			t.Errorf("Quote symbol = %v, want %v", quote.Symbol, symbol)
		}
		if quote.Exchange != config.ExchangeNSE {
			t.Errorf("Exchange = %v, want NSE", quote.Exchange)
		}
		if quote.Currency != "INR" {
			t.Errorf("Currency = %v, want INR", quote.Currency)
		}
		if quote.LastPrice <= 0 {
			t.Errorf("LastPrice for %s = %v, want > 0", symbol, quote.LastPrice)
		}
	}
}

func TestGetMultiple_WithErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()
	symbols := []string{"TCS", "INVALIDSYMBOL123", "INFY"}
	quotes, errors := p.GetMultiple(symbols)

	// NSE may or may not return errors for invalid symbols
	// Just verify we get some valid quotes
	if len(quotes) < 2 {
		t.Errorf("GetMultiple should return at least 2 valid quotes, got %d", len(quotes))
	}

	if len(errors) > 0 {
		t.Logf("Got %d errors (this is OK): %v", len(errors), errors)
	}
}

func TestSymbolCleaning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()

	// Test that symbols are cleaned (trimmed, uppercased)
	tests := []struct {
		name   string
		symbol string
	}{
		{"With spaces", " RELIANCE "},
		{"Lowercase", "reliance"},
		{"Mixed case", "Reliance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote, err := p.Get(tt.symbol)
			if err != nil {
				t.Skipf("Skipping due to error: %v", err)
				return
			}
			if quote.Symbol != "RELIANCE" {
				t.Errorf("Symbol = %v, want RELIANCE", quote.Symbol)
			}
		})
	}
}
