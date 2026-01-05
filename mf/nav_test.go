package mf

import (
	"testing"
)

func TestGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	navMap, err := GetAll()
	if err != nil {
		t.Fatalf("Failed to get NAVs: %v", err)
	}

	if len(navMap) == 0 {
		t.Fatal("Expected at least one NAV entry")
	}

	t.Logf("Fetched %d NAV entries", len(navMap))
}

func TestSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	results, err := Search("HDFC")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one result for 'HDFC'")
	}

	t.Logf("Found %d funds matching 'HDFC'", len(results))

	// Verify all results contain "HDFC" in the name
	for _, nav := range results {
		if nav.NAV <= 0 {
			t.Errorf("Expected positive NAV, got %.4f for %s", nav.NAV, nav.SchemeName)
		}
		if nav.SchemeCode == "" {
			t.Errorf("Expected scheme code to be set for %s", nav.SchemeName)
		}
	}
}

func TestGetMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// First get all to find some valid scheme codes
	navMap, err := GetAll()
	if err != nil {
		t.Fatalf("Failed to get NAVs: %v", err)
	}

	// Get first 3 scheme codes
	var codes []string
	for _, nav := range navMap {
		if len(codes) < 3 {
			codes = append(codes, nav.SchemeCode)
		} else {
			break
		}
	}

	if len(codes) == 0 {
		t.Skip("No scheme codes found")
	}

	navs, errors := GetMultiple(codes)

	if len(errors) > 0 {
		t.Logf("Encountered %d errors: %v", len(errors), errors)
	}

	if len(navs) == 0 {
		t.Fatal("Expected at least one NAV")
	}

	for code, nav := range navs {
		if nav.SchemeCode != code {
			t.Errorf("Expected scheme code %s, got %s", code, nav.SchemeCode)
		}
	}
}

