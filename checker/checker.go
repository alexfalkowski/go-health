package checker

import "context"

// Checker reports the health of a dependency or subsystem.
//
// Implementations should return nil when healthy and a non-nil error when the
// dependency is unhealthy. Callers provide the context so they can propagate
// cancellation, deadlines, and request-scoped values into the check.
type Checker interface {
	Check(ctx context.Context) error
}
