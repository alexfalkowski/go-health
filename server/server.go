package server

import (
	"context"
	"errors"
	"iter"

	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-sync"
)

// ErrServiceNotFound is returned when a service name has not been registered.
var ErrServiceNotFound = errors.New("health: service not found")

// NewServer returns a Server with no registered services.
func NewServer() *Server {
	return &Server{services: make(map[string]*Service), mux: sync.Mutex{}}
}

// Server maintains registered services and can start or stop them as a group.
//
// Register and Observe are setup-time calls. Call Start once the server has
// been configured, and Stop after Start has returned when the process is shutting
// down.
type Server struct {
	services map[string]*Service
	mux      sync.Mutex
	running  bool
}

// Register adds a service with name and registrations.
//
// Register is intended for setup before Start. If a service already exists for
// name, it is replaced.
func (s *Server) Register(name string, regs ...*Registration) {
	service := NewService()
	service.Register(regs...)
	s.services[name] = service
}

// Observers returns all observers of the given kind.
//
// Services that do not have that observer kind are skipped.
func (s *Server) Observers(kind string) iter.Seq2[string, *subscriber.Observer] {
	return func(yield func(string, *subscriber.Observer) bool) {
		for name, service := range s.services {
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
// Observe is intended for setup before Start. The observer will track the probes
// listed in names. Repeated calls with the same service and kind are idempotent.
func (s *Server) Observe(name, kind string, names ...string) error {
	service, ok := s.services[name]
	if !ok {
		return ErrServiceNotFound
	}

	return service.Observe(kind, names...)
}

// Start starts all registered services.
//
// Start is idempotent. It waits for each service's initial checks before
// returning; call Stop after Start has returned during normal shutdown.
func (s *Server) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if s.running {
		return nil
	}

	started := make([]*Service, 0, len(s.services))
	for _, service := range s.services {
		if err := service.Start(ctx); err != nil {
			if stopErr := stopServices(context.WithoutCancel(ctx), started...); stopErr != nil {
				return errors.Join(err, stopErr)
			}
			return err
		}
		started = append(started, service)
	}
	s.running = true
	return nil
}

// Stop stops all registered services.
//
// Stop is idempotent.
func (s *Server) Stop(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.running {
		return nil
	}

	services := make([]*Service, 0, len(s.services))
	for _, service := range s.services {
		services = append(services, service)
	}
	if err := stopServices(ctx, services...); err != nil {
		return err
	}
	s.running = false
	return nil
}

func stopServices(ctx context.Context, services ...*Service) error {
	var g sync.ErrorsGroup
	for _, service := range services {
		g.Go(func() error {
			return service.Stop(ctx)
		})
	}

	return g.Wait()
}
