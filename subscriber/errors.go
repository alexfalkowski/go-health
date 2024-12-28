package subscriber

import (
	"errors"
	"fmt"
)

// Errors for observers.
type Errors map[string]error

// Error the combined errors as one.
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
	for k, v := range e {
		errs[k] = v
	}

	return errs
}

// Set the error at the name.
func (e Errors) Set(name string, err error) {
	e[name] = err
}
