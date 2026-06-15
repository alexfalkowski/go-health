package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var _ Checker = (*OnlineChecker)(nil)

// ErrNotOnline is wrapped when none of the configured connectivity URLs appear healthy.
var ErrNotOnline = errors.New("not online")

// NewOnlineChecker returns an OnlineChecker that checks whether any configured URL
// is reachable.
//
// Passing 0 uses the package default timeout of 30 seconds. It uses the default
// URL list unless overridden via WithURLs.
func NewOnlineChecker(t time.Duration, opts ...Option) *OnlineChecker {
	options := parseOptions(opts...)

	return &OnlineChecker{
		urls:   options.urls,
		client: &http.Client{Transport: options.roundTripper, Timeout: timeout(t)},
	}
}

// OnlineChecker checks a list of URLs concurrently and wraps ErrNotOnline if
// none of them respond with 200 OK or 204 No Content.
//
// The intent is to answer "can this process reach the outside world?" rather than
// "is one specific upstream healthy?" Any single successful response is enough for
// the checker to report healthy.
type OnlineChecker struct {
	client *http.Client
	urls   []string
}

// Check performs the online check.
//
// Requests are issued concurrently with the supplied context. Individual request
// errors are ignored unless every configured URL fails or returns an unexpected
// status code. The checker cancels remaining requests after the first healthy
// response.
func (c *OnlineChecker) Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	checkCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan bool, len(c.urls))
	for _, url := range c.urls {
		go func() {
			healthy := c.checkURL(checkCtx, url)
			if healthy {
				cancel()
			}

			results <- healthy
		}()
	}

	for range c.urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case healthy := <-results:
			if healthy {
				return nil
			}
		}
	}

	return fmt.Errorf("online checker: %w", ErrNotOnline)
}

func (c *OnlineChecker) checkURL(ctx context.Context, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent
}
