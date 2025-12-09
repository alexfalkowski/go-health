package probe

// NewTick with name and error.
func NewTick(name string, err error) *Tick {
	return &Tick{name: name, err: err}
}

// Tick defines a tick in time for a probe.
type Tick struct {
	err  error
	name string
}

// Name of the probe tick.
func (t *Tick) Name() string {
	return t.name
}

// Error of the probe tick.
func (t *Tick) Error() error {
	return t.err
}
