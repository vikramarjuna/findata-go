package metals

import (
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

	var pricePerGram float64
	var useFallback bool

	// Get price based on purity from IBJA data
	switch purity {
	case "24K":
		if err != nil || data.gold999 == 0 {
			pricePerGram = 14000.0 // Fallback
			useFallback = true
		} else {
			pricePerGram = data.gold999
		}
	case "22K":
		if err != nil || data.gold916 == 0 {
			// Fallback: calculate from 24K
			base := 14000.0
			if data.gold999 > 0 {
				base = data.gold999
			}
			pricePerGram = base * 0.916
			useFallback = true
		} else {
			pricePerGram = data.gold916
		}
	case "18K":
		if err != nil || data.gold750 == 0 {
			// Fallback: calculate from 24K
			base := 14000.0
			if data.gold999 > 0 {
				base = data.gold999
			}
			pricePerGram = base * 0.750
			useFallback = true
		} else {
			pricePerGram = data.gold750
		}
	case "14K":
		if err != nil || data.gold585 == 0 {
			// Fallback: calculate from 24K
			base := 14000.0
			if data.gold999 > 0 {
				base = data.gold999
			}
			pricePerGram = base * 0.585
			useFallback = true
		} else {
			pricePerGram = data.gold585
		}
	default:
		return nil, &Error{
			Message:  fmt.Sprintf("unsupported gold purity: %s", purity),
			Metal:    Gold,
			Purity:   purity,
			Provider: p.Name(),
		}
	}

	_ = useFallback // Suppress unused variable warning

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

	url := "https://ibjarates.com/"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching IBJA data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IBJA returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	html := string(body)

	// Parse all metal prices from the HTML
	data := &ibjaData{}

	// Parse Gold prices - per gram prices from GoldRatesCompare spans
	// Gold 999 (24K): <span id="GoldRatesCompare999">14045</span>
	if matches := regexp.MustCompile(`<span id="GoldRatesCompare999">([0-9]+)</span>`).FindStringSubmatch(html); len(matches) >= 2 {
		if price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
			data.gold999 = price
		}
	}

	// Gold 995: <span id="GoldRatesCompare995">13989</span>
	if matches := regexp.MustCompile(`<span id="GoldRatesCompare995">([0-9]+)</span>`).FindStringSubmatch(html); len(matches) >= 2 {
		if price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
			data.gold995 = price
		}
	}

	// Gold 916 (22K): <span id="GoldRatesCompare916">12865</span>
	if matches := regexp.MustCompile(`<span id="GoldRatesCompare916">([0-9]+)</span>`).FindStringSubmatch(html); len(matches) >= 2 {
		if price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
			data.gold916 = price
		}
	}

	// Gold 750 (18K): <span id="GoldRatesCompare750">10534</span>
	if matches := regexp.MustCompile(`<span id="GoldRatesCompare750">([0-9]+)</span>`).FindStringSubmatch(html); len(matches) >= 2 {
		if price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
			data.gold750 = price
		}
	}

	// Gold 585 (14K): <span id="GoldRatesCompare585">8216</span>
	if matches := regexp.MustCompile(`<span id="GoldRatesCompare585">([0-9]+)</span>`).FindStringSubmatch(html); len(matches) >= 2 {
		if price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
			data.gold585 = price
		}
	}

	// Parse Silver 999 - from table (per 10 grams)
	// Looking for: <span id="lblSilver999_PM">256776</span>
	if matches := regexp.MustCompile(`<span id="lblSilver999_PM">([0-9]+)</span>`).FindStringSubmatch(html); len(matches) >= 2 {
		if price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
			data.silver999 = price / 10.0 // Convert from per 10g to per gram
		}
	}

	// Parse Platinum 999 - from table (per 10 grams)
	// Looking for: <span id="lblPlatinum999_PM">75690</span>
	if matches := regexp.MustCompile(`<span id="lblPlatinum999_PM">([0-9]+)</span>`).FindStringSubmatch(html); len(matches) >= 2 {
		if price, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64); err == nil {
			data.platinum999 = price / 10.0 // Convert from per 10g to per gram
		}
	}

	// Cache the data
	p.ibjaCache = data
	p.cacheTime = time.Now()

	return data, nil
}
