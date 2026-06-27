package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-health/v2/watcher"
	"github.com/alexfalkowski/go-sync"
)

// ErrObserverNotFound is returned when an observer kind has not been registered.
var ErrObserverNotFound = errors.New("health: observer not found")

// ErrProbeNotFound is returned when a probe name has not been registered.
var ErrProbeNotFound = errors.New("health: probe not found")

// NewService returns a Service with no registered probes or observers.
func NewService() *Service {
	return &Service{
		registry:      make(map[string]*probe.Probe),
		observers:     make(map[string]*subscriber.Observer),
		subscriptions: make(map[string]*subscriber.Subscriber),
		done:          make(chan struct{}),
		ticks:         make(chan *probe.Tick, 1),
		registryWG:    &sync.WaitGroup{},
		subscriberWG:  &sync.WaitGroup{},
		subscribers:   []*subscriber.Subscriber{},
	}
}

// Service maintains probes, subscribers, and observers for a single service.
//
// Use NewService to create a Service; the zero value is not initialized for
// registration.
//
// Register and Observe are setup-time calls. Start begins running all probes,
// waits for their initial checks, and fan-outs their ticks to subscribers. Start
// and Stop are idempotent. Stop is safe before Start and preserves observers so
// the service can be started again later with the same observer instances.
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
//
// Register is intended for setup before Start. Later registrations with the same
// name replace earlier ones.
func (s *Service) Register(regs ...*Registration) {
	for _, reg := range regs {
		s.registry[reg.Name] = probe.NewProbe(reg.Name, reg.Period, reg.Checker)
	}
}

// Observer returns the observer for kind.
//
// It returns ErrObserverNotFound if kind has not been registered.
func (s *Service) Observer(kind string) (*subscriber.Observer, error) {
	observer, ok := s.observers[kind]
	if !ok {
		return nil, ErrObserverNotFound
	}

	return observer, nil
}

// Error returns the observer error for kind.
//
// It returns ErrObserverNotFound if kind has not been registered. A nil result
// means the observer is currently healthy.
func (s *Service) Error(kind string) error {
	observer, err := s.Observer(kind)
	if err != nil {
		return err
	}

	return observer.Error()
}

// Watch returns a watcher for current and future observer errors for kind.
//
// It returns ErrObserverNotFound if kind has not been registered. The watcher
// receives the observer's current error immediately, then receives the current
// error again after each matching probe tick is processed. Sends are best-effort
// and coalesced to the latest error when the receiver is slow. Close the watcher
// when the receiver no longer needs updates; stopping the service does not close
// existing watchers.
func (s *Service) Watch(kind string) (watcher.Subscription, error) {
	observer, err := s.Observer(kind)
	if err != nil {
		return nil, err
	}

	return observer.Watch(), nil
}

// Observe registers an observer kind that tracks the probes listed in names.
//
// Observe is intended for setup before Start. When creating a new observer kind,
// it returns an error if any probe name has not been registered. Repeated calls
// with the same kind are idempotent, keep the original probe set, and do not
// validate the new names. Unknown probe names for a new kind are reported with
// ErrProbeNotFound; multiple unknown probes are joined into one error.
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
//
// Existing observers continue receiving updates if the service is stopped and
// started again later. Start runs probes concurrently and waits for each probe's
// initial check before returning; observer state is updated asynchronously after
// those ticks are fanned out. Repeated calls while the service is running are
// no-ops. Call Stop after Start has returned during normal shutdown. If a probe
// fails to start, Start cleans up partially started probes and subscriptions,
// leaves the service stopped, and the service may be started again later.
func (s *Service) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if s.running {
		return nil
	}

	s.prepareStart()

	tickChannels, err := s.startProbes(ctx)
	if err != nil {
		s.cleanupStartFailure(ctx)
		return err
	}

	s.mergeChannels(tickChannels)
	s.subscriberWG.Go(func() {
		s.sendToSubscribers()
	})

	s.running = true
	return nil
}

// Stop stops all probes and closes all subscribers.
//
// Stop is safe before Start and safe to call multiple times. It waits for
// in-flight fan-in and fan-out work to finish before returning. Observer
// instances are preserved; if the service is started again later, those
// observers are attached to fresh subscribers. If ctx expires before shutdown
// work finishes, Stop returns an error and keeps the service marked running so
// callers can retry with a valid context.
func (s *Service) Stop(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.running {
		s.closeSubscriptions()
		return s.waitObservers(ctx)
	}

	// Stop probes first so their tick channels can close and sendTick goroutines can exit cleanly.
	if err := s.stopProbes(ctx); err != nil {
		return err
	}

	// Signal all sendTick goroutines to stop and wait for them to finish.
	close(s.done)
	if err := wait(ctx, s.registryWG); err != nil {
		return err
	}

	// Now it is safe to close the fan-in channel: no further sends can occur.
	close(s.ticks)

	// Wait for the fan-out goroutine to drain/exit and close all subscribers.
	if err := wait(ctx, s.subscriberWG); err != nil {
		return err
	}

	// Ensure observers have finished draining their subscriber channels before a restart.
	if err := s.waitObservers(ctx); err != nil {
		return err
	}

	s.running = false
	return nil
}

func (s *Service) subscribe(kind string, names ...string) *subscriber.Subscriber {
	sub := subscriber.NewSubscriber(names)
	s.subscriptions[kind] = sub
	s.subscribers = append(s.subscribers, sub)
	return sub
}

func (s *Service) closeSubscriptions() {
	for _, sub := range s.subscriptions {
		sub.Close()
	}
}

func (s *Service) waitObservers(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, observer := range s.observers {
		wg.Go(func() {
			observer.Wait()
		})
	}

	return wait(ctx, &wg)
}

func (s *Service) startProbes(ctx context.Context) ([]<-chan *probe.Tick, error) {
	tickChannels := make([]<-chan *probe.Tick, len(s.registry))
	probes := make([]*probe.Probe, 0, len(s.registry))
	for _, p := range s.registry {
		probes = append(probes, p)
	}

	var g sync.ErrorsGroup
	for i, p := range probes {
		g.Go(func() error {
			ch, err := p.Start(ctx)
			tickChannels[i] = ch
			return err
		})
	}

	if err := g.Wait(); err != nil {
		if stopErr := s.stopProbes(context.WithoutCancel(ctx)); stopErr != nil {
			return nil, errors.Join(err, stopErr)
		}
		return nil, err
	}

	return tickChannels, nil
}

func (s *Service) mergeChannels(tickChannels []<-chan *probe.Tick) {
	for _, ch := range tickChannels {
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

func (s *Service) cleanupStartFailure(ctx context.Context) {
	s.closeSubscriptions()
	_ = s.waitObservers(context.WithoutCancel(ctx))
	close(s.done)
	close(s.ticks)
}

func (s *Service) stopProbes(ctx context.Context) error {
	var g sync.ErrorsGroup
	for _, p := range s.registry {
		g.Go(func() error {
			return p.Stop(ctx)
		})
	}

	return g.Wait()
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

func wait(ctx context.Context, wg *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
