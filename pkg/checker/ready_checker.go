package checker

import (
	"context"
	"sync/atomic"
)

// NewReadyChecker with a specific error.
func NewReadyChecker(err error) *ReadyChecker {
	return &ReadyChecker{err: err}
}

// ReadyChecker for when prepared for something.
type ReadyChecker struct {
	flag int32
	err  error
}

// Check the if ready.
func (c *ReadyChecker) Check(ctx context.Context) error {
	if c.notReady() {
		return c.err
	}

	return nil
}

// Ready marks the checker as done.
func (c *ReadyChecker) Ready() {
	atomic.StoreInt32(&(c.flag), 1)
}

func (c *ReadyChecker) notReady() bool {
	return atomic.LoadInt32(&(c.flag)) == 0
}
