package checker_test

import (
	"context"
	"net"
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
