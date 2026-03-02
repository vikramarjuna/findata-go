package metals

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Vikramarjuna/findata-go/config"
)

// IndianProvider fetches metal prices for the Indian market (INR)
type IndianProvider struct {
	client    *http.Client
	ibjaCache *ibjaData
	cacheTime time.Time
	cacheTTL  time.Duration
}

// ibjaData holds parsed IBJA metal prices
type ibjaData struct {
	gold999     float64 // per gram (24K)
	gold995     float64 // per gram
	gold916     float64 // per gram (22K)
	gold750     float64 // per gram (18K)
	gold585     float64 // per gram (14K)
	silver999   float64 // per gram
	platinum999 float64 // per gram
}

// newIndianProvider creates a new Indian metal price provider
func newIndianProvider() *IndianProvider {
	return &IndianProvider{
		client:   config.GetHTTPClient(),
		cacheTTL: 5 * time.Minute, // Cache IBJA data for 5 minutes
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
	// Fetch IBJA data (cached)
	data, err := p.fetchIBJAData()

	pricePerGram := p.getGoldPriceByPurity(purity, data, err)
	if pricePerGram == 0 {
		return nil, &Error{
			Message:  fmt.Sprintf("unsupported gold purity: %s", purity),
			Metal:    Gold,
			Purity:   purity,
			Provider: p.Name(),
		}
	}

	return &Price{
		Metal:        Gold,
		Purity:       purity,
		PricePerGram: pricePerGram,
		Currency:     "INR",
		Source:       p.Name(),
		UpdatedAt:    time.Now(),
	}, nil
}

// getGoldPriceByPurity returns the gold price for a specific purity
func (p *IndianProvider) getGoldPriceByPurity(purity string, data *ibjaData, fetchErr error) float64 {
	switch purity {
	case "24K":
		return p.getGoldPriceWithFallback(data.gold999, 1.000, data, fetchErr)
	case "22K":
		return p.getGoldPriceWithFallback(data.gold916, 0.916, data, fetchErr)
	case "18K":
		return p.getGoldPriceWithFallback(data.gold750, 0.750, data, fetchErr)
	case "14K":
		return p.getGoldPriceWithFallback(data.gold585, 0.585, data, fetchErr)
	default:
		return 0
	}
}

// getGoldPriceWithFallback returns the price or calculates a fallback
func (p *IndianProvider) getGoldPriceWithFallback(price, multiplier float64, data *ibjaData, fetchErr error) float64 {
	if fetchErr == nil && price > 0 {
		return price
	}
	// Fallback: calculate from 24K base
	base := 14000.0
	if data != nil && data.gold999 > 0 {
		base = data.gold999
	}
	return base * multiplier
}

// getSilverPrice fetches silver price for a specific purity
func (p *IndianProvider) getSilverPrice(purity string) (*Price, error) {
	// Fetch IBJA data (cached)
	data, err := p.fetchIBJAData()

	var base999Price float64
	if err != nil || data.silver999 == 0 {
		// Use fallback price (updated Jan 2026 - IBJA rates ~₹2567 per 10g)
		base999Price = 257.0 // Approximate 999 silver price per gram in INR
	} else {
		base999Price = data.silver999
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
	// Fetch IBJA data (cached)
	data, err := p.fetchIBJAData()

	var base999Price float64
	if err != nil || data.platinum999 == 0 {
		// Use fallback price (updated Jan 2026 - IBJA rates ~₹7569 per 10g)
		base999Price = 757.0 // Approximate 999 platinum price per gram in INR
	} else {
		base999Price = data.platinum999
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

// fetchIBJAData fetches and caches all metal prices from IBJA website
func (p *IndianProvider) fetchIBJAData() (*ibjaData, error) {
	// Check cache
	if p.ibjaCache != nil && time.Since(p.cacheTime) < p.cacheTTL {
		return p.ibjaCache, nil
	}

	html, err := p.fetchIBJAHTML()
	if err != nil {
		return nil, err
	}

	// Parse all metal prices from the HTML
	data := p.parseIBJAHTML(html)

	// Cache the data
	p.ibjaCache = data
	p.cacheTime = time.Now()

	return data, nil
}

// fetchIBJAHTML fetches the HTML from IBJA website
func (p *IndianProvider) fetchIBJAHTML() (string, error) {
	url := "https://ibjarates.com/"

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching IBJA data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IBJA returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	return string(body), nil
}

// parseIBJAHTML parses metal prices from IBJA HTML
func (p *IndianProvider) parseIBJAHTML(html string) *ibjaData {
	data := &ibjaData{}

	// Parse Gold prices - per gram prices from GoldRatesCompare spans
	data.gold999 = p.extractPrice(html, `<span id="GoldRatesCompare999">(\d+)</span>`, 1.0)
	data.gold995 = p.extractPrice(html, `<span id="GoldRatesCompare995">(\d+)</span>`, 1.0)
	data.gold916 = p.extractPrice(html, `<span id="GoldRatesCompare916">(\d+)</span>`, 1.0)
	data.gold750 = p.extractPrice(html, `<span id="GoldRatesCompare750">(\d+)</span>`, 1.0)
	data.gold585 = p.extractPrice(html, `<span id="GoldRatesCompare585">(\d+)</span>`, 1.0)

	// Parse Silver and Platinum - from table (per 10 grams, convert to per gram)
	data.silver999 = p.extractPrice(html, `<span id="lblSilver999_PM">(\d+)</span>`, 0.1)
	data.platinum999 = p.extractPrice(html, `<span id="lblPlatinum999_PM">(\d+)</span>`, 0.1)

	return data
}

// extractPrice extracts and parses a price from HTML using regex
func (p *IndianProvider) extractPrice(html, pattern string, multiplier float64) float64 {
	matches := regexp.MustCompile(pattern).FindStringSubmatch(html)
	if len(matches) < 2 {
		return 0
	}
	price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64)
	if err != nil {
		return 0
	}
	return price * multiplier
}
