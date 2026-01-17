package probe_test

import (
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/internal/test"
	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/stretchr/testify/require"
)

func TestStart(t *testing.T) {
	checker := checker.NewHTTPChecker(test.StatusURL("400"), time.Second)
	probe := probe.NewProbe("google", time.Millisecond, checker)
	defer probe.Stop()

	probe.Start()
	require.NotNil(t, probe.Start())
}

func TestStop(t *testing.T) {
	checker := checker.NewHTTPChecker(test.StatusURL("400"), time.Second)
	probe := probe.NewProbe("google", time.Millisecond, checker)

	probe.Stop()

	require.NotNil(t, probe.Start())
	probe.Stop()
}
