package svr

import (
	"time"

	"github.com/alexfalkowski/go-health/pkg/chk"
)

const (
	defaultPeriod = 10 * time.Second
)

// Registration for the server.
type Registration struct {
	Name    string
	Period  time.Duration
	Checker chk.Checker
}
