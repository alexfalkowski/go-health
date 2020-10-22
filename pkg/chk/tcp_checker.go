package chk

import (
	"context"
	"fmt"
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
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}

	return conn.Close()
}

func (c *TCPChecker) String() string {
	return fmt.Sprintf("address: %s, timeout: %s", c.address, c.timeout)
}
