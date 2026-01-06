# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of findata-go library
- NSE equity quote fetching with automatic provider detection
- AMFI mutual fund NAV fetching with fuzzy search
- Built-in caching with TTL support
- Market cap detection (LargeCap/MidCap/SmallCap)
- Optional structured logging with slog
- Comprehensive test coverage
- Examples for NSE and MF usage

### Features

#### Equity Package
- `equity.Get(symbol)` - Fetch NSE quote for a symbol
- Automatic provider detection (NSE India, NSE Gujrat)
- Market cap classification based on indices
- Configurable caching

#### Mutual Fund Package
- `mf.Get(isin)` - Fetch NAV by ISIN
- `mf.Search(query)` - Fuzzy search for mutual funds
- Support for scheme codes and names
- Normalization and matching algorithms

#### Configuration
- `config.SetDefaultMarket()` - Set default market
- `config.SetDefaultExchange()` - Set default exchange
- Functional options pattern

#### Logging
- Optional opt-in logging (silent by default)
- Interface-based design for custom loggers
- Built-in slog adapter
- Multiple log levels (DEBUG, INFO, WARN, ERROR)
- JSON and text formats

## [0.1.0] - TBD

Initial release.

[Unreleased]: https://github.com/Vikramarjuna/findata-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Vikramarjuna/findata-go/releases/tag/v0.1.0

