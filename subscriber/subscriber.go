package subscriber

import "github.com/alexfalkowski/go-health/probe"

// NewSubscriber for multiple probes.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names: names, ticks: make(chan *probe.Tick, 1)}
}

// Subscriber subscribes to multiple probes.
type Subscriber struct {
	ticks chan *probe.Tick
	names []string
}

// Receive from the subscriber.
func (s *Subscriber) Receive() <-chan *probe.Tick {
	return s.ticks
}

// Send tick to subscriber.
func (s *Subscriber) Send(tick *probe.Tick) {
	for _, n := range s.names {
		if n == tick.Name() {
			s.ticks <- tick
		}
	}
}
