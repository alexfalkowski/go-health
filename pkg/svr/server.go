package svr

import (
	"errors"
	"sync"

	"github.com/alexfalkowski/go-health/pkg/prb"
	"github.com/alexfalkowski/go-health/pkg/sub"
)

type status string

const (
	started = status("started")
	stopped = status("stopped")
)

var (
	// ErrProbeExists when we register the same probe.
	ErrProbeExists = errors.New("probe exists")

	// ErrNoRegistrations have been added.
	ErrNoRegistrations = errors.New("no registrations")

	// ErrRegistrationNotFound in register.
	ErrRegistrationNotFound = errors.New("registration not found")
)

// NewServer for health.
func NewServer() *Server {
	registry := make(map[string]*prb.Probe)
	subscribers := []*sub.Subscriber{}

	return &Server{registry: registry, subscribers: subscribers, done: nil, ticks: nil, wg: nil, mux: sync.Mutex{}, st: ""}
}

// Server will maintain all the probes and start and stop them.
type Server struct {
	registry    map[string]*prb.Probe
	subscribers []*sub.Subscriber
	done        chan struct{}
	ticks       chan *prb.Tick
	wg          *sync.WaitGroup
	mux         sync.Mutex
	st          status
}

// Register all the registrations.
func (s *Server) Register(regs ...*Registration) error {
	for _, reg := range regs {
		if _, ok := s.registry[reg.Name]; ok {
			return ErrProbeExists
		}

		period := reg.Period
		if period == 0 {
			period = defaultPeriod
		}

		s.registry[reg.Name] = prb.NewProbe(reg.Name, period, reg.Checker)
	}

	return nil
}

// Subscribe to the names of the probes.
func (s *Server) Subscribe(names ...string) (*sub.Subscriber, error) {
	for _, n := range names {
		if _, ok := s.registry[n]; !ok {
			return nil, ErrRegistrationNotFound
		}
	}

	sub := sub.NewSubscriber(names)

	s.subscribers = append(s.subscribers, sub)

	return sub, nil
}

// Observe the names of the probes.
func (s *Server) Observe(names ...string) (*sub.Observer, error) {
	su, err := s.Subscribe(names...)
	if err != nil {
		return nil, err
	}

	return sub.NewObserver(names, su), nil
}

// Start the server.
func (s *Server) Start() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(s.registry) == 0 {
		return ErrNoRegistrations
	}

	if s.st == started {
		return nil
	}

	s.st = started
	s.wg = &sync.WaitGroup{}
	s.ticks = make(chan *prb.Tick, 1)
	s.done = make(chan struct{}, 1)
	chs := []<-chan *prb.Tick{}

	for _, p := range s.registry {
		s.wg.Add(1)

		chs = append(chs, p.Start())
	}

	s.mergeChannels(chs)

	go s.sendToSubscribers()

	return nil
}

// Stop the server.
func (s *Server) Stop() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(s.registry) == 0 {
		return ErrNoRegistrations
	}

	if s.st == "" || s.st == stopped {
		return nil
	}

	s.st = stopped
	close(s.done)
	s.wg.Wait()

	for _, p := range s.registry {
		p.Stop()
	}

	close(s.ticks)

	return nil
}

func (s *Server) mergeChannels(chs []<-chan *prb.Tick) {
	for _, ch := range chs {
		go s.sendTick(ch)
	}
}

func (s *Server) sendTick(ch <-chan *prb.Tick) {
	defer s.wg.Done()

	for {
		select {
		case <-s.done:
			return
		case t := <-ch:
			s.ticks <- t
		}
	}
}

func (s *Server) sendToSubscribers() {
	for t := range s.ticks {
		for _, sub := range s.subscribers {
			sub.Send(t)
		}
	}
}
