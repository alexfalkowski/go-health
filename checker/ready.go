package checker

import (
	"context"
	"sync/atomic"
)

var _ Checker = (*ReadyChecker)(nil)

// NewReadyChecker returns a ReadyChecker.
//
// Until Ready is called, Check returns err.
func NewReadyChecker(err error) *ReadyChecker {
	return &ReadyChecker{err: err}
}

// ReadyChecker reports a fixed error until it is marked ready.
type ReadyChecker struct {
	err  error
	flag int32
}

// Check returns an error until Ready is called.
func (c *ReadyChecker) Check(_ context.Context) error {
	if c.notReady() {
		return c.err
	}

	return nil
}

// Ready marks the checker as ready.
func (c *ReadyChecker) Ready() {
	atomic.StoreInt32(&(c.flag), 1)
}

func (c *ReadyChecker) notReady() bool {
	return atomic.LoadInt32(&(c.flag)) == 0
}
