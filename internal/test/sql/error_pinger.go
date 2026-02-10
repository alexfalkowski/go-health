package sql

import "context"

// ErrorPinger is a Pinger that always returns Err.
type ErrorPinger struct {
	Err error
}

func (p ErrorPinger) PingContext(context.Context) error {
	return p.Err
}
