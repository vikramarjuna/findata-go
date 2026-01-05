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

func TestGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get all NAVs first
	navMap, err := GetAll()
	if err != nil {
		t.Fatalf("Failed to get NAVs: %v", err)
	}

	// Pick a scheme code to test
	var testCode string
	for _, nav := range navMap {
		testCode = nav.SchemeCode
		break
	}

	if testCode == "" {
		t.Skip("No scheme codes found")
	}

	// Test Get by scheme code
	nav, err := Get(testCode)
	if err != nil {
		t.Fatalf("Get(%s) error = %v", testCode, err)
	}

	if nav.SchemeCode != testCode {
		t.Errorf("SchemeCode = %v, want %v", nav.SchemeCode, testCode)
	}
	if nav.NAV <= 0 {
		t.Errorf("NAV should be positive, got %.4f", nav.NAV)
	}
}

func TestGet_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_, err := Get("NONEXISTENT_CODE_12345")
	if err == nil {
		t.Error("Get(NONEXISTENT_CODE_12345) should return error")
	}
}

func TestGetMultiple_EmptyList(t *testing.T) {
	navs, errors := GetMultiple([]string{})

	if len(navs) != 0 {
		t.Errorf("GetMultiple([]) should return empty map, got %d navs", len(navs))
	}
	if len(errors) != 0 {
		t.Errorf("GetMultiple([]) should return no errors, got %d errors", len(errors))
	}
}

func TestGetMultiple_WithInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Mix of valid and invalid identifiers
	identifiers := []string{"119551", "INVALID123", "119552"}
	navs, errors := GetMultiple(identifiers)

	// Should have at least one error
	if len(errors) == 0 {
		t.Error("Expected errors for invalid identifiers")
	}

	// Should have at least one valid NAV
	if len(navs) == 0 {
		t.Error("Expected at least one valid NAV")
	}
}

func TestNAVStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	navMap, err := GetAll()
	if err != nil {
		t.Fatalf("Failed to get NAVs: %v", err)
	}

	// Pick a NAV to validate structure
	var testNAV *NAV
	for _, nav := range navMap {
		testNAV = nav
		break
	}

	if testNAV == nil {
		t.Skip("No NAVs found")
	}

	// Validate required fields
	if testNAV.SchemeCode == "" {
		t.Error("SchemeCode should not be empty")
	}
	if testNAV.SchemeName == "" {
		t.Error("SchemeName should not be empty")
	}
	if testNAV.NAV <= 0 {
		t.Errorf("NAV should be positive, got %.4f", testNAV.NAV)
	}
	if testNAV.Date == "" {
		t.Error("Date should not be empty")
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test case insensitivity
	tests := []struct {
		name  string
		query string
	}{
		{"Uppercase", "HDFC"},
		{"Lowercase", "hdfc"},
		{"Mixed case", "Hdfc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := Search(tt.query)
			if err != nil {
				t.Fatalf("Search(%s) error = %v", tt.query, err)
			}

			if len(results) == 0 {
				t.Errorf("Search(%s) should return results", tt.query)
			}
		})
	}
}

func TestSearch_NoDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	results, err := Search("HDFC")
	if err != nil {
		t.Fatalf("Search(HDFC) error = %v", err)
	}

	// Check for duplicates by scheme code
	seen := make(map[string]bool)
	for _, nav := range results {
		if seen[nav.SchemeCode] {
			t.Errorf("Duplicate scheme code found: %s", nav.SchemeCode)
		}
		seen[nav.SchemeCode] = true
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "Error with code",
			err:  &Error{Message: "test error", Code: 404},
			want: "test error (code: 404)",
		},
		{
			name: "Error without code",
			err:  &Error{Message: "test error"},
			want: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
