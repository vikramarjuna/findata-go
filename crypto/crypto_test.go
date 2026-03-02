package crypto

import (
	"testing"
)

func TestWithCurrency(t *testing.T) {
	opts := &Options{}
	WithCurrency("INR")(opts)

	if opts.Currency != "INR" {
		t.Errorf("Expected currency to be INR, got %s", opts.Currency)
	}
}

func TestCryptoError(t *testing.T) {
	err := &Error{
		Message:  "test error",
		CoinID:   "bitcoin",
		Currency: "USD",
		Provider: "TestProvider",
	}

	expected := "crypto: bitcoin (USD): test error (provider: TestProvider)"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestCryptoErrorWithoutSymbol(t *testing.T) {
	err := &Error{
		Message:  "test error",
		Provider: "TestProvider",
	}

	expected := "crypto: test error (provider: TestProvider)"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestNewFetcher(t *testing.T) {
	fetcher := NewFetcher()
	if fetcher == nil {
		t.Error("NewFetcher should not return nil")
	}

	_, ok := fetcher.(*DefaultFetcher)
	if !ok {
		t.Error("NewFetcher should return a DefaultFetcher")
	}
}
