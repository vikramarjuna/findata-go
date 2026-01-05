package provider

import (
	"testing"

	"github.com/Vikramarjuna/findata-go/config"
)

func TestQuoteStructure(t *testing.T) {
	quote := &Quote{
		Symbol:        "TEST",
		Exchange:      config.ExchangeNSE,
		CompanyName:   "Test Company",
		Industry:      "Technology",
		Sector:        "IT Services",
		LastPrice:     100.50,
		Currency:      "INR",
		Change:        2.50,
		PChange:       2.55,
		PreviousClose: 98.00,
		Open:          99.00,
		DayHigh:       101.00,
		DayLow:        98.50,
		YearHigh:      120.00,
		YearLow:       80.00,
		Volume:        1000000,
		Value:         100500000,
		Indices:       []string{"NIFTY 50", "NIFTY IT"},
		Metadata:      map[string]string{"key": "value"},
	}

	// Test all fields are set correctly
	if quote.Symbol != "TEST" {
		t.Errorf("Symbol = %v, want TEST", quote.Symbol)
	}
	if quote.Exchange != config.ExchangeNSE {
		t.Errorf("Exchange = %v, want NSE", quote.Exchange)
	}
	if quote.Currency != "INR" {
		t.Errorf("Currency = %v, want INR", quote.Currency)
	}
	if quote.LastPrice != 100.50 {
		t.Errorf("LastPrice = %v, want 100.50", quote.LastPrice)
	}
	if len(quote.Indices) != 2 {
		t.Errorf("Indices length = %v, want 2", len(quote.Indices))
	}
	if quote.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want value", quote.Metadata["key"])
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		wantMsg string
	}{
		{
			name: "Error with code",
			err: &Error{
				Message:  "test error",
				Code:     404,
				Provider: "TestProvider",
			},
			wantMsg: "TestProvider: test error (code: 404)",
		},
		{
			name: "Error without code",
			err: &Error{
				Message:  "test error",
				Provider: "TestProvider",
			},
			wantMsg: "TestProvider: test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestQuoteValidation(t *testing.T) {
	// Test that Quote can handle nil/empty values
	quote := &Quote{}

	if quote.Symbol != "" {
		t.Errorf("Empty quote should have empty symbol")
	}
	if quote.LastPrice != 0 {
		t.Errorf("Empty quote should have zero price")
	}
	if quote.Indices != nil {
		t.Errorf("Empty quote should have nil indices")
	}
	if quote.Metadata != nil {
		t.Errorf("Empty quote should have nil metadata")
	}
}

func TestQuoteWithPartialData(t *testing.T) {
	// Test quote with only essential fields
	quote := &Quote{
		Symbol:    "PARTIAL",
		Exchange:  config.ExchangeNSE,
		LastPrice: 50.00,
		Currency:  "INR",
	}

	if quote.Symbol != "PARTIAL" {
		t.Errorf("Symbol = %v, want PARTIAL", quote.Symbol)
	}
	if quote.CompanyName != "" {
		t.Errorf("CompanyName should be empty")
	}
	if quote.Volume != 0 {
		t.Errorf("Volume should be zero")
	}
}

func TestMultipleExchanges(t *testing.T) {
	exchanges := []config.Exchange{
		config.ExchangeNSE,
		config.ExchangeBSE,
		config.ExchangeNYSE,
		config.ExchangeNASDAQ,
	}

	for _, exchange := range exchanges {
		quote := &Quote{
			Symbol:   "TEST",
			Exchange: exchange,
		}

		if quote.Exchange != exchange {
			t.Errorf("Exchange = %v, want %v", quote.Exchange, exchange)
		}
	}
}

func TestQuoteMetadata(t *testing.T) {
	quote := &Quote{
		Symbol:   "TEST",
		Metadata: make(map[string]string),
	}

	// Add metadata
	quote.Metadata["sector"] = "Technology"
	quote.Metadata["industry"] = "Software"

	if quote.Metadata["sector"] != "Technology" {
		t.Errorf("Metadata[sector] = %v, want Technology", quote.Metadata["sector"])
	}
	if quote.Metadata["industry"] != "Software" {
		t.Errorf("Metadata[industry] = %v, want Software", quote.Metadata["industry"])
	}
}
