package net

import (
	"context"
	"net"
)

// CanceledDialer is a Dialer that returns the supplied context error.
type CanceledDialer struct{}

func (CanceledDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	return nil, ctx.Err()
}
