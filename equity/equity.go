// Package equity provides a unified interface for fetching equity quotes
package equity

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Vikramarjuna/findata-go/cache"
	"github.com/Vikramarjuna/findata-go/config"
	"github.com/Vikramarjuna/findata-go/logger"
	"github.com/Vikramarjuna/findata-go/provider"
	"github.com/Vikramarjuna/findata-go/provider/nse"
)

// Quote is an alias for provider.Quote for convenience
type Quote = provider.Quote

// Options for customizing quote requests
type Options struct {
	Market   config.Market
	Exchange config.Exchange
}

// Option is a functional option for customizing requests
type Option func(*Options)

// WithMarket sets the market for the request
func WithMarket(market config.Market) Option {
	return func(o *Options) {
		o.Market = market
	}
}

// WithExchange sets the exchange for the request
func WithExchange(exchange config.Exchange) Option {
	return func(o *Options) {
		o.Exchange = exchange
	}
}

// Registry of available providers
var providers = []provider.Provider{
	nse.New(),
	// Future: Add more providers here
	// bse.New(),
	// alphavantage.New(),
	// finnhub.New(),
}

// Global cache instance
var (
	globalCache      *cache.Cache
	cacheMu          sync.RWMutex
	cacheInitialized bool
)

// initCache initializes the global cache if not already initialized
func initCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cacheInitialized {
		return
	}

	cacheConfig := config.GetCacheConfig()
	globalCache = cache.New(cache.Config{
		TTL:             cacheConfig.TTL,
		MaxSize:         cacheConfig.MaxSize,
		Enabled:         cacheConfig.Enabled,
		CleanupInterval: cacheConfig.CleanupInterval,
	})
	cacheInitialized = true
}

// getCache returns the global cache, initializing if necessary
func getCache() *cache.Cache {
	if !cacheInitialized {
		initCache()
	}
	return globalCache
}

// ClearCache clears all cached quotes
func ClearCache() {
	c := getCache()
	if c != nil {
		c.Clear()
	}
}

// GetCacheStats returns cache statistics
func GetCacheStats() cache.Stats {
	c := getCache()
	if c != nil {
		return c.GetStats()
	}
	return cache.Stats{}
}

// detectProvider finds the best provider for a symbol
func detectProvider(symbol string, opts *Options) (provider.Provider, error) {
	logger.Debug("detecting provider for symbol", "symbol", symbol, "exchange", opts.Exchange, "market", opts.Market)

	// If exchange is explicitly specified, use that
	if opts.Exchange != "" {
		logger.Debug("using explicitly specified exchange", "exchange", opts.Exchange)
		switch opts.Exchange {
		case config.ExchangeNSE:
			return nse.New(), nil
		case config.ExchangeBSE:
			return nil, fmt.Errorf("BSE provider not yet implemented")
		case config.ExchangeNYSE, config.ExchangeNASDAQ:
			return nil, fmt.Errorf("US market providers not yet implemented")
		}
	}

	// If market is specified, filter providers by market
	if opts.Market != "" && opts.Market != config.MarketAuto {
		logger.Debug("filtering by market", "market", opts.Market)
		switch opts.Market {
		case config.MarketIndia:
			// Try NSE first for Indian market
			if nse.New().SupportsSymbol(symbol) {
				logger.Debug("NSE provider selected for Indian market", "symbol", symbol)
				return nse.New(), nil
			}
		case config.MarketUS:
			return nil, fmt.Errorf("US market providers not yet implemented")
		}
	}

	// Auto-detect based on symbol pattern
	logger.Debug("auto-detecting provider", "symbol", symbol)
	for _, p := range providers {
		if p.SupportsSymbol(symbol) {
			logger.Debug("provider auto-detected", "symbol", symbol, "provider", p.Name())
			return p, nil
		}
	}

	logger.Warn("no provider found for symbol", "symbol", symbol)
	return nil, fmt.Errorf("no provider found for symbol: %s", symbol)
}

// Get fetches a quote for a single symbol
// The symbol is automatically routed to the appropriate provider
// Use options to override default behavior
// Quotes are cached according to the global cache configuration
func Get(symbol string, opts ...Option) (*Quote, error) {
	// Apply options
	options := &Options{
		Market: config.GetDefaultMarket(),
	}
	for _, opt := range opts {
		opt(options)
	}

	// Clean up symbol
	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	logger.Debug("fetching equity quote", "symbol", symbol)

	// Check cache first
	c := getCache()
	if cached, ok := c.Get(symbol); ok {
		if quote, ok := cached.(*Quote); ok {
			logger.Debug("returning cached quote", "symbol", symbol)
			return quote, nil
		}
	}

	logger.Debug("cache miss, fetching from provider", "symbol", symbol)

	// Find the right provider
	p, err := detectProvider(symbol, options)
	if err != nil {
		logger.Error("failed to detect provider", "symbol", symbol, "error", err)
		return nil, err
	}

	// Fetch the quote
	quote, err := p.Get(symbol)
	if err != nil {
		logger.Error("failed to fetch quote from provider", "symbol", symbol, "provider", p.Name(), "error", err)
		return nil, err
	}

	logger.Info("successfully fetched equity quote", "symbol", symbol, "exchange", quote.Exchange, "price", quote.LastPrice)

	// Cache the result
	c.Set(symbol, quote)

	return quote, nil
}

// GetMultiple fetches quotes for multiple symbols
// Returns a map of symbol -> quote and any errors encountered
// Uses cache to avoid duplicate API calls
func GetMultiple(symbols []string, opts ...Option) (map[string]*Quote, []error) {
	logger.Debug("fetching multiple equity quotes", "count", len(symbols))

	// Apply options
	options := &Options{
		Market: config.GetDefaultMarket(),
	}
	for _, opt := range opts {
		opt(options)
	}

	quotes := make(map[string]*Quote)
	var errors []error
	c := getCache()

	// Group symbols by provider, checking cache first
	providerSymbols := make(map[provider.Provider][]string)
	cachedCount := 0

	for _, symbol := range symbols {
		symbol = strings.TrimSpace(strings.ToUpper(symbol))

		// Check cache first
		if cached, ok := c.Get(symbol); ok {
			if quote, ok := cached.(*Quote); ok {
				quotes[symbol] = quote
				cachedCount++
				continue
			}
		}

		// Not in cache, need to fetch
		p, err := detectProvider(symbol, options)
		if err != nil {
			logger.Warn("failed to detect provider for symbol in batch", "symbol", symbol, "error", err)
			errors = append(errors, fmt.Errorf("%s: %w", symbol, err))
			continue
		}
		providerSymbols[p] = append(providerSymbols[p], symbol)
	}

	logger.Debug("batch fetch cache stats", "total", len(symbols), "cached", cachedCount, "to_fetch", len(symbols)-cachedCount)

	// Fetch uncached symbols from each provider
	for p, syms := range providerSymbols {
		logger.Debug("fetching from provider", "provider", p.Name(), "symbols_count", len(syms))
		providerQuotes, providerErrors := p.GetMultiple(syms)
		for symbol, quote := range providerQuotes {
			quotes[symbol] = quote
			// Cache the result
			c.Set(symbol, quote)
		}
		if len(providerErrors) > 0 {
			logger.Warn("provider returned errors", "provider", p.Name(), "error_count", len(providerErrors))
		}
		errors = append(errors, providerErrors...)
	}

	logger.Info("batch fetch completed", "total_requested", len(symbols), "successful", len(quotes), "errors", len(errors))
	return quotes, errors
}
