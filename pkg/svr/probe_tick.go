package svr

// ProbeTick defines a tick in time for a probe.
type ProbeTick struct {
	name string
	err  error
}

// NewProbeTick with name and error.
func NewProbeTick(name string, err error) *ProbeTick {
	return &ProbeTick{name, err}
}

// Name of the probe tick.
func (p *ProbeTick) Name() string {
	return p.name
}

// Error of the probe tick.
func (p *ProbeTick) Error() error {
	return p.err
}
