package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alexfalkowski/go-sync"
)

var _ Checker = (*OnlineChecker)(nil)

// ErrNotOnline is returned when none of the configured connectivity URLs appear healthy.
var ErrNotOnline = errors.New("not online")

// NewOnlineChecker returns an OnlineChecker that checks whether any configured URL
// is reachable.
//
// Passing 0 uses the package default timeout of 30 seconds. It uses the default
// URL list unless overridden via WithURLs.
func NewOnlineChecker(t time.Duration, opts ...Option) *OnlineChecker {
	os := parseOptions(opts...)

	return &OnlineChecker{
		urls:   os.urls,
		client: &http.Client{Transport: os.roundTripper, Timeout: timeout(t)},
	}
}

// OnlineChecker checks a list of URLs concurrently and returns ErrNotOnline if none
// of them respond with 200 OK or 204 No Content.
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
// status code.
func (c *OnlineChecker) Check(ctx context.Context) error {
	var counter sync.Int32
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
