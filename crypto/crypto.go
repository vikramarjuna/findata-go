// Package crypto provides access to cryptocurrency prices
package crypto

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Vikramarjuna/findata-go/internal/cache"
	"github.com/Vikramarjuna/findata-go/internal/provider/coingecko"
)

// Quote represents a cryptocurrency quote with price and market data
type Quote struct {
	CoinID            string            `json:"coin_id"`
	Name              string            `json:"name"`
	CurrentPrice      float64           `json:"current_price"`
	Currency          string            `json:"currency"` // USD, INR, etc.
	MarketCap         float64           `json:"market_cap"`
	Volume24h         float64           `json:"volume_24h"`
	PriceChangePct24h float64           `json:"price_change_pct_24h"`
	Source            string            `json:"source"` // Data source name
	UpdatedAt         time.Time         `json:"updated_at"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

var (
	globalCache     *cache.Cache
	globalCacheLock sync.Mutex
	globalProvider  *coingecko.Provider
	providerLock    sync.Mutex
)

// Default cache TTL for crypto prices (1 minute - crypto prices change frequently)
const defaultCacheTTL = 1 * time.Minute

// Fetcher is the interface for fetching crypto prices.
// This interface allows for easy mocking in tests.
type Fetcher interface {
	// Get fetches the price for a specific cryptocurrency
	Get(coinId string, opts ...Option) (*Quote, error)
	// GetMultiple fetches prices for multiple cryptocurrencies
	GetMultiple(coinIds []string, opts ...Option) (map[string]*Quote, error)
}

// DefaultFetcher is the default implementation of Fetcher
type DefaultFetcher struct{}

// NewFetcher creates a new DefaultFetcher
func NewFetcher() Fetcher {
	return &DefaultFetcher{}
}

// Get implements Fetcher.Get
func (f *DefaultFetcher) Get(coinId string, opts ...Option) (*Quote, error) {
	return Get(coinId, opts...)
}

// GetMultiple implements Fetcher.GetMultiple
func (f *DefaultFetcher) GetMultiple(coinIds []string, opts ...Option) (map[string]*Quote, error) {
	return GetMultiple(coinIds, opts...)
}

// Options for fetching crypto prices
type Options struct {
	Currency string // Currency for prices (default: USD)
}

// Option is a functional option for configuring crypto price requests
type Option func(*Options)

// WithCurrency sets the currency for the price request
func WithCurrency(currency string) Option {
	return func(o *Options) {
		o.Currency = currency
	}
}

// Get fetches the quote for a specific cryptocurrency
// Example: Get("BTC") or Get("ETH", WithCurrency("INR"))
func Get(coinId string, opts ...Option) (*Quote, error) {
	// Apply options
	options := &Options{
		Currency: "USD", // Default to USD
	}
	for _, opt := range opts {
		opt(options)
	}

	// Normalize inputs
	coinId = strings.ToUpper(coinId)
	options.Currency = strings.ToUpper(options.Currency)

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s", coinId, options.Currency)
	c := getCache()
	if cached, ok := c.Get(cacheKey); ok {
		if quote, ok := cached.(*Quote); ok {
			return quote, nil
		}
	}

	// Get provider
	provider := getProvider()

	// Fetch quote data from provider
	data, err := provider.GetPrice(coinId, options.Currency)
	if err != nil {
		return nil, &Error{
			Message:  err.Error(),
			CoinID:   coinId,
			Currency: options.Currency,
			Provider: provider.Name(),
		}
	}

	// Convert to Quote struct
	quote := dataToQuote(coinId, options.Currency, data, provider.Name())

	// Cache the result
	c.Set(cacheKey, quote)

	return quote, nil
}

// GetMultiple fetches quotes for multiple cryptocurrencies
// Example: GetMultiple([]string{"BTC", "ETH"}, WithCurrency("INR"))
func GetMultiple(coinIds []string, opts ...Option) (map[string]*Quote, error) {
	// Apply options
	options := &Options{
		Currency: "USD",
	}
	for _, opt := range opts {
		opt(options)
	}

	// Normalize currency
	options.Currency = strings.ToUpper(options.Currency)

	// Normalize symbols
	normalizedSymbols := make([]string, len(coinIds))
	for i, symbol := range coinIds {
		normalizedSymbols[i] = strings.ToUpper(symbol)
	}

	// Get provider
	provider := getProvider()

	// Fetch quotes data from provider
	dataMap, err := provider.GetMultiplePrices(normalizedSymbols, options.Currency)
	if err != nil {
		return nil, &Error{
			Message:  err.Error(),
			Currency: options.Currency,
			Provider: provider.Name(),
		}
	}

	// Convert to Quote structs
	quotes := make(map[string]*Quote)
	for symbol, data := range dataMap {
		quotes[symbol] = dataToQuote(symbol, options.Currency, data, provider.Name())
	}

	// Cache the results
	c := getCache()
	for symbol, quote := range quotes {
		cacheKey := fmt.Sprintf("%s_%s", symbol, options.Currency)
		c.Set(cacheKey, quote)
	}

	return quotes, nil
}

// getCache returns the global cache instance
func getCache() *cache.Cache {
	globalCacheLock.Lock()
	defer globalCacheLock.Unlock()

	if globalCache == nil {
		globalCache = cache.New(cache.Config{
			TTL:             defaultCacheTTL,
			MaxSize:         1000,
			Enabled:         true,
			CleanupInterval: 5 * time.Minute,
		})
	}
	return globalCache
}

// getProvider returns the crypto price provider
func getProvider() *coingecko.Provider {
	providerLock.Lock()
	defer providerLock.Unlock()

	if globalProvider == nil {
		globalProvider = coingecko.New()
	}
	return globalProvider
}

// dataToQuote converts provider data to Quote struct
func dataToQuote(coinID, currency string, data map[string]interface{}, providerName string) *Quote {
	currencyLower := strings.ToLower(currency)

	return &Quote{
		CoinID:            coinID,
		Name:              getString(data, "coin_name"),
		CurrentPrice:      coingecko.GetFloat(data, currencyLower),
		Currency:          currency,
		MarketCap:         coingecko.GetFloat(data, currencyLower+"_market_cap"),
		Volume24h:         coingecko.GetFloat(data, currencyLower+"_24h_vol"),
		PriceChangePct24h: coingecko.GetFloat(data, currencyLower+"_24h_change"),
		Source:            providerName,
		UpdatedAt:         time.Now(),
	}
}

// getString safely extracts a string from the response data
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// ClearCache clears the global cache
func ClearCache() {
	getCache().Clear()
}

// Error represents a crypto quote error
type Error struct {
	Message  string
	CoinID   string
	Currency string
	Provider string
}

func (e *Error) Error() string {
	if e.CoinID != "" {
		return fmt.Sprintf("crypto: %s (%s): %s (provider: %s)", e.CoinID, e.Currency, e.Message, e.Provider)
	}
	return fmt.Sprintf("crypto: %s (provider: %s)", e.Message, e.Provider)
}
