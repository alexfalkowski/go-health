package server_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/server"
	"github.com/stretchr/testify/require"
)

//nolint:err113
func TestRestartKeepsObserverReceivingTicks(t *testing.T) {
	s := server.NewServer()
	t.Cleanup(s.Stop)

	errBeforeRestart := errors.New("before restart")
	errAfterRestart := errors.New("after restart")

	checker := &dynamicChecker{err: errBeforeRestart}
	registration := server.NewRegistration("dynamic", 10*time.Millisecond, checker)
	s.Register("test", registration)

	require.NoError(t, s.Observe("test", "livez", registration.Name))

	observer, err := s.Observer("test", "livez")
	require.NoError(t, err)

	// Confirm the observer receives the initial unhealthy state.
	s.Start()
	require.Eventually(t, func() bool {
		return errors.Is(observer.Error(), errBeforeRestart)
	}, time.Second, 10*time.Millisecond)

	// Confirm the same observer updates before shutdown.
	checker.Set(nil)
	require.Eventually(t, func() bool {
		return observer.Error() == nil
	}, time.Second, 10*time.Millisecond)

	// Restart the server with a different probe result.
	s.Stop()

	checker.Set(errAfterRestart)
	s.Start()

	// Confirm the existing observer receives ticks after restart.
	require.Eventually(t, func() bool {
		return errors.Is(observer.Error(), errAfterRestart)
	}, time.Second, 10*time.Millisecond)
}

type dynamicChecker struct {
	err error
	mux sync.RWMutex
}

func (c *dynamicChecker) Check(context.Context) error {
	c.mux.RLock()
	defer c.mux.RUnlock()

	return c.err
}

func (c *dynamicChecker) Set(err error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.err = err
}
