package svr

import (
	"fmt"
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

func (r *Registration) String() string {
	return fmt.Sprintf("name: '%s', period: '%s', checker: '%s'", r.Name, r.Period, r.Checker)
}
