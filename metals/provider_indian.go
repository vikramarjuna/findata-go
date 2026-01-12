package metals

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Vikramarjuna/findata-go/config"
)

// IndianProvider fetches metal prices for the Indian market (INR)
type IndianProvider struct {
	client *http.Client
}

// newIndianProvider creates a new Indian metal price provider
func newIndianProvider() *IndianProvider {
	return &IndianProvider{
		client: config.GetHTTPClient(),
	}
}

// Name returns the provider name
func (p *IndianProvider) Name() string {
	return "Indian Metals"
}

// SupportsMetalType checks if this provider supports the given metal type
func (p *IndianProvider) SupportsMetalType(metal MetalType) bool {
	switch metal {
	case Gold, Silver, Platinum:
		return true
	default:
		return false
	}
}

// GetPrice fetches the price for a specific metal and purity
func (p *IndianProvider) GetPrice(metal MetalType, purity string) (*Price, error) {
	if !p.SupportsMetalType(metal) {
		return nil, &Error{
			Message:  fmt.Sprintf("metal type %s not supported", metal),
			Metal:    metal,
			Purity:   purity,
			Provider: p.Name(),
		}
	}

	switch metal {
	case Gold:
		return p.getGoldPrice(purity)
	case Silver:
		return p.getSilverPrice(purity)
	case Platinum:
		return p.getPlatinumPrice(purity)
	default:
		return nil, &Error{
			Message:  fmt.Sprintf("metal type %s not supported", metal),
			Metal:    metal,
			Purity:   purity,
			Provider: p.Name(),
		}
	}
}

// GetAllPrices fetches prices for all supported metals and purities
func (p *IndianProvider) GetAllPrices() ([]*Price, error) {
	var prices []*Price

	// Gold prices (24K, 22K, 18K, 14K)
	goldPurities := []string{"24K", "22K", "18K", "14K"}
	for _, purity := range goldPurities {
		price, err := p.getGoldPrice(purity)
		if err == nil {
			prices = append(prices, price)
		}
	}

	// Silver prices (999, 925)
	silverPurities := []string{"999", "925"}
	for _, purity := range silverPurities {
		price, err := p.getSilverPrice(purity)
		if err == nil {
			prices = append(prices, price)
		}
	}

	// Platinum prices (999, 950)
	platinumPurities := []string{"999", "950"}
	for _, purity := range platinumPurities {
		price, err := p.getPlatinumPrice(purity)
		if err == nil {
			prices = append(prices, price)
		}
	}

	if len(prices) == 0 {
		return nil, &Error{
			Message:  "failed to fetch any metal prices",
			Provider: p.Name(),
		}
	}

	return prices, nil
}

// getGoldPrice fetches gold price for a specific purity
func (p *IndianProvider) getGoldPrice(purity string) (*Price, error) {
	// Fetch base 24K gold price
	base24KPrice, err := p.fetchGoldPriceFromAPI()
	if err != nil {
		// Use fallback price
		base24KPrice = 7200.0 // Approximate 24K gold price per gram in INR (Jan 2026)
	}

	// Calculate price based on purity
	var pricePerGram float64
	var purityMultiplier float64

	switch purity {
	case "24K":
		purityMultiplier = 1.000
	case "22K":
		purityMultiplier = 0.916
	case "18K":
		purityMultiplier = 0.750
	case "14K":
		purityMultiplier = 0.585
	default:
		return nil, &Error{
			Message:  fmt.Sprintf("unsupported gold purity: %s", purity),
			Metal:    Gold,
			Purity:   purity,
			Provider: p.Name(),
		}
	}

	pricePerGram = base24KPrice * purityMultiplier

	return &Price{
		Metal:        Gold,
		Purity:       purity,
		PricePerGram: pricePerGram,
		Currency:     "INR",
		Source:       p.Name(),
		UpdatedAt:    time.Now(),
	}, nil
}

// getSilverPrice fetches silver price for a specific purity
func (p *IndianProvider) getSilverPrice(purity string) (*Price, error) {
	// Fetch base 999 silver price
	base999Price, err := p.fetchSilverPriceFromAPI()
	if err != nil {
		// Use fallback price
		base999Price = 90.0 // Approximate 999 silver price per gram in INR (Jan 2026)
	}

	// Calculate price based on purity
	var pricePerGram float64
	var purityMultiplier float64

	switch purity {
	case "999":
		purityMultiplier = 1.000
	case "925": // Sterling silver
		purityMultiplier = 0.925
	default:
		return nil, &Error{
			Message:  fmt.Sprintf("unsupported silver purity: %s", purity),
			Metal:    Silver,
			Purity:   purity,
			Provider: p.Name(),
		}
	}

	pricePerGram = base999Price * purityMultiplier

	return &Price{
		Metal:        Silver,
		Purity:       purity,
		PricePerGram: pricePerGram,
		Currency:     "INR",
		Source:       p.Name(),
		UpdatedAt:    time.Now(),
	}, nil
}

// getPlatinumPrice fetches platinum price for a specific purity
func (p *IndianProvider) getPlatinumPrice(purity string) (*Price, error) {
	// Fetch base 999 platinum price
	base999Price, err := p.fetchPlatinumPriceFromAPI()
	if err != nil {
		// Use fallback price
		base999Price = 3200.0 // Approximate 999 platinum price per gram in INR (Jan 2026)
	}

	// Calculate price based on purity
	var pricePerGram float64
	var purityMultiplier float64

	switch purity {
	case "999":
		purityMultiplier = 1.000
	case "950": // Common platinum purity for jewelry
		purityMultiplier = 0.950
	default:
		return nil, &Error{
			Message:  fmt.Sprintf("unsupported platinum purity: %s", purity),
			Metal:    Platinum,
			Purity:   purity,
			Provider: p.Name(),
		}
	}

	pricePerGram = base999Price * purityMultiplier

	return &Price{
		Metal:        Platinum,
		Purity:       purity,
		PricePerGram: pricePerGram,
		Currency:     "INR",
		Source:       p.Name(),
		UpdatedAt:    time.Now(),
	}, nil
}

// fetchGoldPriceFromAPI attempts to fetch gold price from external API
func (p *IndianProvider) fetchGoldPriceFromAPI() (float64, error) {
	// Try GoodReturns API
	url := "https://www.goodreturns.in/gold-rates/api/get-gold-price.json"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	// Extract 24K gold price (per 10 grams)
	var pricePer10g float64
	if data, ok := result["data"].(map[string]interface{}); ok {
		if gold24k, ok := data["24k"].(float64); ok {
			pricePer10g = gold24k
		} else if gold24k, ok := data["gold_24k"].(float64); ok {
			pricePer10g = gold24k
		}
	}

	if pricePer10g == 0 {
		return 0, fmt.Errorf("could not parse gold price from API response")
	}

	// Convert from per 10 grams to per gram
	return pricePer10g / 10.0, nil
}

// fetchSilverPriceFromAPI attempts to fetch silver price from external API
func (p *IndianProvider) fetchSilverPriceFromAPI() (float64, error) {
	// TODO: Implement silver price API
	// For now, return error to use fallback
	return 0, fmt.Errorf("silver price API not implemented")
}

// fetchPlatinumPriceFromAPI attempts to fetch platinum price from external API
func (p *IndianProvider) fetchPlatinumPriceFromAPI() (float64, error) {
	// TODO: Implement platinum price API
	// For now, return error to use fallback
	return 0, fmt.Errorf("platinum price API not implemented")
}
