package subscriber

import (
	"slices"
	"sync"
	"sync/atomic"

	"github.com/alexfalkowski/go-health/v2/probe"
)

// NewSubscriber for multiple probes.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names: names, ticks: make(chan *probe.Tick, 1)}
}

// Subscriber subscribes to multiple probes.
type Subscriber struct {
	ticks  chan *probe.Tick
	names  []string
	closed atomic.Bool
	once   sync.Once
}

// Receive from the subscriber.
func (s *Subscriber) Receive() <-chan *probe.Tick {
	return s.ticks
}

// Send tick to subscriber.
//
// This is hardened to:
// - be non-blocking (drops ticks if the subscriber is not keeping up)
// - avoid panics when Close races with Send.
func (s *Subscriber) Send(tick *probe.Tick) {
	if s.closed.Load() {
		return
	}

	if slices.Contains(s.names, tick.Name()) {
		// Best-effort delivery. If the buffer is full, drop.
		select {
		case s.ticks <- tick:
		default:
		}
		return
	}
}

// Close the subscriber.
//
// Close is idempotent and safe to call concurrently.
func (s *Subscriber) Close() {
	s.once.Do(func() {
		s.closed.Store(true)
		close(s.ticks)
	})
}
