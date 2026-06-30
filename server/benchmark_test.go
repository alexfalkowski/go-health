package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/server"
	"github.com/stretchr/testify/require"
)

func BenchmarkServerErrorHealthyObserver(b *testing.B) {
	// Benchmark the public aggregate health read path with initialized observer state.
	b.ReportAllocs()

	s := server.NewServer()
	registration := server.NewRegistration("noop", time.Hour, checker.NewNoopChecker())
	s.Register("test", registration)
	require.NoError(b, s.Observe("test", "livez", registration.Name))

	b.ResetTimer()

	for b.Loop() {
		if err := s.Error("livez"); err != nil {
			require.NoError(b, err)
		}
	}

	b.StopTimer()
	require.NoError(b, s.Stop(context.Background()))
}
