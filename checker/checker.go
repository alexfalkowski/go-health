package checker

import (
	"context"
)

// Checker will check a system.
type Checker interface {
	Check(ctx context.Context) error
}
