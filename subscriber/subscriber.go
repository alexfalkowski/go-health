package subscriber

import (
	"sync"

	"github.com/alexfalkowski/go-health/v2/probe"
)

// NewSubscriber for multiple probes.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names: names, ticks: make(chan *probe.Tick, 1)}
}

// Subscriber subscribes to multiple probes.
type Subscriber struct {
	ticks chan *probe.Tick
	names []string
	once  sync.Once
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

// Close closes the underlying tick channel.
//
// It is safe to call Close multiple times.
func (s *Subscriber) Close() {
	s.once.Do(func() {
		close(s.ticks)
	})
}
