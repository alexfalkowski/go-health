package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/net"
)

var _ Checker = (*TCPChecker)(nil)

// NewTCPChecker with address, timeout.
func NewTCPChecker(address string, timeout time.Duration, opts ...Option) *TCPChecker {
	os := parseOptions(opts...)

	return &TCPChecker{address: address, timeout: timeout, dialer: os.dialer}
}

// TCPChecker for an address.
type TCPChecker struct {
	dialer  net.Dialer
	address string
	timeout time.Duration
}

// Check the address.
func (c *TCPChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("tcp checker: %w", err)
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("tcp checker: %w", err)
	}

	return nil
}
