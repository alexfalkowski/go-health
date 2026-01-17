package subscriber

import "sync"

// NewObserver from probe names and subscriber.
func NewObserver(names []string, sub *Subscriber) *Observer {
	errors := make(Errors)
	for _, n := range names {
		errors[n] = nil
	}

	ob := &Observer{errors: errors, sub: sub, mux: sync.RWMutex{}}
	ob.Start()

	return ob
}

// Observer represents a subscriber that maintains state about probes.
type Observer struct {
	errors  Errors
	sub     *Subscriber
	stop    chan struct{}
	done    chan struct{}
	mux     sync.RWMutex
	running bool
}

// Start begins observing ticks from the underlying Subscriber.
//
// It is safe to call Start multiple times.
func (o *Observer) Start() {
	o.mux.Lock()
	defer o.mux.Unlock()

	if o.running {
		return
	}

	o.stop = make(chan struct{})
	o.done = make(chan struct{})
	o.running = true

	go o.observe(o.stop, o.done)
}

// Stop ends observation and waits for the observe goroutine to exit.
//
// It is safe to call Stop multiple times.
func (o *Observer) Stop() {
	o.mux.Lock()
	defer o.mux.Unlock()

	if !o.running {
		return
	}

	close(o.stop)
	done := o.done

	<-done
	o.running = false
}

// Error is the first error observed.
func (o *Observer) Error() error {
	o.mux.RLock()
	defer o.mux.RUnlock()

	return o.errors.Error()
}

// Errors are a copy of rhe errors.
func (o *Observer) Errors() Errors {
	o.mux.RLock()
	defer o.mux.RUnlock()

	return o.errors.Errors()
}

func (o *Observer) observe(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)

	for {
		select {
		case <-stop:
			return
		case t := <-o.sub.Receive():
			o.mux.Lock()
			o.errors.Set(t.Name(), t.Error())
			o.mux.Unlock()
		}
	}
}
