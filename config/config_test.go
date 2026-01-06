package config

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultMarket(t *testing.T) {
	// Save original
	original := GetDefaultMarket()
	defer SetDefaultMarket(original)

	tests := []struct {
		name   string
		market Market
	}{
		{"Set to India", MarketIndia},
		{"Set to US", MarketUS},
		{"Set to Auto", MarketAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultMarket(tt.market)
			got := GetDefaultMarket()
			if got != tt.market {
				t.Errorf("GetDefaultMarket() = %v, want %v", got, tt.market)
			}
		})
	}
}

func TestHTTPClient(t *testing.T) {
	// Save original
	original := GetHTTPClient()
	defer SetHTTPClient(original)

	// Create custom client
	customClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	SetHTTPClient(customClient)
	got := GetHTTPClient()

	if got.Timeout != customClient.Timeout {
		t.Errorf("GetHTTPClient().Timeout = %v, want %v", got.Timeout, customClient.Timeout)
	}
}

func TestAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		key      string
	}{
		{"AlphaVantage", "alphavantage", "test-key-123"},
		{"Finnhub", "finnhub", "test-key-456"},
		{"IEX", "iex", "test-key-789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetAPIKey(tt.provider, tt.key)
			got, ok := GetAPIKey(tt.provider)
			if !ok {
				t.Errorf("GetAPIKey(%s) returned ok=false", tt.provider)
			}
			if got != tt.key {
				t.Errorf("GetAPIKey(%s) = %v, want %v", tt.provider, got, tt.key)
			}
		})
	}

	// Test non-existent key
	_, ok := GetAPIKey("nonexistent")
	if ok {
		t.Error("GetAPIKey(nonexistent) should return ok=false")
	}
}

func TestUserAgent(t *testing.T) {
	// Save original
	original := GetUserAgent()
	defer SetUserAgent(original)

	customUA := "my-custom-agent/2.0"
	SetUserAgent(customUA)
	got := GetUserAgent()

	if got != customUA {
		t.Errorf("GetUserAgent() = %v, want %v", got, customUA)
	}
}

func TestDefaultValues(t *testing.T) {
	// Test that defaults are set correctly
	market := GetDefaultMarket()
	if market != MarketIndia {
		t.Errorf("Default market should be MarketIndia, got %v", market)
	}

	client := GetHTTPClient()
	if client == nil {
		t.Fatal("Default HTTP client should not be nil")
	}
	if client.Timeout != 30*time.Second {
		t.Errorf("Default timeout should be 30s, got %v", client.Timeout)
	}

	ua := GetUserAgent()
	if ua == "" {
		t.Error("Default user agent should not be empty")
	}
}

func TestConcurrentAccess(_ *testing.T) {
	// Test thread-safety
	done := make(chan bool)

	// Concurrent reads and writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			if n%2 == 0 {
				SetDefaultMarket(MarketIndia)
			} else {
				GetDefaultMarket()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMarketConstants(t *testing.T) {
	tests := []struct {
		name   string
		market Market
		want   string
	}{
		{"India", MarketIndia, "IN"},
		{"US", MarketUS, "US"},
		{"Auto", MarketAuto, "AUTO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.market) != tt.want {
				t.Errorf("Market %s = %v, want %v", tt.name, tt.market, tt.want)
			}
		})
	}
}

func TestExchangeConstants(t *testing.T) {
	tests := []struct {
		name     string
		exchange Exchange
		want     string
	}{
		{"NSE", ExchangeNSE, "NSE"},
		{"BSE", ExchangeBSE, "BSE"},
		{"NYSE", ExchangeNYSE, "NYSE"},
		{"NASDAQ", ExchangeNASDAQ, "NASDAQ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.exchange) != tt.want {
				t.Errorf("Exchange %s = %v, want %v", tt.name, tt.exchange, tt.want)
			}
		})
	}
}
