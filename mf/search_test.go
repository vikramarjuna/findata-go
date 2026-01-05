package mf

import (
	"testing"
)

// createTestNAVMap creates a test NAV map for testing
func createTestNAVMap() map[string]*NAV {
	navMap := make(map[string]*NAV)

	// Add some test NAVs
	navs := []*NAV{
		{
			SchemeCode: "120503",
			ISIN:       "INF082J01014",
			ISINReinv:  "INF082J01022",
			SchemeName: "Quant Mid Cap Fund - Direct Plan - Growth Option",
			NAV:        150.50,
			Date:       "01-Jan-2024",
		},
		{
			SchemeCode: "120504",
			ISIN:       "INF082J01030",
			ISINReinv:  "INF082J01048",
			SchemeName: "Quant Mid Cap Fund - Regular Plan - Growth Option",
			NAV:        145.25,
			Date:       "01-Jan-2024",
		},
		{
			SchemeCode: "120505",
			ISIN:       "INF082J01055",
			ISINReinv:  "INF082J01063",
			SchemeName: "Quant Mid Cap Fund - Direct Plan - IDCW Option",
			NAV:        148.75,
			Date:       "01-Jan-2024",
		},
		{
			SchemeCode: "100001",
			ISIN:       "INF200K01010",
			ISINReinv:  "-",
			SchemeName: "HDFC Flexi Cap Fund - Direct Plan - Growth",
			NAV:        200.00,
			Date:       "01-Jan-2024",
		},
	}

	// Index by all keys
	for _, nav := range navs {
		navMap[nav.SchemeCode] = nav
		navMap[nav.SchemeName] = nav
		if nav.ISIN != "" && nav.ISIN != "-" {
			navMap[nav.ISIN] = nav
		}
		if nav.ISINReinv != "" && nav.ISINReinv != "-" {
			navMap[nav.ISINReinv] = nav
		}
	}

	return navMap
}

func TestFindByISIN(t *testing.T) {
	navMap := createTestNAVMap()

	nav, err := FindInMap(navMap, "", FindOptions{ISIN: "INF082J01014"})
	if err != nil {
		t.Fatalf("Expected to find NAV by ISIN, got error: %v", err)
	}
	if nav.SchemeCode != "120503" {
		t.Errorf("Expected scheme code 120503, got %s", nav.SchemeCode)
	}
}

func TestFindByCode(t *testing.T) {
	navMap := createTestNAVMap()

	nav, err := FindInMap(navMap, "", FindOptions{SchemeCode: "120503"})
	if err != nil {
		t.Fatalf("Expected to find NAV by code, got error: %v", err)
	}
	if nav.SchemeName != "Quant Mid Cap Fund - Direct Plan - Growth Option" {
		t.Errorf("Unexpected scheme name: %s", nav.SchemeName)
	}
}

func TestFindExactMatch(t *testing.T) {
	navMap := createTestNAVMap()

	nav, err := FindInMap(navMap, "Quant Mid Cap Fund - Direct Plan - Growth Option", FindOptions{})
	if err != nil {
		t.Fatalf("Expected exact match, got error: %v", err)
	}
	if nav.SchemeCode != "120503" {
		t.Errorf("Expected scheme code 120503, got %s", nav.SchemeCode)
	}
}

func TestFindPreferDirect(t *testing.T) {
	navMap := createTestNAVMap()

	nav, err := FindInMap(navMap, "Quant Mid Cap Fund", FindOptions{
		PreferDirect: true,
		PreferGrowth: true,
	})
	if err != nil {
		t.Fatalf("Expected to find NAV, got error: %v", err)
	}
	// Should prefer Direct + Growth
	if nav.SchemeCode != "120503" {
		t.Errorf("Expected Direct Growth plan (120503), got %s", nav.SchemeCode)
	}
}

func TestFindPreferGrowth(t *testing.T) {
	navMap := createTestNAVMap()

	nav, err := FindInMap(navMap, "Quant Mid Cap", FindOptions{
		PreferGrowth: true,
	})
	if err != nil {
		t.Fatalf("Expected to find NAV, got error: %v", err)
	}
	// Should prefer Growth over IDCW
	if nav.SchemeName == "Quant Mid Cap Fund - Direct Plan - IDCW Option" {
		t.Error("Should not return IDCW when PreferGrowth is true")
	}
}

func TestFindNotFound(t *testing.T) {
	navMap := createTestNAVMap()

	_, err := FindInMap(navMap, "NonExistent Fund", FindOptions{})
	if err == nil {
		t.Error("Expected error for non-existent fund")
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Fund Name - Direct Plan", "fund name"},
		{"Fund Name - Growth Option", "fund name"},
		{"Fund Name - IDCW Option", "fund name"},
		{"Fund Name - Direct - Growth", "fund name"},
		{"Fund Name", "fund name"},
	}

	for _, tt := range tests {
		result := normalizeName(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

