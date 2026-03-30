package checker_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/stretchr/testify/require"
)

func TestTCPCheckerZeroTimeoutUsesDefault(t *testing.T) {
	dialer := &recordingDialer{}
	checker := checker.NewTCPChecker("example:80", 0, checker.WithDialer(dialer))

	require.NoError(t, checker.Check(context.Background()))
	require.True(t, dialer.hasDeadline)
	require.Greater(t, dialer.remaining, 25*time.Second)
}

type recordingDialer struct {
	hasDeadline bool
	remaining   time.Duration
}

func (d *recordingDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	deadline, ok := ctx.Deadline()
	d.hasDeadline = ok
	if ok {
		d.remaining = time.Until(deadline)
	}

	conn, peer := net.Pipe()
	_ = peer.Close()

	return conn, nil
}
