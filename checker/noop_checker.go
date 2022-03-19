package checker

import (
	"context"
)

// NewNoopChecker with no functionality.
func NewNoopChecker() *NoopChecker {
	return &NoopChecker{}
}

// NoopChecker with no functionality.
type NoopChecker struct {
}

// Check does a NOOP.
func (c *NoopChecker) Check(ctx context.Context) error {
	return nil
}
