package net

import (
	"context"
	"net"
)

// DefaultDialer is the package default Dialer backed by net.Dialer.
var DefaultDialer Dialer = &net.Dialer{}

// Dialer connects to the address on the named network.
//
// It matches the subset of net.Dialer used by the checker package.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
