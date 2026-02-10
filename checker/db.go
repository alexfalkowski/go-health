package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/sql"
)

var _ Checker = (*DBChecker)(nil)

// NewDBChecker returns a DBChecker that pings pinger.
//
// The timeout is applied via context.WithTimeout during Check.
func NewDBChecker(pinger sql.Pinger, t time.Duration) *DBChecker {
	return &DBChecker{pinger: pinger, timeout: timeout(t)}
}

// DBChecker checks a SQL database connection.
type DBChecker struct {
	pinger  sql.Pinger
	timeout time.Duration
}

// Check pings the database.
func (c *DBChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := c.pinger.PingContext(ctx); err != nil {
		return fmt.Errorf("db checker: %w", err)
	}

	return nil
}
