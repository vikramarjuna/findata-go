// Package main demonstrates the usage of the crypto package for fetching cryptocurrency prices.
package main

import (
	"fmt"
	"log"

	"github.com/Vikramarjuna/findata-go/crypto"
)

func main() {
	fmt.Println("=== Cryptocurrency Quotes Demo ===")

	// Example 1: Get specific crypto quote in USD
	fmt.Println("\n1. Get Bitcoin quote in USD:")
	btcQuote, err := crypto.Get("bitcoin")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bitcoin: $%.2f\n", btcQuote.CurrentPrice)
	fmt.Printf("Market Cap: $%.0f\n", btcQuote.MarketCap)
	fmt.Printf("24h Volume: $%.0f\n", btcQuote.Volume24h)
	fmt.Printf("24h Change: %.2f%%\n", btcQuote.PriceChangePct24h)

	// Example 2: Get crypto quote in INR
	fmt.Println("\n2. Get Ethereum quote in INR:")
	ethQuote, err := crypto.Get("ethereum", crypto.WithCurrency("INR"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Ethereum: ₹%.2f\n", ethQuote.CurrentPrice)
	fmt.Printf("Market Cap: ₹%.0f\n", ethQuote.MarketCap)
	fmt.Printf("24h Change: %.2f%%\n", ethQuote.PriceChangePct24h)

	// Example 3: Get multiple crypto quotes
	fmt.Println("\n3. Get multiple cryptocurrency quotes:")
	coinIDs := []string{"bitcoin", "ethereum", "solana", "cardano"}
	quotes, err := crypto.GetMultiple(coinIDs)
	if err != nil {
		log.Fatal(err)
	}

	for coinID, quote := range quotes {
		fmt.Printf("%s: $%.2f (%.2f%%)\n", coinID, quote.CurrentPrice, quote.PriceChangePct24h)
	}

	// Example 4: Get quotes in different currencies
	fmt.Println("\n4. Get Bitcoin quote in different currencies:")
	currencies := []string{"USD", "EUR", "INR"}
	for _, currency := range currencies {
		quote, err := crypto.Get("bitcoin", crypto.WithCurrency(currency))
		if err != nil {
			fmt.Printf("Error fetching %s quote: %v\n", currency, err)
			continue
		}
		fmt.Printf("bitcoin in %s: %.2f\n", currency, quote.CurrentPrice)
	}

	fmt.Println("\n✅ Demo completed successfully!")
	fmt.Println("\nNote: This example uses CoinGecko's free API which has rate limits.")
	fmt.Println("For production use, consider using an API key or implementing rate limiting.")
}
