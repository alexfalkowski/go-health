package server_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
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

func TestServiceStartRunsInitialChecksConcurrently(t *testing.T) {
	s := server.NewService()
	release := make(chan struct{})
	var releaseOnce sync.Once

	first := &blockingStartChecker{started: make(chan struct{}), release: release}
	second := &blockingStartChecker{started: make(chan struct{}), release: release}

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
		return channelClosed(first.started) && channelClosed(second.started)
	}, time.Second, 10*time.Millisecond)

	releaseOnce.Do(func() {
		close(release)
	})

	require.Eventually(t, func() bool {
		return channelClosed(started)
	}, time.Second, 10*time.Millisecond)
}

type blockingStartChecker struct {
	started chan struct{}
	release <-chan struct{}
}

func (c *blockingStartChecker) Check(ctx context.Context) error {
	close(c.started)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.release:
		return nil
	}
}

func channelClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
