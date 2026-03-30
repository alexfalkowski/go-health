package server

import (
	"errors"
	"fmt"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-sync"
)

// ErrObserverNotFound when the observer has not been registered.
var ErrObserverNotFound = errors.New("health: observer not found")

// ErrProbeNotFound when the probe has not been registered.
var ErrProbeNotFound = errors.New("health: probe not found")

// NewService returns a Service.
func NewService() *Service {
	return &Service{
		registry:      make(map[string]*probe.Probe),
		observers:     make(map[string]*subscriber.Observer),
		subscriptions: make(map[string]*subscriber.Subscriber),
		subscribers:   []*subscriber.Subscriber{},
		registryWG:    &sync.WaitGroup{},
		subscriberWG:  &sync.WaitGroup{},
		ticks:         make(chan *probe.Tick, 1),
		done:          make(chan struct{}),
	}
}

// Service maintains probes, subscribers and observers for a service.
type Service struct {
	registry      map[string]*probe.Probe
	observers     map[string]*subscriber.Observer
	subscriptions map[string]*subscriber.Subscriber
	done          chan struct{}
	ticks         chan *probe.Tick
	registryWG    *sync.WaitGroup
	subscriberWG  *sync.WaitGroup
	subscribers   []*subscriber.Subscriber
	mux           sync.Mutex
	running       bool
}

// Register registers all the given probe registrations.
func (s *Service) Register(regs ...*Registration) {
	for _, reg := range regs {
		s.registry[reg.Name] = probe.NewProbe(reg.Name, reg.Period, reg.Checker)
	}
}

// Observer returns the observer for kind.
func (s *Service) Observer(kind string) (*subscriber.Observer, error) {
	observer, ok := s.observers[kind]
	if !ok {
		return nil, ErrObserverNotFound
	}

	return observer, nil
}

// Observe registers an observer kind that tracks the probes listed in names.
//
// It returns an error if any probe name has not been registered.
func (s *Service) Observe(kind string, names ...string) error {
	_, ok := s.observers[kind]
	if !ok {
		if err := s.validateProbeNames(names...); err != nil {
			return err
		}

		s.observers[kind] = subscriber.NewObserver(names, s.subscribe(kind, names...))
	}

	return nil
}

// Start starts all registered probes and begins fan-out to subscribers.
func (s *Service) Start() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.running {
		return
	}

	s.prepareStart()

	chs := []<-chan *probe.Tick{}
	for _, p := range s.registry {
		chs = append(chs, p.Start())
	}

	s.mergeChannels(chs)
	s.subscriberWG.Go(func() {
		s.sendToSubscribers()
	})

	s.running = true
}

// Stop stops all probes and closes all subscribers.
func (s *Service) Stop() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.running {
		return
	}

	// Stop probes first so their tick channels can close and sendTick goroutines can exit cleanly.
	for _, p := range s.registry {
		p.Stop()
	}

	// Signal all sendTick goroutines to stop and wait for them to finish.
	close(s.done)
	s.registryWG.Wait()

	// Now it is safe to close the fan-in channel: no further sends can occur.
	close(s.ticks)

	// Wait for the fan-out goroutine to drain/exit and close all subscribers.
	s.subscriberWG.Wait()

	// Ensure observers have finished draining their subscriber channels before a restart.
	for _, observer := range s.observers {
		observer.Wait()
	}

	s.running = false
}

func (s *Service) subscribe(kind string, names ...string) *subscriber.Subscriber {
	sub := subscriber.NewSubscriber(names)
	s.subscriptions[kind] = sub
	s.subscribers = append(s.subscribers, sub)
	return sub
}

func (s *Service) mergeChannels(chs []<-chan *probe.Tick) {
	for _, ch := range chs {
		s.registryWG.Go(func() {
			s.sendTick(ch)
		})
	}
}

func (s *Service) prepareStart() {
	s.done = make(chan struct{})
	s.ticks = make(chan *probe.Tick, 1)
	s.registryWG = &sync.WaitGroup{}
	s.subscriberWG = &sync.WaitGroup{}

	s.subscribers = make([]*subscriber.Subscriber, 0, len(s.observers))
	for kind, observer := range s.observers {
		sub, ok := s.subscriptions[kind]
		if !ok || sub.Closed() {
			sub = subscriber.NewSubscriber(observer.Names())
			s.subscriptions[kind] = sub
			observer.Restart(sub)
		}

		s.subscribers = append(s.subscribers, sub)
	}
}

func (s *Service) validateProbeNames(names ...string) error {
	errs := make([]error, 0, len(names))
	for _, name := range names {
		if _, ok := s.registry[name]; !ok {
			errs = append(errs, fmt.Errorf("%w: %s", ErrProbeNotFound, name))
		}
	}
	return errors.Join(errs...)
}

func (s *Service) sendTick(ch <-chan *probe.Tick) {
	for {
		select {
		case <-s.done:
			return
		case t, ok := <-ch:
			if !ok {
				return
			}

			// Avoid sending after shutdown starts.
			select {
			case <-s.done:
				return
			case s.ticks <- t:
			}
		}
	}
}

func (s *Service) sendToSubscribers() {
	for t := range s.ticks {
		for _, sub := range s.subscribers {
			sub.Send(t)
		}
	}

	for _, sub := range s.subscribers {
		sub.Close()
	}
}
