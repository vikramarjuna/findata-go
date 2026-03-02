package metals

import (
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	// Clear cache before tests
	ClearCache()

	tests := []struct {
		name    string
		metal   MetalType
		purity  string
		wantErr bool
	}{
		{"Gold 24K", Gold, "24K", false},
		{"Gold 22K", Gold, "22K", false},
		{"Gold 18K", Gold, "18K", false},
		{"Gold 14K", Gold, "14K", false},
		{"Silver 999", Silver, "999", false},
		{"Silver 925", Silver, "925", false},
		{"Platinum 999", Platinum, "999", false},
		{"Platinum 950", Platinum, "950", false},
		{"Invalid purity", Gold, "99K", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := Get(tt.metal, tt.purity)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if price == nil {
				t.Error("Get() returned nil price")
				return
			}

			if price.Metal != tt.metal {
				t.Errorf("Metal = %v, want %v", price.Metal, tt.metal)
			}

			if price.Purity != tt.purity {
				t.Errorf("Purity = %v, want %v", price.Purity, tt.purity)
			}

			if price.PricePerGram <= 0 {
				t.Errorf("PricePerGram = %v, want > 0", price.PricePerGram)
			}

			if price.Currency != "INR" {
				t.Errorf("Currency = %v, want INR", price.Currency)
			}

			if price.UpdatedAt.IsZero() {
				t.Error("UpdatedAt is zero")
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	ClearCache()

	prices, err := GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(prices) == 0 {
		t.Error("GetAll() returned no prices")
	}

	// Check that we have prices for all metals
	hasGold := false
	hasSilver := false
	hasPlatinum := false

	for _, price := range prices {
		if price.Metal == Gold {
			hasGold = true
		}
		if price.Metal == Silver {
			hasSilver = true
		}
		if price.Metal == Platinum {
			hasPlatinum = true
		}

		// Validate each price
		if price.PricePerGram <= 0 {
			t.Errorf("Invalid price for %s %s: %v", price.Metal, price.Purity, price.PricePerGram)
		}
		if price.Currency != "INR" {
			t.Errorf("Invalid currency for %s %s: %v", price.Metal, price.Purity, price.Currency)
		}
	}

	if !hasGold {
		t.Error("GetAll() missing gold prices")
	}
	if !hasSilver {
		t.Error("GetAll() missing silver prices")
	}
	if !hasPlatinum {
		t.Error("GetAll() missing platinum prices")
	}
}

func TestCache(t *testing.T) {
	ClearCache()

	// First call should fetch from provider
	price1, err := Get(Gold, "24K")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Second call should use cache (same timestamp)
	price2, err := Get(Gold, "24K")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Timestamps should be identical (from cache)
	if !price1.UpdatedAt.Equal(price2.UpdatedAt) {
		t.Error("Cache not working - timestamps differ")
	}

	// Clear cache
	ClearCache()

	// Third call should fetch again (new timestamp)
	time.Sleep(10 * time.Millisecond) // Ensure time difference
	price3, err := Get(Gold, "24K")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if price1.UpdatedAt.Equal(price3.UpdatedAt) {
		t.Error("Cache clear not working - timestamps are same")
	}
}

func TestPriceRelationships(t *testing.T) {
	ClearCache()

	// Get gold prices
	gold24K, _ := Get(Gold, "24K")
	gold22K, _ := Get(Gold, "22K")
	gold18K, _ := Get(Gold, "18K")

	// 24K should be most expensive
	if gold24K.PricePerGram <= gold22K.PricePerGram {
		t.Error("24K gold should be more expensive than 22K")
	}
	if gold22K.PricePerGram <= gold18K.PricePerGram {
		t.Error("22K gold should be more expensive than 18K")
	}
}
