package server

import (
	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-sync"
)

// NewWatcher returns a Watcher for observers.
//
// The watcher subscribes to the supplied observers, publishes the current err
// snapshot before returning, and republishes err after any observer receives a
// tick. Passing nil for err makes the watcher publish nil snapshots.
func NewWatcher(observers []*subscriber.Observer, err func() error) *Watcher {
	if err == nil {
		err = func() error { return nil }
	}

	watcher := &Watcher{updates: make(chan error, 1), err: err}
	watcher.start(observers)

	return watcher
}

// Watcher owns one Server.Watch subscription.
//
// It subscribes to observers, recomputes the server-level aggregate error when
// any observer changes, and publishes that snapshot without blocking probe
// delivery.
type Watcher struct {
	updates  chan error
	err      func() error
	watchers []*subscriber.Watcher
	wg       sync.WaitGroup
	once     sync.Once
}

// Receive returns the channel of aggregate error snapshots.
func (w *Watcher) Receive() <-chan error {
	return w.updates
}

// Close closes the watcher and its receive channel.
func (w *Watcher) Close() {
	w.once.Do(func() {
		for _, watcher := range w.watchers {
			watcher.Close()
		}

		w.wg.Wait()
		close(w.updates)
	})
}

// start publishes the current aggregate state and begins relaying observer updates.
//
// Each observer update means some underlying probe tick was processed. The
// watcher responds by recomputing the server-level error and publishing that
// snapshot to its channel.
func (w *Watcher) start(observers []*subscriber.Observer) {
	for _, observer := range observers {
		w.watch(observer)
	}

	w.publish(w.err())
}

// watch relays one observer's updates into the aggregate channel.
func (w *Watcher) watch(observer *subscriber.Observer) {
	watcher := observer.Watch()
	w.watchers = append(w.watchers, watcher)

	w.wg.Go(func() {
		for range watcher.Receive() {
			w.publish(w.err())
		}
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
