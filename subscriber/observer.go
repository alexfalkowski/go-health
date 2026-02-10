package subscriber

import "sync"

// NewObserver returns an Observer that tracks the latest errors for names.
func NewObserver(names []string, sub *Subscriber) *Observer {
	errors := make(Errors)
	for _, n := range names {
		errors[n] = nil
	}

	ob := &Observer{errors: errors, sub: sub, mux: sync.RWMutex{}}
	go ob.observe()

	return ob
}

// Observer represents a subscriber that maintains state about probes.
type Observer struct {
	errors Errors
	sub    *Subscriber
	mux    sync.RWMutex
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

func (o *Observer) observe() {
	for t := range o.sub.Receive() {
		o.mux.Lock()
		o.errors.Set(t.Name(), t.Error())
		o.mux.Unlock()
	}
}
