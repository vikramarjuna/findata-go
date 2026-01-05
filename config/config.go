// Package config provides configuration for findata-go
package config

import (
	"net/http"
	"sync"
	"time"
)

// Market represents a financial market
type Market string

const (
	// MarketIndia represents Indian markets (NSE, BSE)
	MarketIndia Market = "IN"
	// MarketUS represents US markets (NYSE, NASDAQ)
	MarketUS Market = "US"
	// MarketAuto automatically detects the market
	MarketAuto Market = "AUTO"
)

// Exchange represents a stock exchange
type Exchange string

const (
	// ExchangeNSE is the National Stock Exchange of India
	ExchangeNSE Exchange = "NSE"
	// ExchangeBSE is the Bombay Stock Exchange
	ExchangeBSE Exchange = "BSE"
	// ExchangeNYSE is the New York Stock Exchange
	ExchangeNYSE Exchange = "NYSE"
	// ExchangeNASDAQ is the NASDAQ exchange
	ExchangeNASDAQ Exchange = "NASDAQ"
)

// Config holds the global configuration
type Config struct {
	mu            sync.RWMutex
	defaultMarket Market
	httpClient    *http.Client
	apiKeys       map[string]string
	userAgent     string
}

var globalConfig = &Config{
	defaultMarket: MarketIndia, // Default to Indian market
	httpClient: &http.Client{
		Timeout: 30 * time.Second,
	},
	apiKeys:   make(map[string]string),
	userAgent: "findata-go/1.0",
}

// SetDefaultMarket sets the default market for symbol lookups
func SetDefaultMarket(market Market) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.defaultMarket = market
}

// GetDefaultMarket returns the default market
func GetDefaultMarket() Market {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.defaultMarket
}

// SetHTTPClient sets a custom HTTP client
func SetHTTPClient(client *http.Client) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.httpClient = client
}

// GetHTTPClient returns the HTTP client
func GetHTTPClient() *http.Client {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.httpClient
}

// SetAPIKey sets an API key for a provider
// provider can be "alphavantage", "finnhub", "iex", etc.
func SetAPIKey(provider, key string) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.apiKeys[provider] = key
}

// GetAPIKey returns the API key for a provider
func GetAPIKey(provider string) (string, bool) {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	key, ok := globalConfig.apiKeys[provider]
	return key, ok
}

// SetUserAgent sets a custom user agent
func SetUserAgent(ua string) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.userAgent = ua
}

// GetUserAgent returns the user agent
func GetUserAgent() string {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.userAgent
}

