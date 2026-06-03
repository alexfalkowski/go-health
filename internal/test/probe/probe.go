package probe

import (
	"context"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/stretchr/testify/require"
)

// StartResult captures a probe Start result.
type StartResult struct {
	Ticks <-chan *probe.Tick
	Err   error
}

// StartProbe starts p in a goroutine and returns its result channel.
func StartProbe(p *probe.Probe) <-chan StartResult {
	started := make(chan StartResult, 1)
	go func() {
		ticks, err := p.Start(context.Background())
		started <- StartResult{Ticks: ticks, Err: err}
	}()
	return started
}

// RequireNoStartResult requires that started does not receive immediately.
func RequireNoStartResult(t *testing.T, started <-chan StartResult, message string) {
	t.Helper()

	select {
	case <-started:
		require.Fail(t, message)
	case <-time.After(50 * time.Millisecond):
	}
}

// WaitForStart waits until started receives a StartResult.
func WaitForStart(t *testing.T, started <-chan StartResult, name string) StartResult {
	t.Helper()

	select {
	case result := <-started:
		return result
	case <-time.After(time.Second):
		require.Fail(t, name+" did not return")
		return StartResult{}
	}
}

// StopProbe stops p in a goroutine and returns a completion channel.
func StopProbe(p *probe.Probe) <-chan struct{} {
	stopped := make(chan struct{})
	go func() {
		_ = p.Stop(context.Background())
		close(stopped)
	}()
	return stopped
}

// WaitForStopped waits until stopped is closed.
func WaitForStopped(t *testing.T, stopped <-chan struct{}) {
	t.Helper()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		require.Fail(t, "stop did not cancel the in-flight check")
	}
}
