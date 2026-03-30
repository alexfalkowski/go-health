package server_test

import (
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
	s.Observe("livez", registration.Name)

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
