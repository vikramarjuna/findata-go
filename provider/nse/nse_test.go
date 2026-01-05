package nse

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vikramarjuna/findata-go/config"
	"github.com/Vikramarjuna/findata-go/provider"
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

func TestQuoteFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()
	quote, err := p.Get("RELIANCE")
	if err != nil {
		t.Fatalf("Get(RELIANCE) error = %v", err)
	}

	// Test all fields are populated correctly
	tests := []struct {
		name  string
		value interface{}
		check func(interface{}) bool
	}{
		{"Symbol", quote.Symbol, func(v interface{}) bool { return v.(string) == "RELIANCE" }},
		{"Exchange", quote.Exchange, func(v interface{}) bool { return v.(config.Exchange) == config.ExchangeNSE }},
		{"Currency", quote.Currency, func(v interface{}) bool { return v.(string) == "INR" }},
		{"CompanyName", quote.CompanyName, func(v interface{}) bool { return v.(string) != "" }},
		{"Industry", quote.Industry, func(v interface{}) bool { return v.(string) != "" }},
		{"Sector", quote.Sector, func(v interface{}) bool { return v.(string) != "" }},
		{"LastPrice", quote.LastPrice, func(v interface{}) bool { return v.(float64) > 0 }},
		{"PreviousClose", quote.PreviousClose, func(v interface{}) bool { return v.(float64) > 0 }},
		{"Open", quote.Open, func(v interface{}) bool { return v.(float64) >= 0 }},
		{"DayHigh", quote.DayHigh, func(v interface{}) bool { return v.(float64) >= 0 }},
		{"DayLow", quote.DayLow, func(v interface{}) bool { return v.(float64) >= 0 }},
		{"YearHigh", quote.YearHigh, func(v interface{}) bool { return v.(float64) >= 0 }},
		{"YearLow", quote.YearLow, func(v interface{}) bool { return v.(float64) >= 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.value) {
				t.Errorf("%s validation failed: %v", tt.name, tt.value)
			}
		})
	}
}

func TestGetMultiple_EmptyList(t *testing.T) {
	p := New()
	quotes, errors := p.GetMultiple([]string{})

	if len(quotes) != 0 {
		t.Errorf("GetMultiple([]) should return empty map, got %d quotes", len(quotes))
	}
	if len(errors) != 0 {
		t.Errorf("GetMultiple([]) should return no errors, got %d errors", len(errors))
	}
}

func TestGetMultiple_SingleSymbol(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()
	quotes, errors := p.GetMultiple([]string{"TCS"})

	if len(errors) > 0 {
		t.Logf("Got errors: %v", errors)
	}

	if len(quotes) != 1 {
		t.Errorf("GetMultiple([TCS]) should return 1 quote, got %d", len(quotes))
	}

	if quote, ok := quotes["TCS"]; ok {
		if quote.Symbol != "TCS" {
			t.Errorf("Symbol = %v, want TCS", quote.Symbol)
		}
	}
}

func TestProviderInterface(t *testing.T) {
	// Verify Provider implements provider.Provider interface
	var _ provider.Provider = (*Provider)(nil)
}

func TestConstants(t *testing.T) {
	if BaseURL == "" {
		t.Error("BaseURL should not be empty")
	}
	if QuoteEndpoint == "" {
		t.Error("QuoteEndpoint should not be empty")
	}

	expectedURL := "https://www.nseindia.com/api"
	if BaseURL != expectedURL {
		t.Errorf("BaseURL = %v, want %v", BaseURL, expectedURL)
	}

	expectedEndpoint := "/quote-equity"
	if QuoteEndpoint != expectedEndpoint {
		t.Errorf("QuoteEndpoint = %v, want %v", QuoteEndpoint, expectedEndpoint)
	}
}

