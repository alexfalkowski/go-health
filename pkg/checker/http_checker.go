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
	return NewHTTPCheckerWithRoundTripper(url, timeout, nil)
}

// NewHTTPCheckerWithRoundTripper with URL and client.
func NewHTTPCheckerWithRoundTripper(url string, timeout time.Duration, tripper http.RoundTripper) *HTTPChecker {
	if tripper == nil {
		tripper = http.DefaultTransport
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: tripper,
		Jar:       nil,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse // never follow redirects
		},
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
		return fmt.Errorf("http new request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("client do: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidStatusCode
	}

	return nil
}
