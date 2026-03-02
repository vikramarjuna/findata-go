// Package metals provides access to precious metal prices (gold, silver, platinum, etc.)
package metals

import (
	"fmt"
	"strings"
	"time"
)

// MetalType represents the type of metal
type MetalType string

// Supported metal types
const (
	// Gold represents gold metal
	Gold MetalType = "GOLD"
	// Silver represents silver metal
	Silver MetalType = "SILVER"
	// Platinum represents platinum metal
	Platinum MetalType = "PLATINUM"
)

// Price represents a metal price quote
type Price struct {
	Metal        MetalType         `json:"metal"`
	Purity       string            `json:"purity"`         // 24K, 22K, 18K, 925, 999, etc.
	PricePerGram float64           `json:"price_per_gram"` // Price per gram in the specified currency
	Currency     string            `json:"currency"`       // INR, USD, etc.
	Source       string            `json:"source"`         // Data source name
	UpdatedAt    time.Time         `json:"updated_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Provider is the interface that metal price providers must implement
// This is an internal interface for different data sources (e.g., IndianProvider)
type Provider interface {
	// GetPrice fetches the price for a specific metal and purity
	GetPrice(metal MetalType, purity string) (*Price, error)

	// GetAllPrices fetches prices for all supported metals and purities
	GetAllPrices() ([]*Price, error)

	// SupportsMetalType checks if this provider supports the given metal type
	SupportsMetalType(metal MetalType) bool

	// Name returns the provider name
	Name() string
}

// Fetcher is the interface for fetching metal prices.
// This interface allows for easy mocking in tests.
type Fetcher interface {
	// Get fetches the price for a specific metal and purity
	Get(metal MetalType, purity string, opts ...Option) (*Price, error)
	// GetAll fetches prices for all supported metals and purities
	GetAll(opts ...Option) ([]*Price, error)
}

// DefaultFetcher is the default implementation of Fetcher
// that uses the package-level Get and GetAll functions
type DefaultFetcher struct{}

// NewFetcher creates a new DefaultFetcher
func NewFetcher() Fetcher {
	return &DefaultFetcher{}
}

// Get implements Fetcher.Get
func (f *DefaultFetcher) Get(metal MetalType, purity string, opts ...Option) (*Price, error) {
	return Get(metal, purity, opts...)
}

// GetAll implements Fetcher.GetAll
func (f *DefaultFetcher) GetAll(opts ...Option) ([]*Price, error) {
	return GetAll(opts...)
}

// Options for fetching metal prices
type Options struct {
	Currency string // Currency for prices (default: INR)
}

// Option is a functional option for configuring metal price requests
type Option func(*Options)

// WithCurrency sets the currency for the price request
func WithCurrency(currency string) Option {
	return func(o *Options) {
		o.Currency = currency
	}
}

// Get fetches the price for a specific metal and purity
// Example: Get(Gold, "24K") or Get(Silver, "999")
func Get(metal MetalType, purity string, opts ...Option) (*Price, error) {
	// Apply options
	options := &Options{
		Currency: "INR", // Default to INR
	}
	for _, opt := range opts {
		opt(options)
	}

	// Normalize inputs
	metal = MetalType(strings.ToUpper(string(metal)))
	purity = strings.ToUpper(purity)

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s_%s", metal, purity, options.Currency)
	c := getCache()
	if cached, ok := c.Get(cacheKey); ok {
		if price, ok := cached.(*Price); ok {
			return price, nil
		}
	}

	// Get provider
	provider := getProvider(options.Currency)

	// Fetch price
	price, err := provider.GetPrice(metal, purity)
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.Set(cacheKey, price)

	return price, nil
}

// GetAll fetches prices for all supported metals and purities
func GetAll(opts ...Option) ([]*Price, error) {
	// Apply options
	options := &Options{
		Currency: "INR",
	}
	for _, opt := range opts {
		opt(options)
	}

	// Check cache first
	cacheKey := fmt.Sprintf("all_%s", options.Currency)
	c := getCache()
	if cached, ok := c.Get(cacheKey); ok {
		if prices, ok := cached.([]*Price); ok {
			return prices, nil
		}
	}

	// Get provider
	provider := getProvider(options.Currency)

	// Fetch all prices
	prices, err := provider.GetAllPrices()
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.Set(cacheKey, prices)

	return prices, nil
}

// getProvider returns the appropriate provider based on currency
func getProvider(currency string) Provider {
	// For now, we only support INR with the Indian provider
	// In the future, we can add more providers for different currencies
	if currency == "INR" {
		return newIndianProvider()
	}

	// Default to Indian provider
	return newIndianProvider()
}

// Error represents a metal price error
type Error struct {
	Message  string
	Metal    MetalType
	Purity   string
	Provider string
}

func (e *Error) Error() string {
	if e.Metal != "" && e.Purity != "" {
		return fmt.Sprintf("metals: %s %s: %s (provider: %s)", e.Metal, e.Purity, e.Message, e.Provider)
	}
	return fmt.Sprintf("metals: %s (provider: %s)", e.Message, e.Provider)
}
