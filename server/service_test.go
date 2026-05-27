package server_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/internal/test"
	testchecker "github.com/alexfalkowski/go-health/v2/internal/test/checker"
	testsubscriber "github.com/alexfalkowski/go-health/v2/internal/test/subscriber"
	"github.com/alexfalkowski/go-health/v2/server"
	"github.com/stretchr/testify/require"
)

func TestServiceStartStopGuards(t *testing.T) {
	s := server.NewService()

	registration := server.NewRegistration("noop", 10*time.Millisecond, checker.NewNoopChecker())
	s.Register(registration)
	require.NoError(t, s.Observe("livez", registration.Name))

	require.NotPanics(t, func() {
		s.Stop()
	})

	require.NotPanics(t, func() {
		s.Start()
		s.Start()
	})

	require.NotPanics(t, func() {
		s.Stop()
		s.Stop()
	})
}

func TestServiceObserveRejectsUnknownProbes(t *testing.T) {
	s := server.NewService()

	registration := server.NewRegistration("noop", 10*time.Millisecond, checker.NewNoopChecker())
	s.Register(registration)

	err := s.Observe("livez", registration.Name, "missing")
	require.ErrorIs(t, err, server.ErrProbeNotFound)

	_, err = s.Observer("livez")
	require.ErrorIs(t, err, server.ErrObserverNotFound)
}

//nolint:err113
func TestServiceStopBeforeStartClosesObservers(t *testing.T) {
	s := server.NewService()
	errNotReady := errors.New("not ready")

	registration := server.NewRegistration("ready", 10*time.Millisecond, checker.NewReadyChecker(errNotReady))
	s.Register(registration)
	require.NoError(t, s.Observe("livez", registration.Name))

	observer, err := s.Observer("livez")
	require.NoError(t, err)

	s.Stop()
	testsubscriber.RequireObserverStopped(t, observer)

	s.Start()
	t.Cleanup(s.Stop)

	require.Eventually(t, func() bool {
		return errors.Is(observer.Error(), errNotReady)
	}, time.Second, 10*time.Millisecond)
}

func TestServiceStartRunsInitialChecksConcurrently(t *testing.T) {
	s := server.NewService()
	release := make(chan struct{})
	var releaseOnce sync.Once

	first := testchecker.NewReleasableStartChecker(release)
	second := testchecker.NewReleasableStartChecker(release)

	s.Register(
		server.NewRegistration("first", time.Hour, first),
		server.NewRegistration("second", time.Hour, second),
	)

	started := make(chan struct{})
	defer func() {
		releaseOnce.Do(func() {
			close(release)
		})

		select {
		case <-started:
			s.Stop()
		case <-time.After(time.Second):
			require.Fail(t, "service did not finish starting")
		}
	}()

	go func() {
		s.Start()
		close(started)
	}()

	require.Eventually(t, func() bool {
		return test.ChannelClosed(first.Started) && test.ChannelClosed(second.Started)
	}, time.Second, 10*time.Millisecond)

	releaseOnce.Do(func() {
		close(release)
	})

	require.Eventually(t, func() bool {
		return test.ChannelClosed(started)
	}, time.Second, 10*time.Millisecond)
}
