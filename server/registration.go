package server

import (
	"time"

	"github.com/alexfalkowski/go-health/checker"
)

// NewRegistration for server.
func NewRegistration(name string, period time.Duration, ch checker.Checker) *Registration {
	return &Registration{Name: name, Period: period, Checker: ch}
}

// Registration for the server.
type Registration struct {
	Checker checker.Checker
	Name    string
	Period  time.Duration
}
