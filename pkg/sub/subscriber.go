package sub

// NewSubscriber for multiple probes.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names, make(chan error)}
}

// Subscriber subscribes to multiple probes.
type Subscriber struct {
	names []string
	ch    chan error
}

// Receive from the subscriber.
func (s *Subscriber) Receive() <-chan error {
	return s.ch
}

// Send err to subscriber.
func (s *Subscriber) Send(err error) {
	s.ch <- err
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
