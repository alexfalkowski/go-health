package subscriber

import (
	"errors"
	"fmt"
	"maps"
)

// Errors for observers.
type Errors map[string]error

// Error returns all non-nil errors combined into a single error.
func (e Errors) Error() error {
	errs := make([]error, len(e))
	i := 0

	for k, err := range e {
		if err != nil {
			errs[i] = fmt.Errorf("%s: %w", k, err)
		}

		i++
	}

	return errors.Join(errs...)
}

// Errors is a copy.
func (e Errors) Errors() Errors {
	errs := make(Errors, len(e))
	maps.Copy(errs, e)
	return errs
}

// Set sets the error for name.
func (e Errors) Set(name string, err error) {
	e[name] = err
}
