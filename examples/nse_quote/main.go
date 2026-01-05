package main

import (
	"fmt"
	"log"

	"github.com/Vikramarjuna/finance-india/nse"
)

func main() {
	// Fetch a single quote
	fmt.Println("Fetching quote for RELIANCE...")
	quote, err := nse.Get("RELIANCE")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== %s ===\n", quote.Symbol)
	fmt.Printf("Company: %s\n", quote.CompanyName)
	fmt.Printf("Sector: %s\n", quote.Sector)
	fmt.Printf("Industry: %s\n", quote.Industry)
	fmt.Printf("Last Price: ₹%.2f\n", quote.LastPrice)
	fmt.Printf("Change: ₹%.2f (%.2f%%)\n", quote.Change, quote.PChange)
	fmt.Printf("Day Range: ₹%.2f - ₹%.2f\n", quote.DayLow, quote.DayHigh)
	fmt.Printf("52 Week Range: ₹%.2f - ₹%.2f\n", quote.YearLow, quote.YearHigh)
	fmt.Printf("Indices: %v\n", quote.Indices)

	// Fetch multiple quotes
	fmt.Println("\n\nFetching multiple quotes...")
	symbols := []string{"TCS", "INFY", "WIPRO", "HDFCBANK"}
	quotes, errors := nse.GetMultiple(symbols)

	if len(errors) > 0 {
		fmt.Println("\nErrors encountered:")
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	fmt.Println("\nSuccessfully fetched quotes:")
	for symbol, q := range quotes {
		fmt.Printf("  %s: ₹%.2f (%.2f%%)\n", symbol, q.LastPrice, q.PChange)
	}
}

