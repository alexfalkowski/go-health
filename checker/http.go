package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var _ Checker = (*HTTPChecker)(nil)

// ErrInvalidStatusCode is returned when the response status code is in the 4xx or 5xx range.
var ErrInvalidStatusCode = errors.New("invalid status code")

// NewHTTPChecker returns an HTTPChecker that performs a GET request to url.
//
// The timeout is applied to the underlying http.Client.
func NewHTTPChecker(url string, t time.Duration, opts ...Option) *HTTPChecker {
	os := parseOptions(opts...)

	return &HTTPChecker{
		url:    url,
		client: &http.Client{Transport: os.roundTripper, Timeout: timeout(t)},
	}
}

// HTTPChecker performs an HTTP GET request to a URL.
type HTTPChecker struct {
	client *http.Client
	url    string
}

// Check performs an HTTP GET request.
func (c *HTTPChecker) Check(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, http.NoBody)
	if err != nil {
		return fmt.Errorf("http checker: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("http checker: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 && response.StatusCode <= 599 {
		return fmt.Errorf("http checker: %w", ErrInvalidStatusCode)
	}

	return nil
}
