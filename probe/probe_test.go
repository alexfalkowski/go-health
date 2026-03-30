package probe_test

import (
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/stretchr/testify/require"
)

func TestStopWithoutStartDoesNotPanic(t *testing.T) {
	p := probe.NewProbe("noop", 10*time.Millisecond, checker.NewNoopChecker())

	require.NotPanics(t, func() {
		p.Stop()
	})
}
