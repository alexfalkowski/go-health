package subscriber

import (
	"slices"

	"github.com/alexfalkowski/go-sync"
)

// NewObserver returns an Observer that tracks the latest errors for names.
func NewObserver(names []string, sub *Subscriber) *Observer {
	names = slices.Clone(names)
	errors := make(Errors)
	for _, n := range names {
		errors[n] = nil
	}

	ob := &Observer{errors: errors, names: names, mux: sync.RWMutex{}}
	ob.start(sub)

	return ob
}

// Observer represents a subscriber that maintains state about probes.
type Observer struct {
	errors Errors
	names  []string
	mux    sync.RWMutex
	wg     sync.WaitGroup
}

// Error returns all non-nil errors combined into a single error.
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

// Restart waits for the current subscriber to finish and starts observing sub.
func (o *Observer) Restart(sub *Subscriber) {
	o.wg.Wait()
	o.start(sub)
}

// Wait blocks until the current observe loop exits.
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
		o.mux.Unlock()
	}
}
