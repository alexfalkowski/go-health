package subscriber

import (
	"errors"
	"fmt"
	"maps"
)

// Errors stores the latest error for each tracked probe name.
type Errors map[string]error

// Error returns all non-nil errors combined into a single error.
//
// Each non-nil error is wrapped with its probe name before being joined.
func (e Errors) Error() error {
	errs := make([]error, 0, len(e))
	for k, err := range e {
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", k, err))
		}
	}

	return errors.Join(errs...)
}

// Errors returns a shallow copy of the map.
func (e Errors) Errors() Errors {
	errs := make(Errors, len(e))
	maps.Copy(errs, e)
	return errs
}

// Set stores err as the latest error for name.
func (e Errors) Set(name string, err error) {
	e[name] = err
}
