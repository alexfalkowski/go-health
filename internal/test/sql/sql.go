package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
)

// OpenDB returns a test database that supports SELECT 1. If supplied,
// queryError runs with the query context and sets the error returned by Scan.
func OpenDB(queryError func(context.Context) error) *sql.DB {
	return sql.OpenDB(queryDriver(queryError))
}

type queryDriver func(context.Context) error

func (d queryDriver) Open(string) (driver.Conn, error) {
	return d, nil
}

func (d queryDriver) Connect(context.Context) (driver.Conn, error) {
	return d, nil
}

func (d queryDriver) Driver() driver.Driver {
	return d
}

func (queryDriver) Prepare(string) (driver.Stmt, error) {
	return nil, errors.ErrUnsupported
}

func (queryDriver) Close() error {
	return nil
}

func (queryDriver) Begin() (driver.Tx, error) {
	return nil, errors.ErrUnsupported
}

func (d queryDriver) QueryContext(ctx context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if query != "SELECT 1" {
		return nil, fmt.Errorf("unexpected test query: %q", query)
	}

	rows := &queryRows{}
	if d != nil {
		rows.err = d(ctx)
	}

	return rows, nil
}

type queryRows struct {
	err  error
	read bool
}

func (*queryRows) Columns() []string {
	return []string{"one"}
}

func (*queryRows) Close() error {
	return nil
}

func (r *queryRows) Next(dest []driver.Value) error {
	if r.read {
		return io.EOF
	}

	r.read = true
	dest[0] = int64(1)
	return r.err
}
