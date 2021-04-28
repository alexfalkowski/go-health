package subscriber

import (
	"github.com/alexfalkowski/go-health/pkg/probe"
)

// NewSubscriber for multiple probes.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names: names, ticks: make(chan *probe.Tick, 1)}
}

// Subscriber subscribes to multiple probes.
type Subscriber struct {
	names []string
	ticks chan *probe.Tick
}

// Receive from the subscriber.
func (s *Subscriber) Receive() <-chan *probe.Tick {
	return s.ticks
}

// Send tick to subscriber.
func (s *Subscriber) Send(tick *probe.Tick) {
	s.ticks <- tick
}
