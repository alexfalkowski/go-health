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

func TestServiceStartStop(t *testing.T) {
	s := server.NewService()

	registration := server.NewRegistration("noop", 10*time.Millisecond, checker.NewNoopChecker())
	s.Register(registration)
	require.NoError(t, s.Observe("livez", registration.Name))

	require.NoError(t, s.Start(t.Context()))
	require.NoError(t, s.Stop(t.Context()))
}

func TestServiceObserveRejectsUnknownProbes(t *testing.T) {
	t.Parallel()

	s := server.NewService()

	registration := server.NewRegistration("noop", 10*time.Millisecond, checker.NewNoopChecker())
	s.Register(registration)

	err := s.Observe("livez", registration.Name, "missing")
	require.ErrorIs(t, err, server.ErrProbeNotFound)

	_, err = s.Observer("livez")
	require.ErrorIs(t, err, server.ErrObserverNotFound)
}

func TestServiceObserveKeepsOriginalProbeSetForExistingKind(t *testing.T) {
	t.Parallel()

	s := server.NewService()

	first := server.NewRegistration("first", time.Hour, checker.NewNoopChecker())
	second := server.NewRegistration("second", time.Hour, checker.NewNoopChecker())
	s.Register(first, second)

	require.NoError(t, s.Observe("livez", first.Name))
	observer, err := s.Observer("livez")
	require.NoError(t, err)

	require.NoError(t, s.Observe("livez", second.Name))
	sameObserver, err := s.Observer("livez")
	require.NoError(t, err)

	require.Same(t, observer, sameObserver)
	require.Equal(t, []string{first.Name}, sameObserver.Names())
}

func TestServiceErrorReturnsObserverError(t *testing.T) {
	s := server.NewService()
	t.Cleanup(func() { _ = s.Stop(context.Background()) })

	errNotReady := errors.New("not ready")
	registration := server.NewRegistration("ready", time.Hour, checker.NewReadyChecker(errNotReady))
	s.Register(registration)
	require.NoError(t, s.Observe("readyz", registration.Name))
	watcher, err := s.Watch("readyz")
	require.NoError(t, err)
	defer watcher.Close()
	updates := watcher.Receive()
	require.NoError(t, receiveError(t, updates))

	require.NoError(t, s.Start(t.Context()))

	require.Eventually(t, func() bool {
		return errors.Is(s.Error("readyz"), errNotReady)
	}, time.Second, 10*time.Millisecond)
	require.ErrorIs(t, receiveMatchingError(t, updates, func(err error) bool {
		return errors.Is(err, errNotReady)
	}), errNotReady)
}

func TestServiceErrorAndWatchRejectUnknownObserver(t *testing.T) {
	t.Parallel()

	s := server.NewService()

	require.ErrorIs(t, s.Error("readyz"), server.ErrObserverNotFound)

	_, err := s.Watch("readyz")
	require.ErrorIs(t, err, server.ErrObserverNotFound)
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

func TestServiceStartFailureCleansUpStartedProbesAndObservers(t *testing.T) {
	s := server.NewService()
	started := testchecker.NewCancelablePeriodicChecker()
	blocking := testchecker.NewBlockingChecker()

	startedRegistration := server.NewRegistration("started", time.Millisecond, started)
	blockingRegistration := server.NewRegistration("blocking", time.Hour, blocking)
	s.Register(startedRegistration, blockingRegistration)
	require.NoError(t, s.Observe("livez", startedRegistration.Name))

	observer, err := s.Observer("livez")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	errc := make(chan error, 1)
	go func() {
		errc <- s.Start(ctx)
	}()

	testchecker.WaitForStarted(t, blocking.Started)
	testchecker.WaitForStarted(t, started.PeriodicStarted)

	cancel()

	require.ErrorIs(t, <-errc, context.Canceled)
	testchecker.WaitForCanceled(t, blocking.Canceled)
	testchecker.WaitForCanceled(t, started.PeriodicCanceled)
	testsubscriber.RequireObserverStopped(t, observer)
	require.NoError(t, s.Stop(t.Context()))
}
