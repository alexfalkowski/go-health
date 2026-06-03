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

// NewHTTPChecker returns an HTTPChecker that performs an HTTP GET request to url.
//
// The timeout is applied to the underlying http.Client. Passing 0 uses the
// package default of 30 seconds. Use WithRoundTripper to provide a custom
// transport for tests or bespoke network behavior.
func NewHTTPChecker(url string, t time.Duration, opts ...Option) *HTTPChecker {
	options := parseOptions(opts...)

	return &HTTPChecker{
		url:    url,
		client: &http.Client{Transport: options.roundTripper, Timeout: timeout(t)},
	}
}

// HTTPChecker performs an HTTP GET request to a URL.
//
// Responses with status codes below 400 are considered healthy. Responses with
// status codes in the 4xx or 5xx range return ErrInvalidStatusCode wrapped with
// context.
type HTTPChecker struct {
	client *http.Client
	url    string
}

// Check performs the HTTP GET request with the supplied context.
func (c *HTTPChecker) Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, http.NoBody)
	if err != nil {
		return fmt.Errorf("http checker: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return fmt.Errorf("http checker: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("http checker: %w", ErrInvalidStatusCode)
	}

	return nil
}
