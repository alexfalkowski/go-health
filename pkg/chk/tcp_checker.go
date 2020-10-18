package chk

import (
	"context"
	"net"
	"time"
)

// NewTCPChecker with address.
func NewTCPChecker(address string, timeout time.Duration) *TCPChecker {
	return &TCPChecker{address, timeout}
}

// TCPChecker for an address.
type TCPChecker struct {
	address string
	timeout time.Duration
}

// Check the address.
func (c *TCPChecker) Check(ctx context.Context) error {
	_, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}

	return nil
}
