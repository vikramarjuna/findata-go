# Migration Guide: Using findata-go in investment_tracker

This guide shows how to replace the existing price sync code in `investment_tracker` with the new `findata-go` library.

## Installation

Add the library to your `go.mod`:

```bash
cd investment_tracker
go get github.com/Vikramarjuna/findata-go
```

## Changes Required

### 1. Update `service/pricesync.go`

Replace the NSE quote fetching code:

**Before:**
```go
func (s *PriceSyncService) FetchNSEQuote(symbol string) (*NSEQuoteResponse, error) {
    url := fmt.Sprintf("https://www.nseindia.com/api/quote-equity?symbol=%s", symbol)
    // ... HTTP request code ...
}
```

**After:**

```go
import "github.com/Vikramarjuna/findata-go/equity"

func (s *PriceSyncService) FetchNSEQuote(symbol string) (*equity.Quote, error) {
    return equity.Get(symbol)  // Auto-detects NSE!
}
```

### 2. Update AMFI NAV Fetching

**Before:**

```go
func (s *PriceSyncService) FetchAMFINAVs() (map[string]AMFINav, error) {
    url := "https://portal.amfiindia.com/spages/NAVAll.txt"
    // ... parsing code ...
}
```

**After:**

```go
import "github.com/Vikramarjuna/findata-go/mf"

func (s *PriceSyncService) FetchAMFINAVs() (map[string]*mf.NAV, error) {
    return mf.GetAll()
}
```

### 3. Update Type Mappings

The library uses cleaner type names. Update your code:

| Old Type | New Type |
|----------|----------|
| `NSEQuoteResponse` | `equity.Quote` |
| `AMFINav` | `mf.NAV` |

### 4. Update Field Access

Some fields have been renamed for clarity:

**Equity Quote Fields:**

```go
// Old
quote.PriceInfo.LastPrice
quote.Info.CompanyName
quote.IndustryInfo.Sector

// New (cleaner!)
quote.LastPrice
quote.CompanyName
quote.Sector
quote.Exchange  // NEW: Shows which exchange (NSE, BSE, etc.)
quote.Currency  // NEW: Shows currency (INR, USD, etc.)
```

**AMFI NAV Fields:**
```go
// Old
nav.SchemeName
nav.NAV

// New (same)
nav.SchemeName
nav.NAV
```

## Benefits

1. **Cleaner Code**: Remove ~300 lines of HTTP/parsing code
2. **Reusable**: Can be used in other projects
3. **Tested**: Comes with comprehensive tests
4. **Maintained**: Separate library with its own versioning
5. **Type-Safe**: Better type definitions and error handling

## Example: Complete Refactor

Here's how the sync function would look:

```go
import (
    "github.com/Vikramarjuna/findata-go/config"
    "github.com/Vikramarjuna/findata-go/equity"
)

func init() {
    // Set default market once at startup
    config.SetDefaultMarket(config.MarketIndia)
}

func (s *PriceSyncService) syncEquities(equities []repository.Holding, result *SyncResult) {
    // Get unique symbols
    symbols := make([]string, 0, len(equities))
    for _, h := range equities {
        symbols = append(symbols, h.Symbol)
    }

    // Fetch all quotes at once - library auto-detects exchanges!
    quotes, errors := equity.GetMultiple(symbols)

    // Log errors
    for _, err := range errors {
        logger.Error("Failed to fetch quote: %v", err)
        result.Errors = append(result.Errors, err.Error())
    }

    // Update holdings
    for symbol, quote := range quotes {
        for _, h := range equities {
            if h.Symbol == symbol {
                tags := map[string]string{
                    "name":     quote.CompanyName,
                    "sector":   quote.Sector,
                    "industry": quote.Industry,
                    "exchange": string(quote.Exchange),  // NEW: Track exchange
                    "currency": quote.Currency,          // NEW: Track currency
                }
                if err := s.repo.UpdateHoldingPriceAndTags(h.Symbol, quote.LastPrice, tags); err != nil {
                    logger.Error("Failed to update %s: %v", h.Symbol, err)
                    continue
                }
                result.EquitiesSynced++
            }
        }
    }
}
```

## Testing

After migration, run tests to ensure everything works:

```bash
cd investment_tracker
make test
```

## Rollback Plan

If issues arise, the old code is preserved in git history. You can:

1. Revert the changes: `git revert <commit>`
2. Or keep both implementations and use a feature flag

## Next Steps

1. Update `go.mod` to include `finance-india`
2. Refactor `service/pricesync.go`
3. Update tests
4. Run integration tests
5. Deploy and monitor

