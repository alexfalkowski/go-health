package checker

import "context"

var _ Checker = (*NoopChecker)(nil)

// NewNoopChecker returns a Checker that always reports healthy.
//
// It is useful in tests, examples, or when a health group needs a placeholder
// dependency that should never fail.
func NewNoopChecker() *NoopChecker {
	return &NoopChecker{}
}

// NoopChecker is a Checker that always returns nil.
type NoopChecker struct{}

// Check always returns nil unless ctx is canceled.
func (c *NoopChecker) Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}
