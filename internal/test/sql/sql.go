package sql

import "context"

// CanceledPinger is a Pinger that returns the supplied context error.
type CanceledPinger struct{}

func (CanceledPinger) PingContext(ctx context.Context) error {
	return ctx.Err()
}

// ErrorPinger is a Pinger that always returns Err.
type ErrorPinger struct {
	Err error
}

func (p ErrorPinger) PingContext(context.Context) error {
	return p.Err
}

// OKPinger is a Pinger that always succeeds.
type OKPinger struct{}

func (OKPinger) PingContext(context.Context) error {
	return nil
}
