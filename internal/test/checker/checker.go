package checker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// BlockingChecker blocks until the context is canceled.
type BlockingChecker struct {
	Started  chan struct{}
	Canceled chan struct{}
}

// ReleasableChecker blocks until Release is closed or the context is canceled.
type ReleasableChecker struct {
	Started chan struct{}
	Release chan struct{}
	once    sync.Once
}

// ReleasableStartChecker blocks until Release is closed or the context is canceled.
type ReleasableStartChecker struct {
	Started chan struct{}
	Release <-chan struct{}
}

// NewBlockingChecker returns a BlockingChecker with initialized channels.
func NewBlockingChecker() *BlockingChecker {
	return &BlockingChecker{
		Started:  make(chan struct{}),
		Canceled: make(chan struct{}),
	}
}

// NewReleasableChecker returns a ReleasableChecker with initialized channels.
func NewReleasableChecker() *ReleasableChecker {
	return &ReleasableChecker{
		Started: make(chan struct{}),
		Release: make(chan struct{}),
	}
}

// NewReleasableStartChecker returns a ReleasableStartChecker with initialized channels.
func NewReleasableStartChecker(release <-chan struct{}) *ReleasableStartChecker {
	return &ReleasableStartChecker{
		Started: make(chan struct{}),
		Release: release,
	}
}

// Check reports that it started, then waits for context cancellation.
func (c *BlockingChecker) Check(ctx context.Context) error {
	close(c.Started)
	<-ctx.Done()
	close(c.Canceled)
	return ctx.Err()
}

// Check reports that it started, then waits for release or cancellation.
func (c *ReleasableChecker) Check(ctx context.Context) error {
	c.once.Do(func() {
		close(c.Started)
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.Release:
		return nil
	}
}

// Check reports that it started, then waits for release or cancellation.
func (c *ReleasableStartChecker) Check(ctx context.Context) error {
	close(c.Started)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.Release:
		return nil
	}
}

// WaitForStarted waits until started is closed.
func WaitForStarted(t *testing.T, started <-chan struct{}) {
	t.Helper()

	select {
	case <-started:
	case <-time.After(time.Second):
		require.Fail(t, "checker did not start")
	}
}

// WaitForCanceled waits until canceled is closed.
func WaitForCanceled(t *testing.T, canceled <-chan struct{}) {
	t.Helper()

	select {
	case <-canceled:
	case <-time.After(time.Second):
		require.Fail(t, "checker did not observe cancellation")
	}
}
