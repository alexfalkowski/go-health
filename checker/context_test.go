package checker_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/internal/test/net"
	"github.com/alexfalkowski/go-health/v2/internal/test/sql"
	"github.com/stretchr/testify/require"
)

func TestCheckersReturnCanceledContext(t *testing.T) {
	errNotReady := errors.New("not ready")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	checkers := []struct {
		check checker.Checker
		name  string
	}{
		{check: checker.NewNoopChecker(), name: "noop"},
		{check: checker.NewReadyChecker(errNotReady), name: "ready"},
		{check: checker.NewHTTPChecker("://bad-url", time.Second), name: "http"},
		{check: checker.NewOnlineChecker(time.Second, checker.WithURLs("://bad-url")), name: "online"},
		{check: checker.NewDBChecker(sql.CanceledPinger{}, time.Second), name: "db"},
		{check: checker.NewTCPChecker("example:80", time.Second, checker.WithDialer(net.CanceledDialer{})), name: "tcp"},
	}

	for _, tt := range checkers {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.check.Check(ctx)

			require.ErrorIs(t, err, context.Canceled)
			require.NotErrorIs(t, err, errNotReady)
		})
	}
}

func TestOnlineCheckerReturnsCanceledContextWhileRequestsAreRunning(t *testing.T) {
	started := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	t.Cleanup(upstream.Close)

	check := checker.NewOnlineChecker(time.Second, checker.WithURLs(upstream.URL))
	ctx, cancel := context.WithCancel(t.Context())

	errc := make(chan error, 1)
	go func() {
		errc <- check.Check(ctx)
	}()

	<-started
	cancel()

	require.ErrorIs(t, <-errc, context.Canceled)
}
