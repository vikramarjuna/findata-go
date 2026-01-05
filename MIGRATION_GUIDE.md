# Migration Guide: Using finance-india in investment_tracker

This guide shows how to replace the existing price sync code in `investment_tracker` with the new `finance-india` library.

## Installation

Add the library to your `go.mod`:

```bash
cd investment_tracker
go get github.com/Vikramarjuna/finance-india
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
import "github.com/Vikramarjuna/finance-india/nse"

func (s *PriceSyncService) FetchNSEQuote(symbol string) (*nse.Quote, error) {
    return nse.Get(symbol)
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
import "github.com/Vikramarjuna/finance-india/mf"

func (s *PriceSyncService) FetchAMFINAVs() (map[string]*mf.NAV, error) {
    return mf.GetAll()
}
```

### 3. Update Type Mappings

The library uses cleaner type names. Update your code:

| Old Type | New Type |
|----------|----------|
| `NSEQuoteResponse` | `nse.Quote` |
| `AMFINav` | `mf.NAV` |

### 4. Update Field Access

Some fields have been renamed for clarity:

**NSE Quote Fields:**
```go
// Old
quote.PriceInfo.LastPrice
quote.Info.CompanyName
quote.IndustryInfo.Sector

// New
quote.LastPrice
quote.CompanyName
quote.Sector
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
func (s *PriceSyncService) syncEquities(equities []repository.Holding, result *SyncResult) {
    // Get unique symbols
    symbols := make([]string, 0, len(equities))
    for _, h := range equities {
        symbols = append(symbols, h.Symbol)
    }
    
    // Fetch all quotes at once
    quotes, errors := nse.GetMultiple(symbols)
    
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

