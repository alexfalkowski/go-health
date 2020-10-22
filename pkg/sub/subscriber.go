package sub

import (
	"fmt"

	"github.com/alexfalkowski/go-health/pkg/prb"
)

// NewSubscriber for multiple probes.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names, make(chan *prb.Tick, 1)}
}

// Subscriber subscribes to multiple probes.
type Subscriber struct {
	names []string
	ticks chan *prb.Tick
}

// Receive from the subscriber.
func (s *Subscriber) Receive() <-chan *prb.Tick {
	return s.ticks
}

// Send tick to subscriber.
func (s *Subscriber) Send(tick *prb.Tick) {
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
