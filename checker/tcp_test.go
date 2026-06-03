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

func TestTCPCheckerClosesSuccessfulConnection(t *testing.T) {
	dialer := &recordingDialer{}
	checker := checker.NewTCPChecker("example:80", time.Second, checker.WithDialer(dialer))

	require.NoError(t, checker.Check(context.Background()))
	require.True(t, dialer.conn.closed)
}

type recordingDialer struct {
	conn        *recordingConn
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

	d.conn = &recordingConn{}

	return d.conn, nil
}

type recordingConn struct {
	closed bool
}

func (c *recordingConn) Close() error {
	c.closed = true
	return nil
}

func (c *recordingConn) Read([]byte) (int, error) {
	return 0, net.ErrClosed
}

func (c *recordingConn) Write([]byte) (int, error) {
	return 0, net.ErrClosed
}

func (c *recordingConn) LocalAddr() net.Addr {
	return recordingAddr("local")
}

func (c *recordingConn) RemoteAddr() net.Addr {
	return recordingAddr("remote")
}

func (c *recordingConn) SetDeadline(time.Time) error {
	return nil
}

func (c *recordingConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *recordingConn) SetWriteDeadline(time.Time) error {
	return nil
}

type recordingAddr string

func (a recordingAddr) Network() string {
	return string(a)
}

func (a recordingAddr) String() string {
	return string(a)
}
