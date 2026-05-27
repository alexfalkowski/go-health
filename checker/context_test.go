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

	checkers := map[string]checker.Checker{
		"noop":   checker.NewNoopChecker(),
		"ready":  checker.NewReadyChecker(errNotReady),
		"http":   checker.NewHTTPChecker("://bad-url", time.Second),
		"online": checker.NewOnlineChecker(time.Second, checker.WithURLs("://bad-url")),
		"db":     checker.NewDBChecker(sql.CanceledPinger{}, time.Second),
		"tcp":    checker.NewTCPChecker("example:80", time.Second, checker.WithDialer(net.CanceledDialer{})),
	}

	for name, check := range checkers {
		t.Run(name, func(t *testing.T) {
			err := check.Check(ctx)

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
