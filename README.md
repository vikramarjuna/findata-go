# findata-go

[![CI](https://github.com/Vikramarjuna/findata-go/workflows/CI/badge.svg)](https://github.com/Vikramarjuna/findata-go/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/Vikramarjuna/findata-go?status.svg)](https://godoc.org/github.com/Vikramarjuna/findata-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Vikramarjuna/findata-go)](https://goreportcard.com/report/github.com/Vikramarjuna/findata-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/Vikramarjuna/findata-go)](https://github.com/Vikramarjuna/findata-go/releases)

A unified Go library for accessing financial markets data across multiple exchanges and countries.

**Currently Supported:**
- 🇮🇳 Indian equities (NSE)
- 🇮🇳 Indian mutual funds (AMFI)
- 💰 Precious metals (Gold, Silver, Platinum)

**Coming Soon:**
- 🇺🇸 US equities (NYSE, NASDAQ)
- 🇮🇳 BSE stocks

Inspired by [finance-go](https://github.com/piquette/finance-go), this library provides a clean, exchange-agnostic API that automatically routes symbols to the appropriate data provider.

## Features

✅ **Exchange-Agnostic API** - Just use symbols, library handles routing
✅ **Auto-Detection** - Automatically detects market from symbol pattern
✅ **Configurable** - Set default markets and override per-request
✅ **Batch Fetching** - Fetch multiple quotes efficiently
✅ **Zero Dependencies** - Only uses Go standard library
✅ **Type-Safe** - Comprehensive type definitions
✅ **Well-Tested** - Includes integration tests
✅ **Structured Logging** - Built-in logging with multiple levels and formats

## Installation

**Requirements:** Go 1.22 or higher (1.24 recommended)

```bash
go get github.com/Vikramarjuna/findata-go
```

## Quick Start

### Equity Quotes (Exchange-Agnostic)

```go
package main

import (
    "fmt"
    "log"

    "github.com/Vikramarjuna/findata-go/equity"
)

func main() {
    // Just use the symbol - library auto-detects the exchange!
    quote, err := equity.Get("RELIANCE")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%s (%s): %s %.2f (%.2f%%)\n",
        quote.Symbol, quote.Exchange, quote.Currency,
        quote.LastPrice, quote.PChange)

    // Batch fetch multiple symbols
    symbols := []string{"TCS", "INFY", "WIPRO"}
    quotes, _ := equity.GetMultiple(symbols)

    for symbol, q := range quotes {
        fmt.Printf("%s: %s %.2f\n", symbol, q.Currency, q.LastPrice)
    }
}
```

**Output:**
```
RELIANCE (NSE): INR 1577.00 (-0.96%)
TCS: INR 3213.00
INFY: INR 1604.10
WIPRO: INR 263.00
```

### Configuration & Options

```go
import (
    "github.com/Vikramarjuna/findata-go/config"
    "github.com/Vikramarjuna/findata-go/equity"
)

func main() {
    // Set default market (optional - defaults to India)
    config.SetDefaultMarket(config.MarketIndia)

    // Now all symbols default to Indian market
    quote, _ := equity.Get("RELIANCE")  // Uses NSE

    // Override per-request with options
    quote, _ = equity.Get("RELIANCE",
        equity.WithExchange(config.ExchangeNSE))

    // Or specify market
    quote, _ = equity.Get("RELIANCE",
        equity.WithMarket(config.MarketIndia))

    // Custom HTTP client
    client := &http.Client{Timeout: 60 * time.Second}
    config.SetHTTPClient(client)
}
```

### Mutual Fund NAVs

```go
import "github.com/Vikramarjuna/findata-go/mf"

func main() {
    // Search for funds
    results, _ := mf.Search("HDFC Flexi Cap")
    for _, nav := range results {
        fmt.Printf("%s: ₹%.4f\n", nav.SchemeName, nav.NAV)
    }

    // Get by ISIN, scheme code, or name
    nav, _ := mf.Get("INF179K01997")
    fmt.Printf("NAV: ₹%.4f (as of %s)\n", nav.NAV, nav.Date)

    // Batch fetch
    navs, _ := mf.GetMultiple([]string{"119551", "119552"})
}
```

### Metal Prices

```go
import "github.com/Vikramarjuna/findata-go/metals"

func main() {
    // Get specific metal price
    goldPrice, _ := metals.Get(metals.Gold, "24K")
    fmt.Printf("Gold 24K: ₹%.2f per gram\n", goldPrice.PricePerGram)

    // Get all metal prices
    allPrices, _ := metals.GetAll()
    for _, price := range allPrices {
        fmt.Printf("%s %s: ₹%.2f per gram\n",
            price.Metal, price.Purity, price.PricePerGram)
    }

    // Supported metals and purities:
    // Gold: 24K, 22K, 18K, 14K
    // Silver: 999, 925
    // Platinum: 999, 950
}
```

## API Reference

### Equity Package

The main package for fetching stock quotes across all supported exchanges.

**Functions:**

- `equity.Get(symbol string, opts ...Option) (*Quote, error)` - Fetch a single quote
- `equity.GetMultiple(symbols []string, opts ...Option) (map[string]*Quote, []error)` - Batch fetch

**Options:**

- `equity.WithMarket(market config.Market)` - Override default market
- `equity.WithExchange(exchange config.Exchange)` - Specify exact exchange

**Quote Structure:**

```go
type Quote struct {
    Symbol        string
    Exchange      config.Exchange  // NSE, BSE, NYSE, NASDAQ
    CompanyName   string
    Industry      string
    Sector        string
    LastPrice     float64
    Currency      string           // INR, USD, etc.
    Change        float64
    PChange       float64
    PreviousClose float64
    Open          float64
    DayHigh       float64
    DayLow        float64
    YearHigh      float64
    YearLow       float64
    Volume        float64
    Value         float64
    Indices       []string
    Metadata      map[string]string
}
```

### Config Package

Global configuration for the library.

**Functions:**

- `config.SetDefaultMarket(market Market)` - Set default market (India, US, Auto)
- `config.GetDefaultMarket() Market` - Get current default market
- `config.SetHTTPClient(client *http.Client)` - Set custom HTTP client
- `config.SetAPIKey(provider, key string)` - Set API key for providers
- `config.SetUserAgent(ua string)` - Set custom user agent

**Constants:**

```go
// Markets
config.MarketIndia  // "IN"
config.MarketUS     // "US"
config.MarketAuto   // "AUTO"

// Exchanges
config.ExchangeNSE     // "NSE"
config.ExchangeBSE     // "BSE"
config.ExchangeNYSE    // "NYSE"
config.ExchangeNASDAQ  // "NASDAQ"
```

### MF Package

Mutual fund NAV data from AMFI.

**Functions:**

- `mf.GetAll() (map[string]*NAV, error)` - Fetch all NAVs
- `mf.Get(identifier string) (*NAV, error)` - Get by code/ISIN/name
- `mf.Search(query string) ([]*NAV, error)` - Search by name
- `mf.GetMultiple(identifiers []string) (map[string]*NAV, []error)` - Batch fetch

**NAV Structure:**

```go
type NAV struct {
    SchemeCode string
    ISIN       string  // Primary ISIN
    ISINReinv  string  // Reinvestment ISIN
    SchemeName string
    NAV        float64
    Date       string
}
```

## Logging

The library provides **optional** structured logging using Go's standard `log/slog` package.

**By default, logging is disabled** (silent). This follows Go library best practices - the library won't pollute your logs unless you explicitly enable it.

### Enable Logging

```go
import "github.com/Vikramarjuna/findata-go/logger"

// Enable with default settings (INFO level, text format, stderr)
logger.SetLogger(logger.NewSlogLogger())

// Enable with custom configuration
logger.SetLogger(logger.NewSlogLogger(
    logger.WithLevel(logger.LevelDebug),      // DEBUG, INFO, WARN, ERROR
    logger.WithFormat(logger.FormatJSON),     // FormatText or FormatJSON
    logger.WithOutput(logFile),               // Any io.Writer
))

// Disable logging
logger.SetLogger(nil)
```

### Example Output

```text
time=2026-01-06T21:54:51.887+05:30 level=INFO msg="successfully fetched NSE quote" symbol=RELIANCE price=1577.00 currency=INR
```

### Use Your Own Logger

Implement the `logger.Logger` interface to use your own logging framework:

```go
type Logger interface {
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
}

logger.SetLogger(myCustomLogger)
```

See the [logger package documentation](https://pkg.go.dev/github.com/Vikramarjuna/findata-go/logger) for full API details.

## Examples

See the [examples](./examples) directory for complete working examples.

## Versioning

This project follows [Semantic Versioning](https://semver.org/). See [CHANGELOG.md](./CHANGELOG.md) for release history.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for details on:

- How to report bugs
- How to suggest features
- Development setup
- Coding standards
- Pull request process

## License

MIT License - see [LICENSE](./LICENSE) file for details.

## Acknowledgments

- Inspired by [finance-go](https://github.com/piquette/finance-go)
- Data sources: NSE India, AMFI India
