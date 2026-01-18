package server

import (
	"errors"
	"sync"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/subscriber"
)

// ErrObserverNotFound when the observer has not been registered.
var ErrObserverNotFound = errors.New("health: observer not found")

// NewService for health.
func NewService() *Service {
	return &Service{
		registry:     make(map[string]*probe.Probe),
		observers:    make(map[string]*subscriber.Observer),
		subscribers:  []*subscriber.Subscriber{},
		registryWG:   &sync.WaitGroup{},
		subscriberWG: &sync.WaitGroup{},
		ticks:        make(chan *probe.Tick, 1),
		done:         make(chan struct{}),
	}
}

// Service will maintain all the probes and start and stop them.
type Service struct {
	registry     map[string]*probe.Probe
	observers    map[string]*subscriber.Observer
	done         chan struct{}
	ticks        chan *probe.Tick
	registryWG   *sync.WaitGroup
	subscriberWG *sync.WaitGroup
	subscribers  []*subscriber.Subscriber
}

// Register all the registrations.
func (s *Service) Register(regs ...*Registration) {
	for _, reg := range regs {
		s.registry[reg.Name] = probe.NewProbe(reg.Name, reg.Period, reg.Checker)
	}
}

// Observer for kind.
func (s *Service) Observer(kind string) (*subscriber.Observer, error) {
	observer, ok := s.observers[kind]
	if !ok {
		return nil, ErrObserverNotFound
	}

	return observer, nil
}

// Observe a kind with the names of the probes.
func (s *Service) Observe(kind string, names ...string) {
	_, ok := s.observers[kind]
	if !ok {
		s.observers[kind] = subscriber.NewObserver(names, s.subscribe(names...))
	}
}

// Start the service.
func (s *Service) Start() {
	chs := []<-chan *probe.Tick{}
	for _, p := range s.registry {
		s.registryWG.Add(1)
		chs = append(chs, p.Start())
	}

	s.mergeChannels(chs)
	s.subscriberWG.Go(func() {
		s.sendToSubscribers()
	})
}

// Stop the server.
func (s *Service) Stop() {
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
}

func (s *Service) subscribe(names ...string) *subscriber.Subscriber {
	sub := subscriber.NewSubscriber(names)
	s.subscribers = append(s.subscribers, sub)
	return sub
}

func (s *Service) mergeChannels(chs []<-chan *probe.Tick) {
	for _, ch := range chs {
		go s.sendTick(ch)
	}
}

func (s *Service) sendTick(ch <-chan *probe.Tick) {
	defer s.registryWG.Done()

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
