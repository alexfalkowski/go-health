package checker

import "context"

// Checker checks a system.
type Checker interface {
	Check(ctx context.Context) error
}
