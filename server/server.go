package server

import (
	"sync"

	"github.com/alexfalkowski/go-health/probe"
	"github.com/alexfalkowski/go-health/subscriber"
)

type status string

const (
	started = status("started")
	stopped = status("stopped")
)

// NewServer for health.
func NewServer() *Server {
	registry := make(map[string]*probe.Probe)
	subscribers := []*subscriber.Subscriber{}

	return &Server{registry: registry, subscribers: subscribers, done: nil, ticks: nil, wg: nil, mux: sync.Mutex{}, st: ""}
}

// Server will maintain all the probes and start and stop them.
type Server struct {
	registry    map[string]*probe.Probe
	subscribers []*subscriber.Subscriber
	done        chan struct{}
	ticks       chan *probe.Tick
	wg          *sync.WaitGroup
	mux         sync.Mutex
	st          status
}

// Register all the registrations.
func (s *Server) Register(regs ...*Registration) {
	for _, reg := range regs {
		s.registry[reg.Name] = probe.NewProbe(reg.Name, reg.Period, reg.Checker)
	}
}

// Subscribe to the names of the probes.
func (s *Server) Subscribe(names ...string) *subscriber.Subscriber {
	sub := subscriber.NewSubscriber(names)

	s.subscribers = append(s.subscribers, sub)

	return sub
}

// Observe the names of the probes.
func (s *Server) Observe(names ...string) *subscriber.Observer {
	return subscriber.NewObserver(names, s.Subscribe(names...))
}

// Start the server.
func (s *Server) Start() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.st == started {
		return
	}

	s.st = started
	s.wg = &sync.WaitGroup{}
	s.ticks = make(chan *probe.Tick, 1)
	s.done = make(chan struct{}, 1)
	chs := []<-chan *probe.Tick{}

	for _, p := range s.registry {
		s.wg.Add(1)

		chs = append(chs, p.Start())
	}

	s.mergeChannels(chs)

	go s.sendToSubscribers()
}

// Stop the server.
func (s *Server) Stop() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.st == "" || s.st == stopped {
		return
	}

	s.st = stopped
	close(s.done)
	s.wg.Wait()

	for _, p := range s.registry {
		p.Stop()
	}

	close(s.ticks)
}

func (s *Server) mergeChannels(chs []<-chan *probe.Tick) {
	for _, ch := range chs {
		go s.sendTick(ch)
	}
}

func (s *Server) sendTick(ch <-chan *probe.Tick) {
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
