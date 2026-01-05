// Package finance provides access to Indian financial markets data
// including NSE stocks and AMFI mutual funds.
package finance

import (
	"net/http"
	"time"
)

const (
	// Version is the current version of the library
	Version = "0.1.0"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultUserAgent is the default User-Agent header
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Client is the HTTP client used for all requests
var Client = &http.Client{
	Timeout: DefaultTimeout,
}

// SetHTTPClient allows users to set a custom HTTP client
func SetHTTPClient(client *http.Client) {
	Client = client
}

// Error represents an error from the finance API
type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	if e.Code > 0 {
		return e.Message + " (HTTP " + string(rune(e.Code)) + ")"
	}
	return e.Message
}

