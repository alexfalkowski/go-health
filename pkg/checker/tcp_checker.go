package checker

import (
	"context"
	"fmt"
	snet "net"
	"time"

	"github.com/alexfalkowski/go-health/pkg/net"
)

// NewTCPChecker for address.
func NewTCPChecker(address string, timeout time.Duration) *TCPChecker {
	return NewTCPCheckerWithDialer(address, timeout, &snet.Dialer{})
}

// NewTCPCheckerWithDialer for address.
func NewTCPCheckerWithDialer(address string, timeout time.Duration, dialer net.Dialer) *TCPChecker {
	return &TCPChecker{address: address, timeout: timeout, dialer: dialer}
}

// TCPChecker for an address.
type TCPChecker struct {
	address string
	timeout time.Duration
	dialer  net.Dialer
}

// Check the address.
func (c *TCPChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("net dial: %w", err)
	}

	return conn.Close()
}
