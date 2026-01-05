# finance-india

[![GoDoc](https://godoc.org/github.com/Vikramarjuna/finance-india?status.svg)](https://godoc.org/github.com/Vikramarjuna/finance-india)
[![Go Report Card](https://goreportcard.com/badge/github.com/Vikramarjuna/finance-india)](https://goreportcard.com/report/github.com/Vikramarjuna/finance-india)

A Go library for accessing Indian financial markets data, including NSE stocks and AMFI mutual funds.

Inspired by [finance-go](https://github.com/piquette/finance-go), this library provides a clean, idiomatic Go interface to Indian financial data sources.

## Features

| Description | Source |
|-------------|--------|
| NSE Stock Quotes | NSE India API |
| Mutual Fund NAVs | AMFI Portal |
| Batch Quote Fetching | NSE India API |
| Mutual Fund Search | AMFI Portal |

## Installation

```bash
go get github.com/Vikramarjuna/finance-india
```

## Usage

### NSE Stock Quotes

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Vikramarjuna/finance-india/nse"
)

func main() {
    // Get a single quote
    quote, err := nse.Get("RELIANCE")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("%s: ₹%.2f (%.2f%%)\n", 
        quote.Symbol, quote.LastPrice, quote.PChange)
    fmt.Printf("Sector: %s | Industry: %s\n", 
        quote.Sector, quote.Industry)
    
    // Get multiple quotes
    symbols := []string{"TCS", "INFY", "WIPRO"}
    quotes, errors := nse.GetMultiple(symbols)
    
    for symbol, q := range quotes {
        fmt.Printf("%s: ₹%.2f\n", symbol, q.LastPrice)
    }
}
```

### Mutual Fund NAVs

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Vikramarjuna/finance-india/mf"
)

func main() {
    // Search for funds
    results, err := mf.Search("HDFC Flexi Cap")
    if err != nil {
        log.Fatal(err)
    }
    
    for _, nav := range results {
        fmt.Printf("%s: ₹%.4f\n", nav.SchemeName, nav.NAV)
    }
    
    // Get by ISIN or scheme code
    nav, err := mf.Get("INF179K01997")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("NAV: ₹%.4f (as of %s)\n", nav.NAV, nav.Date)
}
```

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
    
    "github.com/Vikramarjuna/finance-india"
)

func main() {
    // Set custom timeout
    client := &http.Client{
        Timeout: 60 * time.Second,
    }
    finance.SetHTTPClient(client)
    
    // Now all requests will use this client
}
```

## API Reference

### NSE Package

#### `nse.Get(symbol string) (*Quote, error)`
Fetches a quote for the given NSE symbol.

#### `nse.GetMultiple(symbols []string) (map[string]*Quote, []error)`
Fetches quotes for multiple symbols. Returns a map of symbol -> quote and any errors encountered.

#### Quote Structure
```go
type Quote struct {
    Symbol            string
    CompanyName       string
    Industry          string
    Sector            string
    LastPrice         float64
    Change            float64
    PChange           float64
    PreviousClose     float64
    Open              float64
    DayHigh           float64
    DayLow            float64
    YearHigh          float64
    YearLow           float64
    TotalTradedVolume float64
    TotalTradedValue  float64
    Indices           []string
}
```

### MF Package

#### `mf.GetAll() (map[string]*NAV, error)`
Fetches all NAVs from AMFI. Returns a map indexed by scheme code, name, and ISINs.

#### `mf.Get(identifier string) (*NAV, error)`
Fetches NAV for a specific scheme by code, name, or ISIN.

#### `mf.Search(query string) ([]*NAV, error)`
Searches for mutual funds by name (case-insensitive partial match).

#### `mf.GetMultiple(identifiers []string) (map[string]*NAV, []error)`
Fetches NAVs for multiple identifiers.

#### NAV Structure
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

## Examples

See the [examples](./examples) directory for complete working examples.

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

