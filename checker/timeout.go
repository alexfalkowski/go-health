package checker

import (
	"fmt"
	"time"

	"github.com/alexfalkowski/go-sync"
)

// ErrTimeout is the timeout cause used by derived checker contexts.
//
// It wraps [sync.ErrTimeout], so [errors.Is] also matches
// [context.DeadlineExceeded].
var ErrTimeout = fmt.Errorf("checker: %w", sync.ErrTimeout)

func timeout(t time.Duration) time.Duration {
	if t == 0 {
		t = 30 * time.Second
	}

	return t
}
