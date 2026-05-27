package probe_test

import (
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
		p.Stop()
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

	_, ok := <-result.Ticks
	require.False(t, ok)
}

func TestConcurrentStartWaitsForInitialCheck(t *testing.T) {
	ch := testchecker.NewReleasableChecker()
	p := probe.NewProbe("blocking", time.Hour, ch)
	t.Cleanup(p.Stop)

	first := testprobe.StartProbe(p)
	testchecker.WaitForStarted(t, ch.Started)

	second := testprobe.StartProbe(p)

	close(ch.Release)

	firstResult := testprobe.WaitForStart(t, first, "first start")
	secondResult := testprobe.WaitForStart(t, second, "second start")

	require.Equal(t, firstResult.Ticks, secondResult.Ticks)
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
