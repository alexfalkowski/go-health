package subscriber

import (
	"sync"
)

// NewObserver from probe names and subscriber.
func NewObserver(names []string, sub *Subscriber) *Observer {
	errors := make(Errors)
	for _, n := range names {
		errors[n] = nil
	}

	ob := &Observer{errors: errors, sub: sub, mux: sync.Mutex{}}

	go ob.observe()

	return ob
}

// Observer represents a subscriber that mantaines state about probes.
type Observer struct {
	errors Errors
	sub    *Subscriber
	mux    sync.Mutex
}

// Error is the first error observed.
func (o *Observer) Error() error {
	o.mux.Lock()
	defer o.mux.Unlock()

	return o.errors.Error()
}

func (o *Observer) observe() {
	for t := range o.sub.Receive() {
		o.mux.Lock()
		o.errors.Set(t.Name(), t.Error())
		o.mux.Unlock()
	}
}
