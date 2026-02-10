// Package net contains small network-related interfaces.
package net

import (
	"context"
	"net"
)

// DefaultDialer is just a net.Dialer.
var DefaultDialer Dialer = &net.Dialer{}

// Dialer connects to the address on the named network.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
