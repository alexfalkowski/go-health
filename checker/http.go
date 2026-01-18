package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var _ Checker = (*HTTPChecker)(nil)

// ErrInvalidStatusCode when we have errors that are not in the 200-300 range.
var ErrInvalidStatusCode = errors.New("invalid status code")

// NewHTTPChecker with url, round tripper and timeout.
func NewHTTPChecker(url string, t time.Duration, opts ...Option) *HTTPChecker {
	os := parseOptions(opts...)

	return &HTTPChecker{
		url:    url,
		client: &http.Client{Transport: os.roundTripper, Timeout: timeout(t)},
	}
}

// HTTPChecker with url, round tripper and timeout..
type HTTPChecker struct {
	client *http.Client
	url    string
}

// Check the URL with a GET.
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
