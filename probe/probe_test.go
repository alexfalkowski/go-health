package probe_test

import (
	"context"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	testchecker "github.com/alexfalkowski/go-health/v2/internal/test/checker"
	testprobe "github.com/alexfalkowski/go-health/v2/internal/test/probe"
	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/stretchr/testify/require"
)

func TestStopWithoutStartDoesNotPanic(t *testing.T) {
	p := probe.NewProbe("noop", 10*time.Millisecond, checker.NewNoopChecker())

	require.NotPanics(t, func() {
		_ = p.Stop(t.Context())
	})
}

func TestStopCancelsInFlightCheck(t *testing.T) {
	ch := testchecker.NewBlockingChecker()
	p := probe.NewProbe("blocking", time.Hour, ch)

	started := testprobe.StartProbe(p)
	testprobe.RequireNoStartResult(t, started, "start returned before the initial check was canceled")
	testchecker.WaitForStarted(t, ch.Started)

	stopped := testprobe.StopProbe(p)
	testprobe.WaitForStopped(t, stopped)
	testchecker.WaitForCanceled(t, ch.Canceled)

	result := testprobe.WaitForStart(t, started, "start")
	require.NoError(t, result.Err)

	_, ok := <-result.Ticks
	require.False(t, ok)
}

func TestStopCancelsPeriodicInFlightCheck(t *testing.T) {
	ch := testchecker.NewCancelablePeriodicChecker()
	p := probe.NewProbe("blocking", time.Millisecond, ch)

	ticks, err := p.Start(t.Context())
	require.NoError(t, err)
	require.NotNil(t, <-ticks)
	testchecker.WaitForStarted(t, ch.PeriodicStarted)

	stopped := testprobe.StopProbe(p)
	testprobe.WaitForStopped(t, stopped)
	testchecker.WaitForCanceled(t, ch.PeriodicCanceled)

	_, ok := <-ticks
	require.False(t, ok)
}

func TestConcurrentStartWaitsForInitialCheck(t *testing.T) {
	ch := testchecker.NewReleasableChecker()
	p := probe.NewProbe("blocking", time.Hour, ch)
	t.Cleanup(func() { _ = p.Stop(context.Background()) })

	first := testprobe.StartProbe(p)
	testchecker.WaitForStarted(t, ch.Started)

	second := testprobe.StartProbe(p)

	close(ch.Release)

	firstResult := testprobe.WaitForStart(t, first, "first start")
	secondResult := testprobe.WaitForStart(t, second, "second start")

	require.NoError(t, firstResult.Err)
	require.NoError(t, secondResult.Err)
	require.Equal(t, firstResult.Ticks, secondResult.Ticks)
}

func TestStartWithCanceledContextReturnsContextError(t *testing.T) {
	p := probe.NewProbe("noop", time.Hour, checker.NewNoopChecker())
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	ticks, err := p.Start(ctx)

	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, ticks)
}

func TestStartReturnsContextError(t *testing.T) {
	ch := testchecker.NewBlockingChecker()
	p := probe.NewProbe("blocking", time.Hour, ch)

	ctx, cancel := context.WithCancel(t.Context())
	started := make(chan testprobe.StartResult, 1)
	go func() {
		ticks, err := p.Start(ctx)
		started <- testprobe.StartResult{Ticks: ticks, Err: err}
	}()

	testchecker.WaitForStarted(t, ch.Started)
	cancel()
	testchecker.WaitForCanceled(t, ch.Canceled)

	result := testprobe.WaitForStart(t, started, "start")
	require.ErrorIs(t, result.Err, context.Canceled)
	require.Nil(t, result.Ticks)
}

func TestCanceledConcurrentStartDoesNotStopExistingStart(t *testing.T) {
	ch := testchecker.NewReleasableChecker()
	p := probe.NewProbe("blocking", time.Hour, ch)
	t.Cleanup(func() { _ = p.Stop(context.Background()) })

	first := testprobe.StartProbe(p)
	testchecker.WaitForStarted(t, ch.Started)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	ticks, err := p.Start(ctx)
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, ticks)

	close(ch.Release)

	result := testprobe.WaitForStart(t, first, "first start")
	require.NoError(t, result.Err)
	require.NotNil(t, <-result.Ticks)
}

func TestStopReturnsContextError(t *testing.T) {
	ch := testchecker.NewBlockingPeriodicChecker()
	p := probe.NewProbe("blocking", time.Millisecond, ch)
	t.Cleanup(func() {
		close(ch.Release)
		_ = p.Stop(context.Background())
	})

	ticks, err := p.Start(t.Context())
	require.NoError(t, err)
	require.NotNil(t, <-ticks)
	testchecker.WaitForStarted(t, ch.PeriodicStarted)

	ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
	defer cancel()

	require.ErrorIs(t, p.Stop(ctx), context.DeadlineExceeded)
}

func TestStartWithInvalidPeriodReturnsErrorTick(t *testing.T) {
	tests := []struct {
		name   string
		period time.Duration
	}{
		{name: "zero", period: 0},
		{name: "negative", period: -time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := probe.NewProbe("noop", tt.period, checker.NewNoopChecker())

			require.NotPanics(t, func() {
				ticks, err := p.Start(t.Context())
				require.NoError(t, err)

				tick, ok := <-ticks
				require.True(t, ok)
				require.Equal(t, "noop", tick.Name())
				require.ErrorIs(t, tick.Error(), probe.ErrInvalidPeriod)
				require.ErrorContains(t, tick.Error(), tt.period.String())

				_, ok = <-ticks
				require.False(t, ok)
			})
		})
	}
}
