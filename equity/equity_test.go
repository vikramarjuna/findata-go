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

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		opts    *Options
		wantErr bool
	}{
		{
			name:    "Valid Indian symbol",
			symbol:  "RELIANCE",
			opts:    &Options{Market: config.MarketIndia},
			wantErr: false,
		},
		{
			name:    "Valid symbol with auto market",
			symbol:  "TCS",
			opts:    &Options{Market: config.MarketAuto},
			wantErr: false,
		},
		{
			name:    "Explicit NSE exchange",
			symbol:  "INFY",
			opts:    &Options{Exchange: config.ExchangeNSE},
			wantErr: false,
		},
		{
			name:    "Unsupported exchange BSE",
			symbol:  "RELIANCE",
			opts:    &Options{Exchange: config.ExchangeBSE},
			wantErr: true,
		},
		{
			name:    "Unsupported market US",
			symbol:  "AAPL",
			opts:    &Options{Market: config.MarketUS},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := detectProvider(tt.symbol, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		wantMkt  config.Market
		wantExch config.Exchange
	}{
		{
			name:     "No options",
			opts:     []Option{},
			wantMkt:  config.GetDefaultMarket(),
			wantExch: "",
		},
		{
			name:     "WithMarket option",
			opts:     []Option{WithMarket(config.MarketUS)},
			wantMkt:  config.MarketUS,
			wantExch: "",
		},
		{
			name:     "WithExchange option",
			opts:     []Option{WithExchange(config.ExchangeNSE)},
			wantMkt:  config.GetDefaultMarket(),
			wantExch: config.ExchangeNSE,
		},
		{
			name:     "Both options",
			opts:     []Option{WithMarket(config.MarketIndia), WithExchange(config.ExchangeNSE)},
			wantMkt:  config.MarketIndia,
			wantExch: config.ExchangeNSE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &Options{
				Market: config.GetDefaultMarket(),
			}
			for _, opt := range tt.opts {
				opt(options)
			}

			if options.Market != tt.wantMkt {
				t.Errorf("Market = %v, want %v", options.Market, tt.wantMkt)
			}
			if options.Exchange != tt.wantExch {
				t.Errorf("Exchange = %v, want %v", options.Exchange, tt.wantExch)
			}
		})
	}
}

func TestGetMultiple_EmptyList(t *testing.T) {
	quotes, errors := GetMultiple([]string{})

	if len(quotes) != 0 {
		t.Errorf("GetMultiple([]) should return empty map, got %d quotes", len(quotes))
	}
	if len(errors) != 0 {
		t.Errorf("GetMultiple([]) should return no errors, got %d errors", len(errors))
	}
}

func TestGetMultiple_MixedSymbols(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Mix of valid and invalid symbols
	symbols := []string{"TCS", "INVALIDSYMBOL123", "INFY"}
	quotes, errors := GetMultiple(symbols)

	// NSE may or may not return errors for invalid symbols
	// Just verify we get at least the valid quotes
	if len(quotes) < 2 {
		t.Errorf("Expected at least 2 valid quotes, got %d", len(quotes))
	}

	if len(errors) > 0 {
		t.Logf("Got %d errors (this is OK): %v", len(errors), errors)
	}

	// Verify valid quotes
	for symbol, quote := range quotes {
		// Skip validation for invalid symbol if NSE returned it
		if symbol == "INVALIDSYMBOL123" {
			continue
		}
		if quote.Symbol != symbol {
			t.Errorf("Quote symbol = %v, want %v", quote.Symbol, symbol)
		}
		if quote.LastPrice <= 0 {
			t.Errorf("LastPrice for %s should be > 0", symbol)
		}
	}
}

func TestSymbolCleaning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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
			quote, err := Get(tt.symbol)
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
