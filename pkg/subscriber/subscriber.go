package subscriber

import (
	"fmt"

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
	if !s.hasName(tick.Name()) {
		return
	}

	s.ticks <- tick
}

func (s *Subscriber) String() string {
	return fmt.Sprintf("names: '%s'", s.names)
}

func (s *Subscriber) hasName(name string) bool {
	for _, n := range s.names {
		if name == n {
			return true
		}
	}

	return false
}
