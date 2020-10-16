package chk

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// ErrInvalidStatusCode when we have errors that are not in the 200 range.
var ErrInvalidStatusCode = errors.New("invalid status code")

// NewHTTPChecker with URL and client.
func NewHTTPChecker(url string, timeout time.Duration, client *http.Client) Checker {
	if client == nil {
		client = http.DefaultClient
	}

	return &httpChecker{url, timeout, client}
}

type httpChecker struct {
	url     string
	timeout time.Duration
	client  *http.Client
}

func (c *httpChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !c.isValidStatusCode(resp.StatusCode) {
		return ErrInvalidStatusCode
	}

	return ctx.Err()
}

func (c *httpChecker) isValidStatusCode(sc int) bool {
	return sc >= 200 && sc <= 299
}
