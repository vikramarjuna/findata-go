# Testing Guide

This document describes how to run tests in the findata-go project.

## Quick Start

```bash
# Run all tests with race detector and parallelism
make test

# Run tests without integration tests
make test-short

# Run tests with higher parallelism (8 workers)
make test-parallel

# Run tests for a specific package
make test-pkg PKG=./provider/nse

# Generate coverage report
make coverage
```

## Parallel Test Execution

### What's Already Parallel

✅ **Package-level parallelism**: Go automatically runs tests from different packages in parallel
✅ **CI/CD parallelism**: GitHub Actions runs tests across 3 Go versions (1.22, 1.23, 1.24) simultaneously
✅ **Race detector**: All tests run with `-race` flag to detect data races
✅ **Configurable workers**: Tests use `-parallel 4` flag (4 concurrent tests per package)

### Test Execution Flags

| Flag | Purpose | Default |
|------|---------|---------|
| `-race` | Enable race detector | Always on |
| `-parallel N` | Max concurrent tests per package | 4 |
| `-timeout` | Test timeout | 10m (CI only) |
| `-short` | Skip long-running tests | Off |
| `-v` | Verbose output | On |

### Performance

Current test execution time:
- **Local**: ~16-22 seconds (with race detector)
- **CI**: ~30-40 seconds per Go version (parallel across versions)

## Adding Parallel Tests

To make individual tests run in parallel, add `t.Parallel()`:

```go
func TestSomething(t *testing.T) {
    t.Parallel()  // This test can run in parallel with other parallel tests
    
    // Your test code here
}
```

**When to use `t.Parallel()`:**
- ✅ Tests that don't share state
- ✅ Tests that don't modify global variables
- ✅ Tests that don't use the same external resources

**When NOT to use `t.Parallel()`:**
- ❌ Tests that modify shared state
- ❌ Tests that use `t.Setenv()` (incompatible with parallel)
- ❌ Tests that depend on execution order
- ❌ Tests with `time.Sleep()` (can cause flakiness)

## CI/CD Parallel Execution

GitHub Actions runs tests in parallel:

```yaml
strategy:
  fail-fast: false  # Don't cancel other jobs if one fails
  matrix:
    go-version: ['1.22', '1.23', '1.24']
```

This creates 3 parallel jobs, one for each Go version.

## Advanced Usage

### Run tests with custom parallelism

```bash
# Run with 8 parallel workers
go test -v -race -parallel 8 ./...

# Run with 1 worker (sequential)
go test -v -race -parallel 1 ./...
```

### Run specific tests

```bash
# Run a specific test
go test -v -race -run TestCache_GetSet ./cache

# Run tests matching a pattern
go test -v -race -run "TestCache_.*" ./cache
```

### Benchmark tests

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Run benchmarks for a specific package
go test -bench=. -benchmem ./provider/nse
```

### Coverage

```bash
# Generate coverage report
make coverage

# View coverage in terminal
go test -cover ./...

# View coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Troubleshooting

### Race detector failures

If you see race detector warnings:
```
WARNING: DATA RACE
```

This means concurrent access to shared memory. Fix by:
1. Using mutexes (`sync.Mutex`)
2. Using channels
3. Avoiding shared state

### Timeout errors

If tests timeout:
```bash
# Increase timeout
go test -timeout 30m ./...
```

### Flaky tests

If tests fail intermittently:
1. Check for race conditions (run with `-race`)
2. Avoid `time.Sleep()` - use channels or `time.After()`
3. Don't rely on execution order
4. Make tests deterministic

## Best Practices

1. ✅ **Always run with race detector**: `go test -race`
2. ✅ **Use table-driven tests**: Makes it easy to add test cases
3. ✅ **Use subtests**: `t.Run()` for better organization
4. ✅ **Mock external dependencies**: Don't rely on network/filesystem
5. ✅ **Keep tests fast**: Aim for < 1 second per test
6. ✅ **Test edge cases**: Empty inputs, nil values, errors
7. ✅ **Use meaningful test names**: `TestCache_GetSet_WithExpiredKey`

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)

