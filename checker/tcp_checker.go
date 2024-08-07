package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	hn "github.com/alexfalkowski/go-health/net"
)

// NewTCPChecker for address.
func NewTCPChecker(address string, timeout time.Duration) *TCPChecker {
	return NewTCPCheckerWithDialer(address, timeout, &net.Dialer{})
}

// NewTCPCheckerWithDialer for address.
func NewTCPCheckerWithDialer(address string, timeout time.Duration, dialer hn.Dialer) *TCPChecker {
	return &TCPChecker{address: address, timeout: timeout, dialer: dialer}
}

// TCPChecker for an address.
type TCPChecker struct {
	dialer  hn.Dialer
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

	return conn.Close()
}
