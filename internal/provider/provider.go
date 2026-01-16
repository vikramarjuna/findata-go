// Package provider defines the interface for data providers
package provider

import (
	"fmt"

	"github.com/Vikramarjuna/findata-go/config"
)

// Quote represents a stock quote from any provider
type Quote struct {
	Symbol        string            `json:"symbol"`
	Exchange      config.Exchange   `json:"exchange"`
	CompanyName   string            `json:"company_name"`
	Industry      string            `json:"industry"`
	Sector        string            `json:"sector"`
	LastPrice     float64           `json:"last_price"`
	Currency      string            `json:"currency"`
	Change        float64           `json:"change"`
	PChange       float64           `json:"pchange"`
	PreviousClose float64           `json:"previous_close"`
	Open          float64           `json:"open"`
	DayHigh       float64           `json:"day_high"`
	DayLow        float64           `json:"day_low"`
	YearHigh      float64           `json:"year_high"`
	YearLow       float64           `json:"year_low"`
	Volume        float64           `json:"volume"`
	Value         float64           `json:"value"`
	Indices       []string          `json:"indices,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Provider is the interface that all data providers must implement
type Provider interface {
	// Get fetches a quote for a single symbol
	Get(symbol string) (*Quote, error)

	// GetMultiple fetches quotes for multiple symbols
	GetMultiple(symbols []string) (map[string]*Quote, []error)

	// SupportsSymbol checks if this provider can handle the given symbol
	SupportsSymbol(symbol string) bool

	// Name returns the provider name
	Name() string
}

// Error represents a provider error
type Error struct {
	Message  string
	Code     int
	Provider string
}

func (e *Error) Error() string {
	if e.Code > 0 {
		return fmt.Sprintf("%s: %s (code: %d)", e.Provider, e.Message, e.Code)
	}
	return e.Provider + ": " + e.Message
}
