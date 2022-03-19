package subscriber

import (
	"sync"
)

// NewObserver from probe names and subscriber.
func NewObserver(names []string, sub *Subscriber) *Observer {
	values := make(map[string]error)
	for _, n := range names {
		values[n] = nil
	}

	ob := &Observer{values: values, sub: sub, mux: sync.Mutex{}}

	go ob.observe()

	return ob
}

// Observer represents a subscriber that mantaines state about probes.
type Observer struct {
	values map[string]error
	sub    *Subscriber
	mux    sync.Mutex
}

// Error is the first error observed.
func (o *Observer) Error() error {
	o.mux.Lock()
	defer o.mux.Unlock()

	for _, err := range o.values {
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Observer) observe() {
	for t := range o.sub.Receive() {
		o.mux.Lock()
		o.values[t.Name()] = t.Error()
		o.mux.Unlock()
	}
}
