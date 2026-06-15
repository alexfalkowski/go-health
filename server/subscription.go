package server

import (
	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-health/v2/watcher"
	"github.com/alexfalkowski/go-sync"
)

type subscription struct {
	updates  chan error
	server   *Server
	kind     string
	watchers []watcher.Subscription
	wg       sync.WaitGroup
	once     sync.Once
}

func (s *subscription) Receive() <-chan error {
	return s.updates
}

func (s *subscription) Close() {
	s.once.Do(func() {
		for _, sub := range s.watchers {
			sub.Close()
		}

		s.wg.Wait()
		close(s.updates)
	})
}

// start publishes the current aggregate state and begins relaying observer updates.
//
// Each observer update means some underlying probe tick was processed. The
// watcher responds by recomputing the server-level error and publishing that
// snapshot to its channel.
func (s *subscription) start() {
	s.publish(s.snapshot())

	for _, observer := range s.server.observers(s.kind) {
		s.watch(observer)
	}
}

// watch relays one observer's updates into the aggregate channel.
func (s *subscription) watch(observer *subscriber.Observer) {
	sub := observer.Watch()
	s.watchers = append(s.watchers, sub)

	s.wg.Go(func() {
		for range sub.Receive() {
			s.publish(s.snapshot())
		}
	})
}

func (s *subscription) snapshot() error {
	return s.server.Error(s.kind)
}

func (s *subscription) publish(err error) {
	if s.send(err) {
		return
	}

	s.drop()
	_ = s.send(err)
}

func (s *subscription) send(err error) bool {
	select {
	case s.updates <- err:
		return true
	default:
		return false
	}
}

func (s *subscription) drop() {
	select {
	case <-s.updates:
	default:
	}
}
