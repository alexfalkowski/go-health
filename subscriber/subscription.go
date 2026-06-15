package subscriber

import "github.com/alexfalkowski/go-sync"

type subscription struct {
	updates chan error
	done    chan struct{}
	mux     sync.Mutex
	once    sync.Once
}

func (s *subscription) Receive() <-chan error {
	return s.updates
}

func (s *subscription) Close() {
	s.once.Do(func() {
		s.mux.Lock()
		defer s.mux.Unlock()

		close(s.done)
		close(s.updates)
	})
}

func (s *subscription) publish(err error) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	select {
	case <-s.done:
		return false
	default:
	}

	if s.send(err) {
		return true
	}

	s.drop()
	_ = s.send(err)

	return true
}

func (s *subscription) closed() <-chan struct{} {
	return s.done
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
