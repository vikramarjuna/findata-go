// Package mf provides access to AMFI mutual fund NAV data
package mf

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Vikramarjuna/findata-go/cache"
	"github.com/Vikramarjuna/findata-go/config"
	"github.com/Vikramarjuna/findata-go/logger"
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
	logger.Debug("fetching all NAVs from AMFI")

	// Check cache first
	c := getCache()
	if cached, ok := c.Get(cacheKeyAllNAVs); ok {
		if navMap, ok := cached.(map[string]*NAV); ok {
			logger.Debug("returning cached NAV data", "count", len(navMap))
			return navMap, nil
		}
	}

	logger.Info("cache miss, fetching NAV data from AMFI", "url", AMFIURL)

	// Fetch from AMFI
	resp, err := fetchAMFIData()
	if err != nil {
		logger.Error("failed to fetch AMFI data", "error", err)
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Parse the response
	navMap, err := parseAMFIData(resp.Body)
	if err != nil {
		logger.Error("failed to parse AMFI data", "error", err)
		return nil, err
	}

	logger.Info("successfully fetched and parsed NAV data", "count", len(navMap))

	// Cache the result
	c.Set(cacheKeyAllNAVs, navMap)

	return navMap, nil
}

func fetchAMFIData() (*http.Response, error) {
	logger.Debug("creating HTTP request to AMFI", "url", AMFIURL)
	req, err := http.NewRequestWithContext(context.Background(), "GET", AMFIURL, http.NoBody)
	if err != nil {
		logger.Error("failed to create HTTP request", "error", err)
		return nil, &Error{Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	logger.Debug("sending HTTP request to AMFI")
	resp, err := config.GetHTTPClient().Do(req)
	if err != nil {
		logger.Error("HTTP request to AMFI failed", "error", err, "url", AMFIURL)
		return nil, &Error{Message: fmt.Sprintf("failed to fetch AMFI data: %v", err)}
	}

	if resp.StatusCode != http.StatusOK {
		logger.Warn("AMFI returned non-OK status", "status_code", resp.StatusCode)
		_ = resp.Body.Close()
		return nil, &Error{
			Message: "AMFI portal returned error",
			Code:    resp.StatusCode,
		}
	}

	return resp, nil
}

func parseAMFIData(body interface{ Read([]byte) (int, error) }) (map[string]*NAV, error) {
	navMap := make(map[string]*NAV)
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := scanner.Text()
		if shouldSkipLine(line) {
			continue
		}

		navEntry := parseNAVLine(line)
		if navEntry == nil {
			continue
		}

		indexNAVEntry(navMap, navEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, &Error{Message: fmt.Sprintf("error reading AMFI data: %v", err)}
	}

	return navMap, nil
}

func shouldSkipLine(line string) bool {
	return line == "" ||
		strings.HasPrefix(line, "Scheme") ||
		strings.HasPrefix(line, "Open Ended") ||
		strings.HasPrefix(line, "Close Ended") ||
		strings.Contains(line, "(Erstwhile")
}

func parseNAVLine(line string) *NAV {
	parts := strings.Split(line, ";")
	if len(parts) < 6 {
		return nil
	}

	schemeCode := strings.TrimSpace(parts[0])
	isin := strings.TrimSpace(parts[1])
	isinReinv := strings.TrimSpace(parts[2])
	schemeName := strings.TrimSpace(parts[3])
	navStr := strings.TrimSpace(parts[4])
	date := strings.TrimSpace(parts[5])

	if navStr == "N.A." || navStr == "" {
		return nil
	}

	var nav float64
	_, _ = fmt.Sscanf(navStr, "%f", &nav)

	if nav <= 0 {
		return nil
	}

	return &NAV{
		SchemeCode: schemeCode,
		ISIN:       isin,
		ISINReinv:  isinReinv,
		SchemeName: schemeName,
		NAV:        nav,
		Date:       date,
	}
}

func indexNAVEntry(navMap map[string]*NAV, navEntry *NAV) {
	navMap[navEntry.SchemeCode] = navEntry
	navMap[navEntry.SchemeName] = navEntry
	if navEntry.ISIN != "" && navEntry.ISIN != "-" {
		navMap[navEntry.ISIN] = navEntry
	}
	if navEntry.ISINReinv != "" && navEntry.ISINReinv != "-" {
		navMap[navEntry.ISINReinv] = navEntry
	}
}

// Get fetches NAV for a specific scheme by code, name, or ISIN
func Get(identifier string) (*NAV, error) {
	logger.Debug("fetching NAV", "identifier", identifier)

	navMap, err := GetAll()
	if err != nil {
		return nil, err
	}

	nav, ok := navMap[identifier]
	if !ok {
		logger.Warn("NAV not found", "identifier", identifier)
		return nil, &Error{Message: fmt.Sprintf("NAV not found for: %s", identifier)}
	}

	logger.Debug("NAV found", "identifier", identifier, "scheme_name", nav.SchemeName, "nav", nav.NAV)
	return nav, nil
}

// Search searches for mutual funds by name (case-insensitive partial match)
// Returns all matching NAVs
func Search(query string) ([]*NAV, error) {
	logger.Debug("searching for mutual funds", "query", query)

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

	logger.Info("search completed", "query", query, "results_count", len(results))
	return results, nil
}

// GetMultiple fetches NAVs for multiple identifiers
// Returns a map of identifier -> NAV, and any errors encountered
func GetMultiple(identifiers []string) (map[string]*NAV, []error) {
	logger.Debug("fetching multiple NAVs", "count", len(identifiers))

	navMap, err := GetAll()
	if err != nil {
		logger.Error("failed to get all NAVs for batch fetch", "error", err)
		return nil, []error{err}
	}

	results := make(map[string]*NAV)
	var errors []error

	for _, id := range identifiers {
		nav, ok := navMap[id]
		if !ok {
			logger.Debug("NAV not found in batch", "identifier", id)
			errors = append(errors, fmt.Errorf("%s: NAV not found", id))
			continue
		}
		results[id] = nav
	}

	logger.Info("batch NAV fetch completed", "requested", len(identifiers), "found", len(results), "errors", len(errors))
	return results, errors
}
