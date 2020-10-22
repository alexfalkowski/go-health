package chk

import (
	"context"
	"fmt"
)

// Checker will check a system.
type Checker interface {
	fmt.Stringer

	Check(ctx context.Context) error
}
