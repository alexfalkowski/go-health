package prb

// Tick defines a tick in time for a probe.
type Tick struct {
	name string
	err  error
}

// NewTick with name and error.
func NewTick(name string, err error) *Tick {
	return &Tick{name, err}
}

// Name of the probe tick.
func (t *Tick) Name() string {
	return t.name
}

// Error of the probe tick.
func (t *Tick) Error() error {
	return t.err
}
