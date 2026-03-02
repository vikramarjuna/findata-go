// Package coingecko provides CoinGecko API data provider for cryptocurrency prices
package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Vikramarjuna/findata-go/config"
)

const (
	// BaseURL is the CoinGecko API base URL
	BaseURL = "https://api.coingecko.com/api/v3"
)

// Provider implements the CoinGecko data provider
type Provider struct {
	client  *http.Client
	baseURL string
}

// New creates a new CoinGecko provider
func New() *Provider {
	return &Provider{
		client:  config.GetHTTPClient(),
		baseURL: BaseURL,
	}
}

// NewWithBaseURL creates a new CoinGecko provider with custom base URL (for testing)
func NewWithBaseURL(baseURL string) *Provider {
	return &Provider{
		client:  config.GetHTTPClient(),
		baseURL: baseURL,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "CoinGecko"
}

// GetPrice fetches the price for a specific cryptocurrency
// coinID should be the CoinGecko coin ID (e.g., "bitcoin", "ethereum", "solana")
// currency is the fiat currency (e.g., "USD", "INR")
func (p *Provider) GetPrice(coinID, currency string) (map[string]interface{}, error) {
	// Normalize coin ID to lowercase
	coinID = strings.ToLower(coinID)

	// Build URL
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=%s&include_market_cap=true&include_24hr_vol=true&include_24hr_change=true&include_last_updated_at=true",
		p.baseURL, coinID, strings.ToLower(currency))

	// Make request
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract data
	data, ok := result[coinID]
	if !ok {
		return nil, fmt.Errorf("cryptocurrency not found in response")
	}

	// Add metadata
	data["coin_id"] = coinID

	return data, nil
}

// GetMultiplePrices fetches prices for multiple cryptocurrencies
// coinIDs should be CoinGecko coin IDs (e.g., ["bitcoin", "ethereum", "solana"])
func (p *Provider) GetMultiplePrices(coinIDs []string, currency string) (map[string]map[string]interface{}, error) {
	if len(coinIDs) == 0 {
		return make(map[string]map[string]interface{}), nil
	}

	// Normalize coin IDs to lowercase
	normalizedIDs := make([]string, len(coinIDs))
	for i, id := range coinIDs {
		normalizedIDs[i] = strings.ToLower(id)
	}

	// Build URL
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=%s&include_market_cap=true&include_24hr_vol=true&include_24hr_change=true&include_last_updated_at=true",
		p.baseURL, strings.Join(normalizedIDs, ","), strings.ToLower(currency))

	// Make request
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Add metadata to each result
	for coinID, data := range result {
		data["coin_id"] = coinID
	}

	return result, nil
}

// GetFloat safely extracts a float64 from the response data
func GetFloat(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}
