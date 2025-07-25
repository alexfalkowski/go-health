package server

import (
	"sync"

	"github.com/alexfalkowski/go-health/v2/subscriber"
)

// NewServer for health.
func NewServer() *Server {
	return &Server{services: make(map[string]*Service), mux: sync.Mutex{}}
}

// Server will maintain all the services and start and stop them.
type Server struct {
	services map[string]*Service
	status   Status
	mux      sync.Mutex
}

// Register a service with name and registrations.
func (s *Server) Register(name string, regs ...*Registration) {
	service := NewService()
	service.Register(regs...)
	s.services[name] = service
}

// Observer from the service with name and kind of observer.
func (s *Server) Observer(name, kind string) *subscriber.Observer {
	return s.services[name].Observer(kind)
}

// Observe a service with name and kind of observer with names of the probes.
func (s *Server) Observe(name, kind string, names ...string) *subscriber.Observer {
	service := s.services[name]
	return service.Observe(kind, names...)
}

// Start the server.
func (s *Server) Start() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.status.IsStarted() {
		return
	}

	s.status = Started
	for _, service := range s.services {
		service.Start()
	}
}

// Stop the server.
func (s *Server) Stop() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.status.IsEmpty() || s.status.IsStopped() {
		return
	}

	s.status = Stopped
	for _, service := range s.services {
		service.Stop()
	}
}
