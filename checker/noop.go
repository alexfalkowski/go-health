package checker

import "context"

var _ Checker = (*NoopChecker)(nil)

// NewNoopChecker returns a Checker that always reports healthy.
func NewNoopChecker() *NoopChecker {
	return &NoopChecker{}
}

// NoopChecker is a Checker that always returns nil.
type NoopChecker struct{}

// Check always returns nil.
func (c *NoopChecker) Check(_ context.Context) error {
	return nil
}
