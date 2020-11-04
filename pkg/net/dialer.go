package net

import (
	"context"
	"net"
)

// Dialer connects to the address on the named network.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
