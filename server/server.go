package server

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-health/v2/watcher"
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
// Use NewServer to create a Server; the zero value is not initialized for
// registration.
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
// Services that do not have that observer kind are skipped. Iteration order is
// unspecified.
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

// Error returns all non-nil observer errors for kind joined into one error.
//
// Services without kind are skipped. Each service error is annotated with the
// service name before being joined. It returns ErrObserverNotFound when no
// service has registered kind. A nil result means every registered observer for
// kind is currently healthy.
func (s *Server) Error(kind string) error {
	if !s.hasObserver(kind) {
		return ErrObserverNotFound
	}

	errs := make([]error, 0)
	for name, observer := range s.Observers(kind) {
		if err := observer.Error(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}

	return errors.Join(errs...)
}

// Watch returns a watcher for current and future aggregate observer errors for kind.
//
// Services without kind are skipped. It returns ErrObserverNotFound when no
// service has registered kind. The watcher receives the current aggregate error
// immediately, then receives the current aggregate error again after any observer
// of kind receives a probe tick. Sends are best-effort and coalesced to the
// latest error when the receiver is slow. Close the watcher when the receiver no
// longer needs updates; stopping the server does not close existing watchers.
// Register and Observe remain setup-time calls and are not added to an existing
// watch.
func (s *Server) Watch(kind string) (watcher.Subscription, error) {
	if !s.hasObserver(kind) {
		return nil, ErrObserverNotFound
	}

	sub := &subscription{
		updates: make(chan error, 1),
		server:  s,
		kind:    kind,
	}
	sub.start()

	return sub, nil
}

func (s *Server) hasObserver(kind string) bool {
	for range s.Observers(kind) {
		return true
	}

	return false
}

// Observer returns an observer for the service name and observer kind.
//
// It returns ErrServiceNotFound if name has not been registered, or
// ErrObserverNotFound if the service does not have kind.
func (s *Server) Observer(name, kind string) (*subscriber.Observer, error) {
	service, ok := s.services[name]
	if !ok {
		return nil, ErrServiceNotFound
	}

	return service.Observer(kind)
}

// Observe registers an observer kind for the service name.
//
// Observe is intended for setup before Start. When creating a new observer kind,
// the observer will track the probes listed in names. Repeated calls with the
// same service and kind are idempotent, keep the original probe set, and do not
// validate the new names. It returns ErrServiceNotFound if name has not been
// registered, or ErrProbeNotFound if any probe name is unknown for a new
// observer kind.
func (s *Server) Observe(name, kind string, names ...string) error {
	service, ok := s.services[name]
	if !ok {
		return ErrServiceNotFound
	}

	return service.Observe(kind, names...)
}

// Start starts all registered services.
//
// Start is intended to be called once after setup. It waits for each service's
// initial checks before returning; call Stop once after Start has returned
// during normal shutdown. If a service fails to start, Start stops any services
// started during that attempt and leaves the server stopped. Calling Start
// again with a live context while the server is already running is a no-op that
// returns nil.
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
// Stop is intended to be called once after Start returns. Use a context that can
// remain valid until cleanup completes; if ctx expires before shutdown work
// finishes, Stop returns ctx.Err(). Calling Stop before Start or after a prior
// Stop is a no-op and returns nil.
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
