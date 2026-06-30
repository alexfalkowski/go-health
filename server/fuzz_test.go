package server_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/server"
	"github.com/stretchr/testify/require"
)

// FuzzServiceObserveValidation explores registration and observer names because Service owns observer
// validation and idempotent probe sets.
func FuzzServiceObserveValidation(f *testing.F) {
	f.Add("livez", "db", "cache", "db", "cache", "ignored")
	f.Add("readyz", "db", "cache", "db", "missing", "cache")
	f.Add("", "", "cache", "", "cache", "missing")
	f.Add("livez", "duplicate", "duplicate", "duplicate", "missing", "duplicate")

	f.Fuzz(func(t *testing.T, kind string, reg1 string, reg2 string, obs1 string, obs2 string, replacement string) {
		s := server.NewService()
		t.Cleanup(func() { _ = s.Stop(context.Background()) })

		s.Register(
			server.NewRegistration(reg1, time.Hour, checker.NewNoopChecker()),
			server.NewRegistration(reg2, time.Hour, checker.NewNoopChecker()),
		)

		err := s.Observe(kind, obs1, obs2)
		allObservedRegistered := containsAll([]string{reg1, reg2}, obs1, obs2)
		if !allObservedRegistered {
			require.ErrorIs(t, err, server.ErrProbeNotFound)

			_, observerErr := s.Observer(kind)
			require.ErrorIs(t, observerErr, server.ErrObserverNotFound)
			return
		}
		require.NoError(t, err)

		observer, err := s.Observer(kind)
		require.NoError(t, err)
		originalNames := observer.Names()

		require.NoError(t, s.Observe(kind, replacement))

		sameObserver, err := s.Observer(kind)
		require.NoError(t, err)
		require.Same(t, observer, sameObserver)
		require.Equal(t, originalNames, sameObserver.Names())
	})
}

func containsAll(names []string, observed ...string) bool {
	for _, name := range observed {
		if !slices.Contains(names, name) {
			return false
		}
	}

	return true
}
