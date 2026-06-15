package subscriber

import (
	"slices"

	"github.com/alexfalkowski/go-sync"
)

// NewObserver returns an Observer that tracks the latest errors for names.
//
// The observer starts consuming the subscriber immediately in a background
// goroutine. The names slice is cloned and used to seed the error map with nil
// values for all tracked probes, so the observer appears healthy until matching
// ticks arrive. The observer trusts sub for tick filtering; pass a subscriber
// configured with the same names when Names and Errors should stay aligned.
func NewObserver(names []string, sub *Subscriber) *Observer {
	names = slices.Clone(names)
	errs := make(Errors)
	for _, n := range names {
		errs[n] = nil
	}

	ob := &Observer{
		errors:   errs,
		names:    names,
		watchers: make(map[*Watcher]struct{}),
		mux:      sync.RWMutex{},
	}
	ob.start(sub)

	return ob
}

// Observer maintains the latest health state for probe ticks it receives.
type Observer struct {
	errors   Errors
	watchers map[*Watcher]struct{}
	names    []string
	wg       sync.WaitGroup
	mux      sync.RWMutex
}

// Error returns all non-nil errors combined into a single error.
//
// Each individual error is annotated with the probe name before being joined.
func (o *Observer) Error() error {
	o.mux.RLock()
	defer o.mux.RUnlock()

	return o.errors.Error()
}

// Errors returns a copy of the current error map.
func (o *Observer) Errors() Errors {
	o.mux.RLock()
	defer o.mux.RUnlock()

	return o.errors.Errors()
}

// Names returns the probe names tracked by the observer.
func (o *Observer) Names() []string {
	return slices.Clone(o.names)
}

// Watch returns a watcher for current and future observer errors.
//
// The watcher receives the observer's current error immediately, then receives
// the current error again after each matching probe tick is processed. Sends are
// best-effort and coalesced to the latest error when the receiver is slow. Close
// the watcher when the receiver no longer needs updates.
func (o *Observer) Watch() *Watcher {
	watcher := NewWatcher(o.close)

	o.mux.Lock()
	o.watchers[watcher] = struct{}{}
	watcher.publish(o.errors.Error())
	o.mux.Unlock()

	return watcher
}

// Restart waits for the current subscriber to finish and starts observing sub.
//
// Restart does not close the current subscriber. Callers must close it first so
// the current observe loop can exit; the replacement subscriber starts only
// after that channel closes. Existing error state is retained until new ticks
// arrive.
func (o *Observer) Restart(sub *Subscriber) {
	o.wg.Wait()
	o.start(sub)
}

// Wait blocks until the current observe loop exits.
//
// Direct callers should close the current subscriber before waiting; otherwise
// Wait can block indefinitely.
func (o *Observer) Wait() {
	o.wg.Wait()
}

func (o *Observer) start(sub *Subscriber) {
	o.wg.Go(func() {
		o.observe(sub)
	})
}

func (o *Observer) observe(sub *Subscriber) {
	for t := range sub.Receive() {
		o.mux.Lock()
		o.errors.Set(t.Name(), t.Error())
		err := o.errors.Error()
		o.send(err)
		o.mux.Unlock()
	}
}

func (o *Observer) send(err error) {
	for watcher := range o.watchers {
		watcher.publish(err)
	}
}

func (o *Observer) close(watcher *Watcher) {
	o.mux.Lock()
	delete(o.watchers, watcher)
	o.mux.Unlock()
}
