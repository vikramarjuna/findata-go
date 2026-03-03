// Package nse provides NSE India data provider
package nse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Vikramarjuna/findata-go/config"
	"github.com/Vikramarjuna/findata-go/internal/provider"
	"github.com/Vikramarjuna/findata-go/logger"
)

const (
	// BaseURL is the NSE India API base URL
	BaseURL = "https://www.nseindia.com/api"

	// QuoteEndpoint is the endpoint for equity quotes
	QuoteEndpoint = "/quote-equity"
)

// Provider implements the NSE data provider
type Provider struct {
	baseURL string // Allow overriding for testing
}

// New creates a new NSE provider
func New() *Provider {
	return &Provider{
		baseURL: BaseURL,
	}
}

// NewWithBaseURL creates a new NSE provider with custom base URL (for testing)
func NewWithBaseURL(baseURL string) *Provider {
	return &Provider{
		baseURL: baseURL,
	}
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

// flexibleStringArray handles both string and []string from JSON
type flexibleStringArray []string

// UnmarshalJSON implements custom unmarshaling to handle both string and []string
func (f *flexibleStringArray) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		// Filter out "NA" values
		filtered := make([]string, 0, len(arr))
		for _, val := range arr {
			if val != "NA" && val != "" {
				filtered = append(filtered, val)
			}
		}
		*f = filtered
		return nil
	}

	// If that fails, try as a single string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Ignore "NA" and empty strings
		if str != "" && str != "NA" {
			*f = []string{str}
		} else {
			*f = []string{}
		}
		return nil
	}

	// If both fail, return empty array
	*f = []string{}
	return nil
}

// nseQuoteResponse represents the raw NSE API response
type nseQuoteResponse struct {
	Info struct {
		Symbol      string `json:"symbol"`
		CompanyName string `json:"companyName"`
		Industry    string `json:"industry"`
	} `json:"info"`
	Metadata struct {
		PdSectorIndAll flexibleStringArray `json:"pdSectorIndAll"`
	} `json:"metadata"`
	PriceInfo struct {
		LastPrice       float64 `json:"lastPrice"`
		Change          float64 `json:"change"`
		PChange         float64 `json:"pChange"`
		PreviousClose   float64 `json:"previousClose"`
		Open            float64 `json:"open"`
		Close           any     `json:"close"`
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
	logger.Debug("fetching NSE quote", "symbol", symbol, "provider", p.Name())

	// Fetch data from NSE API
	result, err := p.fetchNSEData(symbol)
	if err != nil {
		return nil, err
	}

	// Convert to provider.Quote
	quote := p.mapToQuote(result)
	logger.Info("successfully fetched NSE quote", "symbol", symbol, "price", quote.LastPrice, "currency", quote.Currency)
	return quote, nil
}

func (p *Provider) fetchNSEData(symbol string) (*nseQuoteResponse, error) {
	apiURL, err := url.Parse(p.baseURL + QuoteEndpoint)
	if err != nil {
		return nil, &provider.Error{
			Message:  "failed to parse base URL: " + err.Error(),
			Provider: p.Name(),
		}
	}

	query := apiURL.Query()
	query.Set("symbol", symbol)
	apiURL.RawQuery = query.Encode()

	logger.Debug("creating NSE API request", "url", apiURL.String(), "symbol", symbol)

	req, err := http.NewRequestWithContext(context.Background(), "GET", apiURL.String(), http.NoBody)
	if err != nil {
		logger.Error("failed to create NSE request", "error", err, "symbol", symbol)
		return nil, &provider.Error{
			Message:  "failed to create request: " + err.Error(),
			Provider: p.Name(),
		}
	}

	// Set headers to mimic browser request
	req.Header.Set("User-Agent", config.GetUserAgent())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	logger.Debug("sending HTTP request to NSE", "symbol", symbol)
	resp, err := config.GetHTTPClient().Do(req)
	if err != nil {
		logger.Error("NSE HTTP request failed", "error", err, "symbol", symbol, "url", apiURL.String())
		return nil, &provider.Error{
			Message:  "HTTP request failed: " + err.Error(),
			Provider: p.Name(),
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Warn("NSE API returned non-OK status", "status_code", resp.StatusCode, "symbol", symbol, "response", string(body))
		return nil, &provider.Error{
			Message:  "NSE API returned error: " + string(body),
			Code:     resp.StatusCode,
			Provider: p.Name(),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read NSE response body", "error", err, "symbol", symbol)
		return nil, &provider.Error{
			Message:  "failed to read response: " + err.Error(),
			Provider: p.Name(),
		}
	}

	logger.Debug("parsing NSE JSON response", "symbol", symbol, "body_size", len(body))
	var result nseQuoteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("failed to parse NSE JSON", "error", err, "symbol", symbol)
		return nil, &provider.Error{
			Message:  "failed to parse JSON: " + err.Error(),
			Provider: p.Name(),
		}
	}

	return &result, nil
}

func (p *Provider) mapToQuote(result *nseQuoteResponse) *provider.Quote {
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
		Indices:       []string(result.Metadata.PdSectorIndAll),
	}

	// Populate metadata with market cap classification
	if quote.Metadata == nil {
		quote.Metadata = make(map[string]string)
	}
	quote.Metadata["market_cap"] = determineMarketCap(quote.Indices)

	return quote
}

// determineMarketCap determines the market cap category based on NSE indices membership
func determineMarketCap(indices []string) string {
	hasLargeCap := false
	hasMidCap := false
	hasSmallCap := false

	for _, idx := range indices {
		switch idx {
		case "NIFTY 50", "NIFTY 100", "NIFTY100 EQUAL WEIGHT", "NIFTY100 ESG",
			"NIFTY100 ENHANCED ESG", "NIFTY100 LIQUID 15", "NIFTY100 LOW VOLATILITY 30",
			"NIFTY50 EQUAL WEIGHT", "NIFTY TOP 10 EQUAL WEIGHT", "NIFTY TOP 15 EQUAL WEIGHT",
			"NIFTY TOP 20 EQUAL WEIGHT":
			hasLargeCap = true
		case "NIFTY MIDCAP 50", "NIFTY MIDCAP 100", "NIFTY MIDCAP 150",
			"NIFTY MIDCAP SELECT", "NIFTY MIDCAP150 QUALITY 50":
			hasMidCap = true
		case "NIFTY SMALLCAP 50", "NIFTY SMALLCAP 100", "NIFTY SMALLCAP 250",
			"NIFTY SMLCAP 50", "NIFTY SMLCAP 100", "NIFTY SMLCAP 250":
			hasSmallCap = true
		}
	}

	// Priority: Large Cap > Mid Cap > Small Cap
	if hasLargeCap {
		return "Large Cap"
	}
	if hasMidCap {
		return "Mid Cap"
	}
	if hasSmallCap {
		return "Small Cap"
	}
	return "Other"
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
