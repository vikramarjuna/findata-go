// Package nse provides access to NSE India stock quotes
package nse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Vikramarjuna/finance-india"
)

const (
	// BaseURL is the NSE India API base URL
	BaseURL = "https://www.nseindia.com/api"

	// QuoteEndpoint is the endpoint for equity quotes
	QuoteEndpoint = "/quote-equity"
)

// Quote represents a stock quote from NSE India
type Quote struct {
	Symbol            string   `json:"symbol"`
	CompanyName       string   `json:"company_name"`
	Industry          string   `json:"industry"`
	Sector            string   `json:"sector"`
	LastPrice         float64  `json:"last_price"`
	Change            float64  `json:"change"`
	PChange           float64  `json:"pchange"`
	PreviousClose     float64  `json:"previous_close"`
	Open              float64  `json:"open"`
	DayHigh           float64  `json:"day_high"`
	DayLow            float64  `json:"day_low"`
	YearHigh          float64  `json:"year_high"`
	YearLow           float64  `json:"year_low"`
	TotalTradedVolume float64  `json:"total_traded_volume"`
	TotalTradedValue  float64  `json:"total_traded_value"`
	Indices           []string `json:"indices"`
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
func Get(symbol string) (*Quote, error) {
	url := fmt.Sprintf("%s%s?symbol=%s", BaseURL, QuoteEndpoint, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &finance.Error{Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	// Set headers to mimic browser request
	req.Header.Set("User-Agent", finance.DefaultUserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := finance.Client.Do(req)
	if err != nil {
		return nil, &finance.Error{Message: fmt.Sprintf("HTTP request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &finance.Error{
			Message: fmt.Sprintf("NSE API returned error: %s", string(body)),
			Code:    resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &finance.Error{Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	var result nseQuoteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &finance.Error{Message: fmt.Sprintf("failed to parse JSON: %v", err)}
	}

	// Map to our Quote struct
	quote := &Quote{
		Symbol:            result.Info.Symbol,
		CompanyName:       result.Info.CompanyName,
		Industry:          result.IndustryInfo.Industry,
		Sector:            result.IndustryInfo.Sector,
		LastPrice:         result.PriceInfo.LastPrice,
		Change:            result.PriceInfo.Change,
		PChange:           result.PriceInfo.PChange,
		PreviousClose:     result.PriceInfo.PreviousClose,
		Open:              result.PriceInfo.Open,
		DayHigh:           result.PriceInfo.IntraDayHighLow.Max,
		DayLow:            result.PriceInfo.IntraDayHighLow.Min,
		YearHigh:          result.PriceInfo.WeekHighLow.Max,
		YearLow:           result.PriceInfo.WeekHighLow.Min,
		TotalTradedVolume: result.PreOpenMarket.TotalTradedVolume,
		TotalTradedValue:  result.PreOpenMarket.TotalTradedValue,
		Indices:           result.Metadata.PdSectorIndAll,
	}

	return quote, nil
}

// GetMultiple fetches quotes for multiple symbols
// Returns a map of symbol -> quote, and any errors encountered
func GetMultiple(symbols []string) (map[string]*Quote, []error) {
	quotes := make(map[string]*Quote)
	var errors []error

	for _, symbol := range symbols {
		quote, err := Get(symbol)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", symbol, err))
			continue
		}
		quotes[symbol] = quote
	}

	return quotes, errors
}
