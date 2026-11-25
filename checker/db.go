package checker

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var _ Checker = (*DBChecker)(nil)

// NewDBChecker returns a DBChecker that runs SELECT 1 using db.
//
// The caller supplies an initialized, non-nil database handle and manages its
// lifetime. The database must support SELECT 1 without a FROM clause. For other
// dialects, implement [Checker] with a query the database accepts. Passing 0
// uses the package default of 30 seconds.
func NewDBChecker(db *sql.DB, t time.Duration) *DBChecker {
	return &DBChecker{db: db, timeout: timeout(t)}
}

// DBChecker checks whether a SQL database can execute a query and return a row.
type DBChecker struct {
	db      *sql.DB
	timeout time.Duration
}

// Check runs SELECT 1 and reads the result with a per-call timeout.
//
// If the query fails after the checker's deadline expires, the returned error
// matches [ErrTimeout].
func (c *DBChecker) Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeoutCause(ctx, c.timeout, ErrTimeout)
	defer cancel()

	var one int
	if err := c.db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		if ctx.Err() != nil {
			return context.Cause(ctx)
		}

		return fmt.Errorf("db checker: %w", err)
	}

	return nil
}
