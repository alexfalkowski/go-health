package svr

import (
	"errors"
	"sync"
	"time"

	"github.com/alexfalkowski/go-health/pkg/chk"
	"github.com/alexfalkowski/go-health/pkg/prb"
	"github.com/alexfalkowski/go-health/pkg/sub"
)

var (
	// ErrProbeExists when we register the same probe.
	ErrProbeExists = errors.New("probe exists")

	// ErrNoRegistrations have been added.
	ErrNoRegistrations = errors.New("no registrations")
)

// NewServer for health.
func NewServer() *Server {
	registry := make(map[string]*prb.Probe)
	subscribers := []*sub.Subscriber{}

	return &Server{registry: registry, subscribers: subscribers}
}

// Server will maintain all the probes and start and stop them.
type Server struct {
	registry    map[string]*prb.Probe
	subscribers []*sub.Subscriber
	done        chan struct{}
	ticks       chan *prb.Tick
	wg          *sync.WaitGroup
}

// Register a checker.
func (s *Server) Register(name string, period time.Duration, checker chk.Checker) error {
	if _, ok := s.registry[name]; ok {
		return ErrProbeExists
	}

	s.registry[name] = prb.NewProbe(name, period, checker)

	return nil
}

// Subscribe to the names of the probes.
func (s *Server) Subscribe(names ...string) *sub.Subscriber {
	sub := sub.NewSubscriber(names)

	s.subscribers = append(s.subscribers, sub)

	return sub
}

// Observe the names of the probes.
func (s *Server) Observe(names ...string) *sub.Observer {
	o := sub.NewObserver(names, s.Subscribe(names...))

	o.Observe()

	return o
}

// Start the server.
func (s *Server) Start() error {
	if len(s.registry) == 0 {
		return ErrNoRegistrations
	}

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
	if len(s.registry) == 0 {
		return ErrNoRegistrations
	}

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
