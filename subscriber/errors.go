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

// Set the error at the name.
func (e Errors) Set(name string, err error) {
	e[name] = err
}
