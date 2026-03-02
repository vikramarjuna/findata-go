// Package main demonstrates the usage of the metals package for fetching precious metal prices.
package main

import (
	"fmt"
	"log"

	"github.com/Vikramarjuna/findata-go/metals"
)

func main() {
	fmt.Println("=== Metal Prices Demo ===")

	// Example 1: Get specific metal price
	fmt.Println("1. Get Gold 24K price:")
	goldPrice, err := metals.Get(metals.Gold, "24K")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %s %s: ₹%.2f per gram\n", goldPrice.Metal, goldPrice.Purity, goldPrice.PricePerGram)
	fmt.Printf("   Source: %s\n", goldPrice.Source)
	fmt.Printf("   Updated: %s\n\n", goldPrice.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Example 2: Get multiple gold purities
	fmt.Println("2. Gold prices for different purities:")
	goldPurities := []string{"24K", "22K", "18K", "14K"}
	for _, purity := range goldPurities {
		price, err := metals.Get(metals.Gold, purity)
		if err != nil {
			log.Printf("   Error fetching %s: %v\n", purity, err)
			continue
		}
		fmt.Printf("   %s: ₹%.2f per gram\n", purity, price.PricePerGram)
	}
	fmt.Println()

	// Example 3: Get silver prices
	fmt.Println("3. Silver prices:")
	silverPurities := []string{"999", "925"}
	for _, purity := range silverPurities {
		price, err := metals.Get(metals.Silver, purity)
		if err != nil {
			log.Printf("   Error fetching %s: %v\n", purity, err)
			continue
		}
		fmt.Printf("   %s: ₹%.2f per gram\n", purity, price.PricePerGram)
	}
	fmt.Println()

	// Example 4: Get platinum prices
	fmt.Println("4. Platinum prices:")
	platinumPurities := []string{"999", "950"}
	for _, purity := range platinumPurities {
		price, err := metals.Get(metals.Platinum, purity)
		if err != nil {
			log.Printf("   Error fetching %s: %v\n", purity, err)
			continue
		}
		fmt.Printf("   %s: ₹%.2f per gram\n", purity, price.PricePerGram)
	}
	fmt.Println()

	// Example 5: Get all prices at once
	fmt.Println("5. All metal prices:")
	allPrices, err := metals.GetAll()
	if err != nil {
		log.Fatal(err)
	}

	// Group by metal type
	metalGroups := make(map[metals.MetalType][]*metals.Price)
	for _, price := range allPrices {
		metalGroups[price.Metal] = append(metalGroups[price.Metal], price)
	}

	for metal, prices := range metalGroups {
		fmt.Printf("   %s:\n", metal)
		for _, price := range prices {
			fmt.Printf("      %s: ₹%.2f per gram\n", price.Purity, price.PricePerGram)
		}
	}
	fmt.Println()

	// Example 6: Calculate jewelry value
	fmt.Println("6. Calculate jewelry value:")
	fmt.Println("   22K Gold necklace - 25 grams")
	gold22K, _ := metals.Get(metals.Gold, "22K")
	necklaceWeight := 25.0
	metalValue := gold22K.PricePerGram * necklaceWeight
	makingCharges := metalValue * 0.10 // 10% making charges
	totalValue := metalValue + makingCharges
	fmt.Printf("   Metal value: ₹%.2f\n", metalValue)
	fmt.Printf("   Making charges (10%%): ₹%.2f\n", makingCharges)
	fmt.Printf("   Total value: ₹%.2f\n\n", totalValue)

	// Example 7: Compare metals
	fmt.Println("7. Metal price comparison (per gram):")
	gold24K, _ := metals.Get(metals.Gold, "24K")
	silver999, _ := metals.Get(metals.Silver, "999")
	platinum999, _ := metals.Get(metals.Platinum, "999")

	fmt.Printf("   Gold (24K):     ₹%.2f\n", gold24K.PricePerGram)
	fmt.Printf("   Platinum (999): ₹%.2f\n", platinum999.PricePerGram)
	fmt.Printf("   Silver (999):   ₹%.2f\n", silver999.PricePerGram)
	fmt.Println()

	// Example 8: Cache demonstration
	fmt.Println("8. Cache demonstration:")
	fmt.Println("   First call (fetches from provider)...")
	price1, _ := metals.Get(metals.Gold, "24K")
	fmt.Printf("   Price: ₹%.2f, Updated: %s\n", price1.PricePerGram, price1.UpdatedAt.Format("15:04:05.000"))

	fmt.Println("   Second call (uses cache)...")
	price2, _ := metals.Get(metals.Gold, "24K")
	fmt.Printf("   Price: ₹%.2f, Updated: %s\n", price2.PricePerGram, price2.UpdatedAt.Format("15:04:05.000"))

	if price1.UpdatedAt.Equal(price2.UpdatedAt) {
		fmt.Println("   ✓ Cache is working (same timestamp)")
	}
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}
