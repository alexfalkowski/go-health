package sql

import "context"

// OKPinger is a Pinger that always succeeds.
type OKPinger struct{}

func (OKPinger) PingContext(context.Context) error {
	return nil
}
