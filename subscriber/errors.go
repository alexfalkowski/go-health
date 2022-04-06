package subscriber

// Errors for observers.
type Errors map[string]error

// Error the first error from map.
func (e Errors) Error() error {
	for _, err := range e {
		if err != nil {
			return err
		}
	}

	return nil
}

// Errors is a copy.
func (e Errors) Errors() Errors {
	errors := make(Errors)
	for k, v := range e {
		errors[k] = v
	}

	return errors
}

// Set the error at the name.
func (e Errors) Set(name string, err error) {
	e[name] = err
}
