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

// NewOnlineChecker checks https://antonz.org/is-online/.
func NewOnlineChecker(t time.Duration, opts ...Option) *OnlineChecker {
	os := parseOptions(opts...)

	return &OnlineChecker{
		urls:   os.urls,
		client: &http.Client{Transport: os.roundTripper, Timeout: timeout(t)},
	}
}

// OnlineChecker will verify all the urls and if all are not online, it will return ErrNotOnline.
type OnlineChecker struct {
	client *http.Client
	urls   []string
}

// Check the all the urls.
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
