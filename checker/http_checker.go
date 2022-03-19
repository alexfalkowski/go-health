package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// ErrInvalidStatusCode when we have errors that are not in the 200 range.
var ErrInvalidStatusCode = errors.New("invalid status code")

// NewHTTPChecker with URL and client.
func NewHTTPChecker(url string, client *http.Client) *HTTPChecker {
	return &HTTPChecker{url: url, client: client}
}

// HTTPChecker for a URL.
type HTTPChecker struct {
	url    string
	client *http.Client
}

// Check the URL with a GET.
func (c *HTTPChecker) Check(ctx context.Context) error {
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
