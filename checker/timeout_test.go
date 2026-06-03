package checker_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/sql"
	"github.com/alexfalkowski/go-sync"
	"github.com/stretchr/testify/require"
)

func TestDBCheckerTimeoutCause(t *testing.T) {
	c := checker.NewDBChecker(timeoutPinger{}, time.Microsecond)

	err := c.Check(t.Context())

	require.ErrorIs(t, err, checker.ErrTimeout)
	require.ErrorIs(t, err, sync.ErrTimeout)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestTCPCheckerTimeoutCause(t *testing.T) {
	c := checker.NewTCPChecker("example:80", time.Microsecond, checker.WithDialer(timeoutDialer{}))

	err := c.Check(t.Context())

	require.ErrorIs(t, err, checker.ErrTimeout)
	require.ErrorIs(t, err, sync.ErrTimeout)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestDBCheckerZeroTimeoutUsesDefault(t *testing.T) {
	pinger := &recordingPinger{}
	c := checker.NewDBChecker(pinger, 0)

	require.NoError(t, c.Check(t.Context()))
	requireDefaultDeadline(t, pinger.hasDeadline, pinger.remaining)
}

func TestHTTPCheckerZeroTimeoutUsesDefault(t *testing.T) {
	transport := &deadlineRoundTripper{status: http.StatusNoContent}
	c := checker.NewHTTPChecker("https://example.com/health", 0, checker.WithRoundTripper(transport))

	require.NoError(t, c.Check(t.Context()))
	requireDefaultDeadline(t, transport.hasDeadline, transport.remaining)
}

func TestOnlineCheckerZeroTimeoutUsesDefault(t *testing.T) {
	transport := &deadlineRoundTripper{status: http.StatusNoContent}
	c := checker.NewOnlineChecker(0, checker.WithRoundTripper(transport), checker.WithURLs("https://example.com/health"))

	require.NoError(t, c.Check(t.Context()))
	requireDefaultDeadline(t, transport.hasDeadline, transport.remaining)
}

func requireDefaultDeadline(t *testing.T, hasDeadline bool, remaining time.Duration) {
	t.Helper()

	require.True(t, hasDeadline, "expected checker to configure a default timeout deadline")
	require.Greater(t, remaining, 25*time.Second, "expected deadline to be close to the 30s default")
	require.Less(t, remaining, 31*time.Second, "expected deadline to stay near the 30s default")
}

type timeoutPinger struct{}

func (timeoutPinger) PingContext(ctx context.Context) error {
	<-ctx.Done()
	return context.Cause(ctx)
}

var _ sql.Pinger = timeoutPinger{}

type timeoutDialer struct{}

func (timeoutDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	<-ctx.Done()
	return nil, context.Cause(ctx)
}

type recordingPinger struct {
	hasDeadline bool
	remaining   time.Duration
}

func (p *recordingPinger) PingContext(ctx context.Context) error {
	deadline, ok := ctx.Deadline()
	p.hasDeadline = ok
	if ok {
		p.remaining = time.Until(deadline)
	}

	return nil
}

type deadlineRoundTripper struct {
	hasDeadline bool
	remaining   time.Duration
	status      int
}

func (t *deadlineRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	deadline, ok := req.Context().Deadline()
	t.hasDeadline = ok
	if ok {
		t.remaining = time.Until(deadline)
	}

	return &http.Response{
		StatusCode: t.status,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
