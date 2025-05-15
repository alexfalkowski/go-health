package checker

import "context"

var _ Checker = (*NoopChecker)(nil)

// NewNoopChecker does nothing.
func NewNoopChecker() *NoopChecker {
	return &NoopChecker{}
}

// NoopChecker does nothing..
type NoopChecker struct{}

// Check does a NOOP.
func (c *NoopChecker) Check(_ context.Context) error {
	return nil
}
