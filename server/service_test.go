package server_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/internal/test"
	testchecker "github.com/alexfalkowski/go-health/v2/internal/test/checker"
	testsubscriber "github.com/alexfalkowski/go-health/v2/internal/test/subscriber"
	"github.com/alexfalkowski/go-health/v2/server"
	"github.com/alexfalkowski/go-sync"
	"github.com/stretchr/testify/require"
)

func TestServiceStartStopGuards(t *testing.T) {
	s := server.NewService()

	registration := server.NewRegistration("noop", 10*time.Millisecond, checker.NewNoopChecker())
	s.Register(registration)
	require.NoError(t, s.Observe("livez", registration.Name))

	require.NotPanics(t, func() {
		_ = s.Stop(t.Context())
	})

	require.NotPanics(t, func() {
		_ = s.Start(t.Context())
		_ = s.Start(t.Context())
	})

	require.NotPanics(t, func() {
		_ = s.Stop(t.Context())
		_ = s.Stop(t.Context())
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

	_ = s.Stop(t.Context())
	testsubscriber.RequireObserverStopped(t, observer)

	_ = s.Start(t.Context())
	t.Cleanup(func() { _ = s.Stop(context.Background()) })

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
			_ = s.Stop(t.Context())
		case <-time.After(time.Second):
			require.Fail(t, "service did not finish starting")
		}
	}()

	go func() {
		_ = s.Start(t.Context())
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
