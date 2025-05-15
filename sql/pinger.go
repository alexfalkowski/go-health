package sql

import "context"

// Pinger verifies a connection to the database is still alive.
type Pinger interface {
	PingContext(ctx context.Context) error
}
