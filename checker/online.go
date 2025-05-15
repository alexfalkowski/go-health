package checker

import (
	"time"
)

var _ Checker = (*OnlineChecker)(nil)

// NewOnlineChecker checks https://google.com/generate_204.
func NewOnlineChecker(t time.Duration, opts ...Option) *OnlineChecker {
	return &OnlineChecker{NewHTTPChecker("https://google.com/generate_204", t, opts...)}
}

// OnlineChecker is just HTTPChecker with a url.
type OnlineChecker struct {
	*HTTPChecker
}
