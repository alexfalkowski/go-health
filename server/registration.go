package server

import (
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
)

// NewOnlineRegistration returns a registration for an online checker.
//
// The registration name is fixed to "online". If you need a different name, use
// NewRegistration with checker.NewOnlineChecker directly. The period must be
// positive; non-positive periods are reported through observer state as
// probe.ErrInvalidPeriod after the service starts.
func NewOnlineRegistration(timeout, period time.Duration, opts ...checker.Option) *Registration {
	return &Registration{
		Name:    "online",
		Period:  period,
		Checker: checker.NewOnlineChecker(timeout, opts...),
	}
}

// NewRegistration returns a registration for a checker.
//
// The checker must be non-nil and initialized. The period must be positive.
// Non-positive periods are reported through probe and observer state as
// probe.ErrInvalidPeriod after the service starts.
func NewRegistration(name string, period time.Duration, ch checker.Checker) *Registration {
	return &Registration{Name: name, Period: period, Checker: ch}
}

// Registration describes a probe to run for a service.
type Registration struct {
	// Checker must be non-nil and initialized.
	Checker checker.Checker
	Name    string
	// Period is the interval between checks and must be positive.
	Period time.Duration
}
