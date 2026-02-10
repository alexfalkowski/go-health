package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var _ Checker = (*OnlineChecker)(nil)

// ErrNotOnline when the system is not online.
var ErrNotOnline = errors.New("not online")

// NewOnlineChecker returns an OnlineChecker that checks whether any configured URL
// is reachable.
//
// It uses the default URL list unless overridden via WithURLs.
func NewOnlineChecker(t time.Duration, opts ...Option) *OnlineChecker {
	os := parseOptions(opts...)

	return &OnlineChecker{
		urls:   os.urls,
		client: &http.Client{Transport: os.roundTripper, Timeout: timeout(t)},
	}
}

// OnlineChecker checks a list of URLs concurrently and returns ErrNotOnline if none
// of them respond with 200 OK or 204 No Content.
type OnlineChecker struct {
	client *http.Client
	urls   []string
}

// Check performs the online check.
func (c *OnlineChecker) Check(ctx context.Context) error {
	var counter atomic.Uint64
	var wg sync.WaitGroup

	for _, url := range c.urls {
		wg.Go(func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
			if err != nil {
				return
			}

			resp, err := c.client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				counter.Add(1)
			}
		})
	}

	wg.Wait()

	if counter.Load() == 0 {
		return fmt.Errorf("online checker: %w", ErrNotOnline)
	}
	return nil
}
