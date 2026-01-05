// Package nse provides NSE India data provider
package nse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/Vikramarjuna/findata-go/config"
	"github.com/Vikramarjuna/findata-go/provider"
)

const (
	// BaseURL is the NSE India API base URL
	BaseURL = "https://www.nseindia.com/api"

	// QuoteEndpoint is the endpoint for equity quotes
	QuoteEndpoint = "/quote-equity"
)

// Provider implements the NSE data provider
type Provider struct{}

// New creates a new NSE provider
func New() *Provider {
	return &Provider{}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "NSE"
}

// SupportsSymbol checks if this provider can handle the given symbol
func (p *Provider) SupportsSymbol(symbol string) bool {
	// NSE symbols are typically all uppercase letters and numbers
	// No special characters like dots or dashes
	matched, _ := regexp.MatchString(`^[A-Z0-9&]+$`, symbol)
	return matched
}

// nseQuoteResponse represents the raw NSE API response
type nseQuoteResponse struct {
	Info struct {
		Symbol      string `json:"symbol"`
		CompanyName string `json:"companyName"`
		Industry    string `json:"industry"`
	} `json:"info"`
	Metadata struct {
		PdSectorIndAll []string `json:"pdSectorIndAll"`
	} `json:"metadata"`
	PriceInfo struct {
		LastPrice       float64     `json:"lastPrice"`
		Change          float64     `json:"change"`
		PChange         float64     `json:"pChange"`
		PreviousClose   float64     `json:"previousClose"`
		Open            float64     `json:"open"`
		Close           interface{} `json:"close"`
		IntraDayHighLow struct {
			Max float64 `json:"max"`
			Min float64 `json:"min"`
		} `json:"intraDayHighLow"`
		WeekHighLow struct {
			Max float64 `json:"max"`
			Min float64 `json:"min"`
		} `json:"weekHighLow"`
	} `json:"priceInfo"`
	PreOpenMarket struct {
		TotalTradedVolume float64 `json:"totalTradedVolume"`
		TotalTradedValue  float64 `json:"totalTradedValue"`
	} `json:"preOpenMarket"`
	IndustryInfo struct {
		Macro    string `json:"macro"`
		Sector   string `json:"sector"`
		Industry string `json:"industry"`
	} `json:"industryInfo"`
}

// Get fetches a quote for the given symbol from NSE India
func (p *Provider) Get(symbol string) (*provider.Quote, error) {
	// Clean up symbol (remove any whitespace, convert to uppercase)
	symbol = strings.TrimSpace(strings.ToUpper(symbol))

	url := fmt.Sprintf("%s%s?symbol=%s", BaseURL, QuoteEndpoint, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &provider.Error{
			Message:  fmt.Sprintf("failed to create request: %v", err),
			Provider: p.Name(),
		}
	}

	// Set headers to mimic browser request
	req.Header.Set("User-Agent", config.GetUserAgent())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := config.GetHTTPClient().Do(req)
	if err != nil {
		return nil, &provider.Error{
			Message:  fmt.Sprintf("HTTP request failed: %v", err),
			Provider: p.Name(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &provider.Error{
			Message:  fmt.Sprintf("NSE API returned error: %s", string(body)),
			Code:     resp.StatusCode,
			Provider: p.Name(),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &provider.Error{
			Message:  fmt.Sprintf("failed to read response: %v", err),
			Provider: p.Name(),
		}
	}

	var result nseQuoteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &provider.Error{
			Message:  fmt.Sprintf("failed to parse JSON: %v", err),
			Provider: p.Name(),
		}
	}

	// Map to provider.Quote
	quote := &provider.Quote{
		Symbol:        result.Info.Symbol,
		Exchange:      config.ExchangeNSE,
		CompanyName:   result.Info.CompanyName,
		Industry:      result.IndustryInfo.Industry,
		Sector:        result.IndustryInfo.Sector,
		LastPrice:     result.PriceInfo.LastPrice,
		Currency:      "INR",
		Change:        result.PriceInfo.Change,
		PChange:       result.PriceInfo.PChange,
		PreviousClose: result.PriceInfo.PreviousClose,
		Open:          result.PriceInfo.Open,
		DayHigh:       result.PriceInfo.IntraDayHighLow.Max,
		DayLow:        result.PriceInfo.IntraDayHighLow.Min,
		YearHigh:      result.PriceInfo.WeekHighLow.Max,
		YearLow:       result.PriceInfo.WeekHighLow.Min,
		Volume:        result.PreOpenMarket.TotalTradedVolume,
		Value:         result.PreOpenMarket.TotalTradedValue,
		Indices:       result.Metadata.PdSectorIndAll,
	}

	return quote, nil
}

// GetMultiple fetches quotes for multiple symbols
func (p *Provider) GetMultiple(symbols []string) (map[string]*provider.Quote, []error) {
	quotes := make(map[string]*provider.Quote)
	var errors []error

	for _, symbol := range symbols {
		quote, err := p.Get(symbol)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", symbol, err))
			continue
		}
		quotes[symbol] = quote
	}

	return quotes, errors
}

