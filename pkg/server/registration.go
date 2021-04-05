package server

import (
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/pkg/checker"
)

const (
	defaultPeriod = 10 * time.Second
)

// NewRegistration for server.
func NewRegistration(name string, period time.Duration, ch checker.Checker) *Registration {
	if period == 0 {
		period = defaultPeriod
	}

	return &Registration{Name: name, Period: period, Checker: ch}
}

// Registration for the server.
type Registration struct {
	Name    string
	Period  time.Duration
	Checker checker.Checker
}

func (r *Registration) String() string {
	return fmt.Sprintf("name: '%s', period: '%s', checker: '%s'", r.Name, r.Period, r.Checker)
}
