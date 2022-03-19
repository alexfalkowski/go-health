package checker

import (
	"context"
	"time"

	"github.com/alexfalkowski/go-health/sql"
)

// NewDBChecker for SQL.
func NewDBChecker(pinger sql.Pinger, timeout time.Duration) *DBChecker {
	return &DBChecker{pinger: pinger, timeout: timeout}
}

// DBChecker for SQL.
type DBChecker struct {
	pinger  sql.Pinger
	timeout time.Duration
}

// Check the DB.
func (c *DBChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.pinger.PingContext(ctx)
}
