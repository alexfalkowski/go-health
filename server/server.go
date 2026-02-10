// Package server orchestrates probes and observers for multiple services.
package server

import (
	"errors"
	"iter"
	"maps"
	"sync"

	"github.com/alexfalkowski/go-health/v2/subscriber"
)

// ErrServiceNotFound when the service has not been registered.
var ErrServiceNotFound = errors.New("health: service not found")

// NewServer returns a Server.
func NewServer() *Server {
	return &Server{services: make(map[string]*Service), mux: sync.Mutex{}}
}

// Server maintains registered services and can start/stop them.
type Server struct {
	services map[string]*Service
	mux      sync.Mutex
	running  bool
}

// Register adds a service with name and registrations.
func (s *Server) Register(name string, regs ...*Registration) {
	service := NewService()
	service.Register(regs...)
	s.services[name] = service
}

// Observers returns all observers of the given kind.
func (s *Server) Observers(kind string) iter.Seq2[string, *subscriber.Observer] {
	return func(yield func(string, *subscriber.Observer) bool) {
		for name, service := range maps.All(s.services) {
			observer, err := service.Observer(kind)
			if err != nil {
				continue
			}

			if !yield(name, observer) {
				return
			}
		}
	}
}

// Observer returns an observer for the service name and observer kind.
func (s *Server) Observer(name, kind string) (*subscriber.Observer, error) {
	service, ok := s.services[name]
	if !ok {
		return nil, ErrServiceNotFound
	}

	return service.Observer(kind)
}

// Observe registers an observer kind for the service name.
//
// The observer will track the probes listed in names.
func (s *Server) Observe(name, kind string, names ...string) error {
	service, ok := s.services[name]
	if !ok {
		return ErrServiceNotFound
	}

	service.Observe(kind, names...)
	return nil
}

// Start starts all registered services.
func (s *Server) Start() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.running {
		return
	}

	for _, service := range s.services {
		service.Start()
	}
	s.running = true
}

// Stop stops all registered services.
func (s *Server) Stop() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.running {
		return
	}

	for _, service := range s.services {
		service.Stop()
	}
	s.running = false
}
