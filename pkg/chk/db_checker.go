package chk

import (
	"context"
	"fmt"
	"time"
)

// NewDBChecker for SQL.
func NewDBChecker(pinger Pinger, timeout time.Duration) *DBChecker {
	return &DBChecker{pinger: pinger, timeout: timeout}
}

// DBChecker for SQL.
type DBChecker struct {
	pinger  Pinger
	timeout time.Duration
}

// Check the DB.
func (c *DBChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.pinger.PingContext(ctx)
}

func (c *DBChecker) String() string {
	return fmt.Sprintf("db: 'sql', timeout: '%s'", c.timeout)
}
