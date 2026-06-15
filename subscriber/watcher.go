package subscriber

import "github.com/alexfalkowski/go-sync"

// NewWatcher returns a Watcher.
//
// The watcher uses a one-slot channel and coalesces slow receivers to the latest
// pending error. Close calls cleanup before closing the receive channel. Passing
// nil for cleanup creates a watcher that only closes its receive channel.
func NewWatcher(cleanup func(*Watcher)) *Watcher {
	return &Watcher{updates: make(chan error, 1), close: cleanup}
}

// Watcher receives current and future observer errors.
//
// A Watcher owns one observer subscription. Call Close when the receiver no
// longer needs updates; Close removes the watcher from the observer and closes
// the receive channel.
type Watcher struct {
	updates chan error
	close   func(*Watcher)
	once    sync.Once
}

// Receive returns the channel of error snapshots.
func (w *Watcher) Receive() <-chan error {
	return w.updates
}

// Close closes the watcher and its receive channel.
func (w *Watcher) Close() {
	w.once.Do(func() {
		if w.close != nil {
			w.close(w)
		}

		close(w.updates)
	})
}

func (w *Watcher) publish(err error) {
	if w.send(err) {
		return
	}

	w.drop()
	_ = w.send(err)
}

func (w *Watcher) send(err error) bool {
	select {
	case w.updates <- err:
		return true
	default:
		return false
	}
}

func (w *Watcher) drop() {
	select {
	case <-w.updates:
	default:
	}
}
