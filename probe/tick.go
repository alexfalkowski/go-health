package probe

// NewTick returns a Tick with the given name and error.
func NewTick(name string, err error) *Tick {
	return &Tick{name: name, err: err}
}

// Tick represents the result of a single probe execution.
type Tick struct {
	err  error
	name string
}

// Name returns the probe name.
func (t *Tick) Name() string {
	return t.name
}

// Error returns the probe error.
func (t *Tick) Error() error {
	return t.err
}
