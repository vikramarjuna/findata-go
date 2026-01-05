package mf

import (
	"fmt"
	"strings"
)

// FindOptions configures the find behavior
type FindOptions struct {
	// ISIN to search by (highest priority)
	ISIN string
	// SchemeCode to search by (second priority)
	SchemeCode string
	// PreferDirect prefers Direct plans over Regular plans
	PreferDirect bool
	// PreferGrowth prefers Growth plans over IDCW plans
	PreferGrowth bool
	// ExactMatch requires exact name matching (no fuzzy matching)
	ExactMatch bool
}

// Find searches for a mutual fund NAV using multiple strategies
// Returns the best matching NAV or an error if not found
//
// Find strategies (in order of priority):
// 1. Exact match by ISIN (if provided in options)
// 2. Exact match by scheme code (if provided in options)
// 3. Exact match by name
// 4. Name with " - Growth" suffix
// 5. Partial name match (case-insensitive)
// 6. Normalized name match (removes plan/option suffixes)
// 7. Any partial match as last resort
//
// The function prefers Direct plans over Regular plans and Growth over IDCW
// when multiple matches are found.
func Find(name string, options FindOptions) (*NAV, error) {
	// Fetch all NAVs
	navMap, err := GetAll()
	if err != nil {
		return nil, err
	}

	return FindInMap(navMap, name, options)
}

// FindInMap searches for a NAV in a pre-fetched map
// This is useful when you want to search multiple times without re-fetching
func FindInMap(navMap map[string]*NAV, name string, options FindOptions) (*NAV, error) {
	// Strategy 1: Try ISIN first (most reliable)
	if options.ISIN != "" {
		if nav, ok := navMap[options.ISIN]; ok {
			return nav, nil
		}
	}

	// Strategy 2: Try scheme code
	if options.SchemeCode != "" {
		if nav, ok := navMap[options.SchemeCode]; ok {
			return nav, nil
		}
	}

	// Strategy 3: Exact match by name
	if nav, ok := navMap[name]; ok {
		return nav, nil
	}

	// If exact match is required, stop here
	if options.ExactMatch {
		return nil, &Error{Message: fmt.Sprintf("no exact match found for: %s", name)}
	}

	// Strategy 4: Try with " - Growth" suffix (most common plan type)
	growthName := name + " - Growth"
	if nav, ok := navMap[growthName]; ok {
		return nav, nil
	}

	// Strategy 5: Partial match with scoring
	searchLower := strings.ToLower(name)
	var partialMatches []*NAV
	var partialKeys []string

	for key, nav := range navMap {
		if strings.Contains(strings.ToLower(key), searchLower) {
			partialMatches = append(partialMatches, nav)
			partialKeys = append(partialKeys, key)
		}
	}

	if len(partialMatches) > 0 {
		bestMatch := findBestMatchFromList(partialMatches, partialKeys, options)
		if bestMatch != nil {
			return bestMatch, nil
		}
	}

	// Strategy 6: Normalized name match
	normalizedSearch := normalizeName(name)
	bestMatch := findBestMatch(navMap, normalizedSearch, options)
	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, &Error{Message: fmt.Sprintf("no match found for: %s", name)}
}

// normalizeName extracts the core fund name for fuzzy matching
// Removes common suffixes that vary between sources
func normalizeName(name string) string {
	name = strings.ToLower(name)
	// Remove common suffixes that vary between sources (order matters - longer first)
	suffixes := []string{
		" - direct plan - growth option",
		" - direct plan - idcw option",
		" - regular plan - growth option",
		" - regular plan - idcw option",
		" - direct - growth",
		" - direct - idcw",
		" - regular - growth",
		" - regular - idcw",
		" - direct plan",
		" - regular plan",
		" - growth option",
		" - idcw option",
		" - dividend option",
		" - direct",
		" - regular",
		" - growth",
		" - idcw",
		" - dividend",
		" option",
		" plan",
	}
	// Keep removing suffixes until no more can be removed
	changed := true
	for changed {
		changed = false
		for _, suffix := range suffixes {
			if strings.HasSuffix(name, suffix) {
				name = strings.TrimSuffix(name, suffix)
				changed = true
				break
			}
		}
	}
	return strings.TrimSpace(name)
}

// findBestMatchFromList finds the best match from a list of NAVs and their keys
func findBestMatchFromList(navs []*NAV, keys []string, options FindOptions) *NAV {
	var bestMatch *NAV
	var bestScore int

	for i, nav := range navs {
		keyLower := strings.ToLower(keys[i])
		score := calculateMatchScore(keyLower, options)
		if score > bestScore || bestMatch == nil {
			bestScore = score
			bestMatch = nav
		}
	}

	return bestMatch
}

// findBestMatch finds the best matching NAV based on normalized names
func findBestMatch(navMap map[string]*NAV, normalizedSearch string, options FindOptions) *NAV {
	var bestMatch *NAV
	var bestScore int

	for key, nav := range navMap {
		normalizedKey := normalizeName(key)
		keyLower := strings.ToLower(key)

		// Check if core names match
		if !strings.Contains(normalizedKey, normalizedSearch) && !strings.Contains(normalizedSearch, normalizedKey) {
			continue
		}

		score := calculateMatchScore(keyLower, options)
		if score > bestScore {
			bestScore = score
			bestMatch = nav
		}
	}

	return bestMatch
}

// calculateMatchScore calculates a score for a match based on preferences
// Higher score = better match
func calculateMatchScore(key string, options FindOptions) int {
	score := 0

	// Prefer Direct plans
	if options.PreferDirect && strings.Contains(key, "direct") {
		score += 10
	}

	// Prefer Growth plans
	if options.PreferGrowth && strings.Contains(key, "growth") {
		score += 5
	}

	// Penalize Regular plans if Direct is preferred
	if options.PreferDirect && strings.Contains(key, "regular") {
		score -= 10
	}

	// Penalize IDCW if Growth is preferred
	if options.PreferGrowth && (strings.Contains(key, "idcw") || strings.Contains(key, "dividend")) {
		score -= 5
	}

	return score
}

// FindByISIN searches for a NAV by ISIN
func FindByISIN(isin string) (*NAV, error) {
	return Find("", FindOptions{ISIN: isin})
}

// FindByCode searches for a NAV by scheme code
func FindByCode(code string) (*NAV, error) {
	return Find("", FindOptions{SchemeCode: code})
}

// FindDirect searches for a Direct plan NAV by name
func FindDirect(name string) (*NAV, error) {
	return Find(name, FindOptions{
		PreferDirect: true,
		PreferGrowth: true,
	})
}

// FindGrowth searches for a Growth plan NAV by name
func FindGrowth(name string) (*NAV, error) {
	return Find(name, FindOptions{
		PreferGrowth: true,
	})
}
