// Package mf provides access to AMFI mutual fund NAV data
package mf

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"github.com/Vikramarjuna/finance-india"
)

const (
	// AMFIURL is the AMFI portal URL for NAV data
	AMFIURL = "https://portal.amfiindia.com/spages/NAVAll.txt"
)

// NAV represents a mutual fund NAV entry
type NAV struct {
	SchemeCode string  `json:"scheme_code"`
	ISIN       string  `json:"isin"`       // Primary ISIN (Growth/Payout)
	ISINReinv  string  `json:"isin_reinv"` // Reinvestment ISIN
	SchemeName string  `json:"scheme_name"`
	NAV        float64 `json:"nav"`
	Date       string  `json:"date"`
}

// GetAll fetches all NAVs from AMFI portal
// Returns a map with multiple keys for flexible lookup:
// - scheme code
// - scheme name
// - ISIN (both primary and reinvestment)
func GetAll() (map[string]*NAV, error) {
	resp, err := finance.Client.Get(AMFIURL)
	if err != nil {
		return nil, &finance.Error{Message: fmt.Sprintf("failed to fetch AMFI data: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &finance.Error{
			Message: fmt.Sprintf("AMFI portal returned error"),
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
		return nil, &finance.Error{Message: fmt.Sprintf("error reading AMFI data: %v", err)}
	}

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
		return nil, &finance.Error{Message: fmt.Sprintf("NAV not found for: %s", identifier)}
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
