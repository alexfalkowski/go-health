package server

import (
	"time"

	"github.com/alexfalkowski/go-health/pkg/checker"
)

// NewRegistration for server.
func NewRegistration(name string, period time.Duration, ch checker.Checker) *Registration {
	return &Registration{Name: name, Period: period, Checker: ch}
}

// Registration for the server.
type Registration struct {
	Name    string
	Period  time.Duration
	Checker checker.Checker
}
