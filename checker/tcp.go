package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/net"
)

var _ Checker = (*TCPChecker)(nil)

// NewTCPChecker returns a TCPChecker that connects to address.
//
// The timeout is applied via context.WithTimeout during Check. Passing 0 uses
// the package default of 30 seconds. Use WithDialer to override the dialing
// implementation.
func NewTCPChecker(address string, t time.Duration, opts ...Option) *TCPChecker {
	os := parseOptions(opts...)

	return &TCPChecker{address: address, timeout: timeout(t), dialer: os.dialer}
}

// TCPChecker checks TCP connectivity to an address.
//
// A successful dial is considered healthy. The connection is closed immediately
// after it is established.
type TCPChecker struct {
	dialer  net.Dialer
	address string
	timeout time.Duration
}

// Check attempts to connect to the configured address with a per-call timeout.
func (c *TCPChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("tcp checker: %w", err)
	}
	defer conn.Close()

	return nil
}
