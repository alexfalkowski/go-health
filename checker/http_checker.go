package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrInvalidStatusCode when we have errors that are not in the 200 range.
var ErrInvalidStatusCode = errors.New("invalid status code")

// NewHTTPChecker with URL and client.
func NewHTTPChecker(url string, rt http.RoundTripper, t time.Duration) *HTTPChecker {
	return &HTTPChecker{url: url, client: &http.Client{Transport: rt, Timeout: timeout(t)}}
}

// HTTPChecker for a URL.
type HTTPChecker struct {
	client *http.Client
	url    string
}

// Check the URL with a GET.
func (c *HTTPChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url, http.NoBody)
	if err != nil {
		return fmt.Errorf("http checker: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("http checker: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidStatusCode
	}

	return nil
}
