package chk

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// NewDBChecker for SQL.
func NewDBChecker(db *sql.DB, timeout time.Duration) *DBChecker {
	return &DBChecker{db, timeout}
}

// DBChecker for SQL.
type DBChecker struct {
	db      *sql.DB
	timeout time.Duration
}

// Check the DB.
func (c *DBChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.db.PingContext(ctx)
}

func (c *DBChecker) String() string {
	return fmt.Sprintf("db: sql, timeout: %s", c.timeout)
}
