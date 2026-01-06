// Package main demonstrates how to use the mutual fund NAV functionality.
package main

import (
	"fmt"
	"log"

	"github.com/Vikramarjuna/findata-go/mf"
	// Uncomment to enable logging:
	// "github.com/Vikramarjuna/findata-go/logger"
)

func main() {
	// Optional: Enable logging (disabled by default)
	// Uncomment to see library logs:
	// logger.SetLogger(logger.NewSlogLogger(
	//     logger.WithLevel(logger.LevelInfo),  // Set to LevelDebug for detailed logs
	// ))

	// Search for mutual funds
	fmt.Println("Searching for 'HDFC Flexi Cap' funds...")
	results, err := mf.Search("HDFC Flexi Cap")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d matching funds:\n\n", len(results))
	for i, nav := range results {
		if i >= 5 { // Show only first 5
			fmt.Printf("... and %d more\n", len(results)-5)
			break
		}
		fmt.Printf("%d. %s\n", i+1, nav.SchemeName)
		fmt.Printf("   Code: %s | ISIN: %s\n", nav.SchemeCode, nav.ISIN)
		fmt.Printf("   NAV: ₹%.4f (as of %s)\n\n", nav.NAV, nav.Date)
	}

	// Get specific fund by ISIN
	fmt.Println("\nFetching specific fund by ISIN...")
	isin := "INF179K01997" // Example ISIN
	nav, err := mf.Get(isin)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Fund: %s\n", nav.SchemeName)
		fmt.Printf("NAV: ₹%.4f\n", nav.NAV)
		fmt.Printf("Date: %s\n", nav.Date)
	}

	// Get multiple funds
	fmt.Println("\nFetching multiple funds...")
	identifiers := []string{"119551", "119552", "119553"} // Example scheme codes
	navs, errors := mf.GetMultiple(identifiers)

	if len(errors) > 0 {
		fmt.Println("\nErrors encountered:")
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	fmt.Println("\nSuccessfully fetched NAVs:")
	for id, n := range navs {
		fmt.Printf("  %s: ₹%.4f\n", id, n.NAV)
	}
}