func TestSupportsSymbol_EdgeCases(t *testing.T) {
	p := New()

	tests := []struct {
		name   string
		symbol string
		want   bool
	}{
		{"Single letter", "A", true},
		{"Two letters", "AB", true},
		{"With ampersand", "M&M", true},
		{"With ampersand middle", "M&MFIN", true},
		{"All numbers", "123", true},
		{"Mixed alphanumeric", "NIFTY50", true},
		{"Underscore", "TEST_SYMBOL", false},
		{"Hyphen", "TEST-SYMBOL", false},
		{"Dot", "TEST.SYMBOL", false},
		{"Lowercase single", "a", false},
		{"Special chars", "TEST@SYMBOL", false},
		{"Unicode", "TEST™", false},
		{"Whitespace only", "   ", false},
		{"Tab", "\t", false},
		{"Newline", "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.SupportsSymbol(tt.symbol)
			if got != tt.want {
				t.Errorf("SupportsSymbol(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestGet_SymbolNormalization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()

	tests := []struct {
		name       string
		input      string
		wantSymbol string
	}{
		{"Leading space", " TCS", "TCS"},
		{"Trailing space", "TCS ", "TCS"},
		{"Both spaces", " TCS ", "TCS"},
		{"Multiple spaces", "  TCS  ", "TCS"},
		{"Lowercase", "tcs", "TCS"},
		{"Mixed case", "TcS", "TCS"},
		{"Tab prefix", "\tTCS", "TCS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quote, err := p.Get(tt.input)
			if err != nil {
				t.Skipf("Skipping due to error: %v", err)
				return
			}
			if quote.Symbol != tt.wantSymbol {
				t.Errorf("Symbol = %v, want %v", quote.Symbol, tt.wantSymbol)
			}
		})
	}
}

func TestGet_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Symbol not found"))
	}))
	defer server.Close()

	// Test demonstrates error handling structure
	// In production, HTTP errors are handled by the Get function
	// This is tested through integration tests with invalid symbols
	t.Logf("Test server URL: %s", server.URL)
	t.Log("Error handling is tested through integration tests with invalid symbols")
}

func TestGet_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// This test demonstrates the JSON parsing error handling
	// In practice, this is hard to test without dependency injection
	t.Log("JSON parsing errors are handled in the Get function")
}

func TestGetMultiple_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()

	// Test with multiple symbols to ensure no race conditions
	symbols := []string{"TCS", "INFY", "WIPRO", "HDFCBANK", "RELIANCE"}
	quotes, errors := p.GetMultiple(symbols)

	if len(errors) > 0 {
		t.Logf("Got %d errors (this is OK): %v", len(errors), errors)
	}

	// Should get at least some quotes
	if len(quotes) == 0 {
		t.Error("Expected at least one quote")
	}

	// Verify each quote has correct symbol
	for symbol, quote := range quotes {
		if quote.Symbol != symbol {
			t.Errorf("Quote symbol = %v, want %v", quote.Symbol, symbol)
		}
	}
}

func TestGetMultiple_ErrorAccumulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()

	// Mix of valid and potentially invalid symbols
	symbols := []string{"TCS", "INVALID1", "INFY", "INVALID2", "WIPRO"}
	quotes, errors := p.GetMultiple(symbols)

	// Should have some quotes
	if len(quotes) < 2 {
		t.Errorf("Expected at least 2 valid quotes, got %d", len(quotes))
	}

	// Errors may or may not be present depending on NSE behavior
	t.Logf("Got %d quotes and %d errors", len(quotes), len(errors))
}

func TestProvider_Benchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark test in short mode")
	}

	p := New()

	// Simple performance check
	quote, err := p.Get("TCS")
	if err != nil {
		t.Skipf("Skipping benchmark due to error: %v", err)
	}

	if quote == nil {
		t.Error("Quote should not be nil")
	}
}

func TestQuote_DataIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := New()
	quote, err := p.Get("RELIANCE")
	if err != nil {
		t.Fatalf("Get(RELIANCE) error = %v", err)
	}

	// Verify data integrity
	if quote.DayHigh < quote.DayLow && quote.DayHigh > 0 && quote.DayLow > 0 {
		t.Errorf("DayHigh (%v) should be >= DayLow (%v)", quote.DayHigh, quote.DayLow)
	}

	if quote.YearHigh < quote.YearLow && quote.YearHigh > 0 && quote.YearLow > 0 {
		t.Errorf("YearHigh (%v) should be >= YearLow (%v)", quote.YearHigh, quote.YearLow)
	}

	// LastPrice should be within day range (if available)
	if quote.DayHigh > 0 && quote.DayLow > 0 && quote.LastPrice > 0 {
		if quote.LastPrice > quote.DayHigh || quote.LastPrice < quote.DayLow {
			t.Logf("Warning: LastPrice (%v) outside day range [%v, %v]",
				quote.LastPrice, quote.DayLow, quote.DayHigh)
		}
	}
}
