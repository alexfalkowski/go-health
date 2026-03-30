package probe_test

import (
	"context"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/stretchr/testify/require"
)

func TestStopWithoutStartDoesNotPanic(t *testing.T) {
	p := probe.NewProbe("noop", 10*time.Millisecond, checker.NewNoopChecker())

	require.NotPanics(t, func() {
		p.Stop()
	})
}

func TestStopCancelsInFlightCheck(t *testing.T) {
	ch := &blockingChecker{
		started:  make(chan struct{}),
		canceled: make(chan struct{}),
	}
	p := probe.NewProbe("blocking", time.Hour, ch)

	started := make(chan startResult, 1)
	// Start blocks until the initial check finishes, so run it in the background.
	go func() {
		started <- startResult{ticks: p.Start()}
	}()

	select {
	case <-started:
		require.Fail(t, "start returned before the initial check was canceled")
	case <-ch.started:
	case <-time.After(time.Second):
		require.Fail(t, "checker did not start")
	}

	stopped := make(chan struct{})
	// Stop should cancel the in-flight check rather than waiting forever.
	go func() {
		p.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		require.Fail(t, "stop did not cancel the in-flight check")
	}

	select {
	case <-ch.canceled:
	case <-time.After(time.Second):
		require.Fail(t, "checker did not observe cancellation")
	}

	var result startResult
	// Once the initial check is canceled, Start should return and the probe channel should close.
	select {
	case result = <-started:
	case <-time.After(time.Second):
		require.Fail(t, "start did not return after cancellation")
	}

	_, ok := <-result.ticks
	require.False(t, ok)
}

func TestStartWithInvalidPeriodReturnsErrorTick(t *testing.T) {
	p := probe.NewProbe("noop", 0, checker.NewNoopChecker())

	require.NotPanics(t, func() {
		ticks := p.Start()

		tick, ok := <-ticks
		require.True(t, ok)
		require.Equal(t, "noop", tick.Name())
		require.ErrorIs(t, tick.Error(), probe.ErrInvalidPeriod)
		require.ErrorContains(t, tick.Error(), "0s")

		_, ok = <-ticks
		require.False(t, ok)
	})
}

type blockingChecker struct {
	started  chan struct{}
	canceled chan struct{}
}

type startResult struct {
	ticks <-chan *probe.Tick
}

func (c *blockingChecker) Check(ctx context.Context) error {
	close(c.started)
	<-ctx.Done()
	close(c.canceled)
	return ctx.Err()
}
