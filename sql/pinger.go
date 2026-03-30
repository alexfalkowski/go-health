package sql

import "context"

// Pinger verifies a connection to the database is still alive.
//
// The interface matches the method exposed by *sql.DB so callers can use a real
// database handle directly or substitute a small test double.
type Pinger interface {
	PingContext(ctx context.Context) error
}
