package checker

import (
	"context"
	"testing"
	"time"

	"github.com/alexfalkowski/go-sync"
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

// BlockingPeriodicChecker returns once, then blocks future checks until Release is closed.
type BlockingPeriodicChecker struct {
	InitialStarted  chan struct{}
	PeriodicStarted chan struct{}
	Release         chan struct{}
	periodicOnce    sync.Once
}

// CancelablePeriodicChecker returns once, then blocks future checks until the context is canceled.
type CancelablePeriodicChecker struct {
	InitialStarted   chan struct{}
	PeriodicStarted  chan struct{}
	PeriodicCanceled chan struct{}
	canceledOnce     sync.Once
	periodicOnce     sync.Once
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

// NewBlockingPeriodicChecker returns a BlockingPeriodicChecker with initialized channels.
func NewBlockingPeriodicChecker() *BlockingPeriodicChecker {
	return &BlockingPeriodicChecker{
		InitialStarted:  make(chan struct{}),
		PeriodicStarted: make(chan struct{}),
		Release:         make(chan struct{}),
	}
}

// NewCancelablePeriodicChecker returns a CancelablePeriodicChecker with initialized channels.
func NewCancelablePeriodicChecker() *CancelablePeriodicChecker {
	return &CancelablePeriodicChecker{
		InitialStarted:   make(chan struct{}),
		PeriodicStarted:  make(chan struct{}),
		PeriodicCanceled: make(chan struct{}),
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

// Check returns immediately the first time, then blocks until Release is closed.
func (c *BlockingPeriodicChecker) Check(context.Context) error {
	select {
	case <-c.InitialStarted:
		c.periodicOnce.Do(func() {
			close(c.PeriodicStarted)
		})
		<-c.Release
	default:
		close(c.InitialStarted)
	}

	return nil
}

// Check returns immediately the first time, then waits for context cancellation.
func (c *CancelablePeriodicChecker) Check(ctx context.Context) error {
	select {
	case <-c.InitialStarted:
		c.periodicOnce.Do(func() {
			close(c.PeriodicStarted)
		})
		<-ctx.Done()
		c.canceledOnce.Do(func() {
			close(c.PeriodicCanceled)
		})
		return ctx.Err()
	default:
		close(c.InitialStarted)
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
