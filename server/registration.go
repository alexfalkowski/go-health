package server

import (
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
)

// NewOnlineRegistration returns a registration for an online checker.
//
// The registration name is fixed to "online". If you need a different name, use
// NewRegistration with checker.NewOnlineChecker directly.
func NewOnlineRegistration(timeout, period time.Duration, opts ...checker.Option) *Registration {
	return &Registration{
		Name:    "online",
		Period:  period,
		Checker: checker.NewOnlineChecker(timeout, opts...),
	}
}

// NewRegistration returns a registration for a checker.
func NewRegistration(name string, period time.Duration, ch checker.Checker) *Registration {
	return &Registration{Name: name, Period: period, Checker: ch}
}

// Registration describes a probe to run for a service.
type Registration struct {
	Checker checker.Checker
	Name    string
	Period  time.Duration
}
