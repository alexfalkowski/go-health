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
func NewHTTPChecker(url string, timeout time.Duration) *HTTPChecker {
	return NewHTTPCheckerWithClient(url, timeout, nil)
}

// NewHTTPCheckerWithClient with URL and client.
func NewHTTPCheckerWithClient(url string, timeout time.Duration, client *http.Client) *HTTPChecker {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPChecker{url: url, timeout: timeout, client: client}
}

// HTTPChecker for a URL.
type HTTPChecker struct {
	url     string
	timeout time.Duration
	client  *http.Client
}

// Check the URL with a GET.
func (c *HTTPChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
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
