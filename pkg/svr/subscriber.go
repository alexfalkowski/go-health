package svr

// NewSubscriber for multiple probes.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names, make(chan error)}
}

// Subscriber subscribes to multiple probes.
type Subscriber struct {
	names []string
	ch    chan error
}

// Channel of the subscriber.
func (s *Subscriber) Channel() chan error {
	return s.ch
}

// HasName of probe.
func (s *Subscriber) HasName(name string) bool {
	for _, n := range s.names {
		if name == n {
			return true
		}
	}

	return false
}
