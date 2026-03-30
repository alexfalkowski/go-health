package subscriber

import (
	"slices"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-sync"
)

// NewSubscriber returns a Subscriber for the given probe names.
//
// The names slice is cloned so later caller mutations do not change the
// subscription.
func NewSubscriber(names []string) *Subscriber {
	return &Subscriber{names: slices.Clone(names), ticks: make(chan *probe.Tick, 1)}
}

// Subscriber forwards matching probe ticks to interested consumers.
type Subscriber struct {
	ticks  chan *probe.Tick
	names  []string
	closed bool
	mux    sync.RWMutex
	once   sync.Once
}

// Receive returns the tick channel for this subscription.
func (s *Subscriber) Receive() <-chan *probe.Tick {
	return s.ticks
}

// Closed reports whether the subscriber has been closed.
func (s *Subscriber) Closed() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.closed
}

// Send forwards tick to the subscriber if it matches a configured name.
//
// Delivery is best-effort:
//   - it is non-blocking, so ticks are dropped if the subscriber is not keeping up
//   - it avoids panics when Close races with Send
func (s *Subscriber) Send(tick *probe.Tick) {
	if !slices.Contains(s.names, tick.Name()) {
		return
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	if s.closed {
		return
	}

	// Best-effort delivery. If the buffer is full, drop.
	select {
	case s.ticks <- tick:
	default:
	}
}

// Close closes the subscriber.
//
// Close is idempotent and safe to call concurrently.
func (s *Subscriber) Close() {
	s.once.Do(func() {
		s.mux.Lock()
		defer s.mux.Unlock()

		s.closed = true
		close(s.ticks)
	})
}
