// Package mf provides access to AMFI mutual fund NAV data
package mf

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Vikramarjuna/findata-go/cache"
	"github.com/Vikramarjuna/findata-go/config"
)

const (
	// AMFIURL is the AMFI portal URL for NAV data
	AMFIURL = "https://portal.amfiindia.com/spages/NAVAll.txt"
	// Cache key for all NAVs
	cacheKeyAllNAVs = "mf:all_navs"
)

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

// ClearCache clears all cached NAVs
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

// NAV represents a mutual fund NAV entry
type NAV struct {
	SchemeCode string  `json:"scheme_code"`
	ISIN       string  `json:"isin"`       // Primary ISIN (Growth/Payout)
	ISINReinv  string  `json:"isin_reinv"` // Reinvestment ISIN
	SchemeName string  `json:"scheme_name"`
	NAV        float64 `json:"nav"`
	Date       string  `json:"date"`
}

// Error represents a mutual fund error
type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	if e.Code > 0 {
		return fmt.Sprintf("%s (code: %d)", e.Message, e.Code)
	}
	return e.Message
}

// GetAll fetches all NAVs from AMFI portal
// Returns a map with multiple keys for flexible lookup:
// - scheme code
// - scheme name
// - ISIN (both primary and reinvestment)
// Results are cached according to the global cache configuration
func GetAll() (map[string]*NAV, error) {
	// Check cache first
	c := getCache()
	if cached, ok := c.Get(cacheKeyAllNAVs); ok {
		if navMap, ok := cached.(map[string]*NAV); ok {
			return navMap, nil
		}
	}

	// Fetch from AMFI
	resp, err := config.GetHTTPClient().Get(AMFIURL)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("failed to fetch AMFI data: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &Error{
			Message: "AMFI portal returned error",
			Code:    resp.StatusCode,
		}
	}

	navMap := make(map[string]*NAV)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and headers
		if line == "" || strings.HasPrefix(line, "Scheme") ||
			strings.HasPrefix(line, "Open Ended") ||
			strings.HasPrefix(line, "Close Ended") ||
			strings.Contains(line, "(Erstwhile") {
			continue
		}

		// Format: Scheme Code;ISIN Div Payout/ISIN Growth;ISIN Div Reinvestment;Scheme Name;Net Asset Value;Date
		parts := strings.Split(line, ";")
		if len(parts) >= 6 {
			schemeCode := strings.TrimSpace(parts[0])
			isin := strings.TrimSpace(parts[1])      // ISIN Div Payout / Growth
			isinReinv := strings.TrimSpace(parts[2]) // ISIN Div Reinvestment
			schemeName := strings.TrimSpace(parts[3])
			navStr := strings.TrimSpace(parts[4])
			date := strings.TrimSpace(parts[5])

			// Skip if NAV is N.A.
			if navStr == "N.A." || navStr == "" {
				continue
			}

			var nav float64
			fmt.Sscanf(navStr, "%f", &nav)

			if nav > 0 {
				navEntry := &NAV{
					SchemeCode: schemeCode,
					ISIN:       isin,
					ISINReinv:  isinReinv,
					SchemeName: schemeName,
					NAV:        nav,
					Date:       date,
				}
				// Index by scheme code, scheme name, and ISINs for flexible lookup
				navMap[schemeCode] = navEntry
				navMap[schemeName] = navEntry
				if isin != "" && isin != "-" {
					navMap[isin] = navEntry
				}
				if isinReinv != "" && isinReinv != "-" {
					navMap[isinReinv] = navEntry
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, &Error{Message: fmt.Sprintf("error reading AMFI data: %v", err)}
	}

	// Cache the result
	c.Set(cacheKeyAllNAVs, navMap)

	return navMap, nil
}

// Get fetches NAV for a specific scheme by code, name, or ISIN
func Get(identifier string) (*NAV, error) {
	navMap, err := GetAll()
	if err != nil {
		return nil, err
	}

	nav, ok := navMap[identifier]
	if !ok {
		return nil, &Error{Message: fmt.Sprintf("NAV not found for: %s", identifier)}
	}

	return nav, nil
}

// Search searches for mutual funds by name (case-insensitive partial match)
// Returns all matching NAVs
func Search(query string) ([]*NAV, error) {
	navMap, err := GetAll()
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var results []*NAV
	seen := make(map[string]bool) // To avoid duplicates

	for key, nav := range navMap {
		// Only search by scheme name to avoid duplicates
		if key == nav.SchemeName && strings.Contains(strings.ToLower(nav.SchemeName), queryLower) {
			if !seen[nav.SchemeCode] {
				results = append(results, nav)
				seen[nav.SchemeCode] = true
			}
		}
	}

	return results, nil
}

// GetMultiple fetches NAVs for multiple identifiers
// Returns a map of identifier -> NAV, and any errors encountered
func GetMultiple(identifiers []string) (map[string]*NAV, []error) {
	navMap, err := GetAll()
	if err != nil {
		return nil, []error{err}
	}

	results := make(map[string]*NAV)
	var errors []error

	for _, id := range identifiers {
		nav, ok := navMap[id]
		if !ok {
			errors = append(errors, fmt.Errorf("%s: NAV not found", id))
			continue
		}
		results[id] = nav
	}

	return results, errors
}
