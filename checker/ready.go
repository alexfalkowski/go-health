package checker

import (
	"context"

	"github.com/alexfalkowski/go-sync"
)

var _ Checker = (*ReadyChecker)(nil)

// NewReadyChecker returns a ReadyChecker.
//
// Until Ready is called, Check returns err. Callers typically pass a stable
// sentinel error such as errors.New("not ready") so they can detect the reason
// with errors.Is.
func NewReadyChecker(err error) *ReadyChecker {
	return &ReadyChecker{err: err}
}

// ReadyChecker reports a fixed error until it is marked ready.
//
// Ready is a one-way transition; once marked ready the checker stays healthy for
// the lifetime of the value.
type ReadyChecker struct {
	err  error
	flag sync.Int32
}

// Check returns the configured error until Ready is called.
func (c *ReadyChecker) Check(_ context.Context) error {
	if c.notReady() {
		return c.err
	}

	return nil
}

// Ready marks the checker as ready.
//
// Ready is safe to call multiple times.
func (c *ReadyChecker) Ready() {
	c.flag.Store(1)
}

func (c *ReadyChecker) notReady() bool {
	return c.flag.Load() == 0
}
