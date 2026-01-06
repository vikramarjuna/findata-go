// Package main demonstrates how to fetch NSE stock quotes.
package main

import (
	"fmt"
	"log"

	"github.com/Vikramarjuna/findata-go/config"
	"github.com/Vikramarjuna/findata-go/equity"
)

func main() {
	// Optional: Enable logging (disabled by default)
	// Uncomment to see library logs:
	// logger.SetLogger(logger.NewSlogLogger(
	//     logger.WithLevel(logger.LevelInfo),  // Change to LevelDebug for verbose output
	//     logger.WithFormat(logger.FormatText),
	// ))

	// Optional: Log to a file
	// logFile, _ := os.OpenFile("findata.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// defer logFile.Close()
	// logger.SetLogger(logger.NewSlogLogger(
	//     logger.WithLevel(logger.LevelInfo),
	//     logger.WithOutput(logFile),
	// ))

	// Optional: Set default market (defaults to India)
	config.SetDefaultMarket(config.MarketIndia)

	// Fetch a single quote - library auto-detects it's NSE
	fmt.Println("Fetching quote for RELIANCE...")
	quote, err := equity.Get("RELIANCE")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== %s (%s) ===\n", quote.Symbol, quote.Exchange)
	fmt.Printf("Company: %s\n", quote.CompanyName)
	fmt.Printf("Sector: %s\n", quote.Sector)
	fmt.Printf("Industry: %s\n", quote.Industry)
	fmt.Printf("Last Price: %s %.2f\n", quote.Currency, quote.LastPrice)
	fmt.Printf("Change: %.2f (%.2f%%)\n", quote.Change, quote.PChange)
	fmt.Printf("Day Range: %.2f - %.2f\n", quote.DayLow, quote.DayHigh)
	fmt.Printf("52 Week Range: %.2f - %.2f\n", quote.YearLow, quote.YearHigh)
	fmt.Printf("Indices: %v\n", quote.Indices)

	// Fetch multiple quotes - all auto-detected
	fmt.Println("\n\nFetching multiple quotes...")
	symbols := []string{"TCS", "INFY", "WIPRO", "HDFCBANK"}
	quotes, errors := equity.GetMultiple(symbols)

	if len(errors) > 0 {
		fmt.Println("\nErrors encountered:")
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	fmt.Println("\nSuccessfully fetched quotes:")
	for symbol, q := range quotes {
		fmt.Printf("  %s (%s): %s %.2f (%.2f%%)\n",
			symbol, q.Exchange, q.Currency, q.LastPrice, q.PChange)
	}

	// Example: Explicitly specify exchange
	fmt.Println("\n\nFetching with explicit exchange...")
	quote2, err := equity.Get("RELIANCE", equity.WithExchange(config.ExchangeNSE))
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%s: %s %.2f\n", quote2.Symbol, quote2.Currency, quote2.LastPrice)
	}
}
