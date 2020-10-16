package svr

import (
	"context"
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

// Server will maintain all the probes and start and stop them.
type Server interface {
	Register(string, time.Duration, chk.Checker) error
	Subscribe(...string) *sub.Subscriber
	Start() error
	Stop() error
}

// NewServer for health.
func NewServer() Server {
	registry := make(map[string]*prb.Probe)
	subscribers := []*sub.Subscriber{}

	return &server{registry: registry, subscribers: subscribers}
}

type server struct {
	registry    map[string]*prb.Probe
	subscribers []*sub.Subscriber
	ctx         context.Context
	cancel      context.CancelFunc
	ticks       chan *prb.Tick
	wg          *sync.WaitGroup
}

func (s *server) Register(name string, period time.Duration, checker chk.Checker) error {
	if _, ok := s.registry[name]; ok {
		return ErrProbeExists
	}

	s.registry[name] = prb.NewProbe(name, period, checker)

	return nil
}

func (s *server) Subscribe(names ...string) *sub.Subscriber {
	sub := sub.NewSubscriber(names)

	s.subscribers = append(s.subscribers, sub)

	return sub
}

func (s *server) Start() error {
	if len(s.registry) == 0 {
		return ErrNoRegistrations
	}

	s.wg = &sync.WaitGroup{}
	s.ticks = make(chan *prb.Tick)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	chs := []chan *prb.Tick{}

	for _, p := range s.registry {
		s.wg.Add(1)

		chs = append(chs, p.Start(s.ctx))
	}

	s.mergeChannels(chs)

	go s.sendToSubscribers()

	return nil
}

func (s *server) Stop() error {
	if len(s.registry) == 0 {
		return ErrNoRegistrations
	}

	s.cancel()
	s.wg.Wait()

	close(s.ticks)

	return nil
}

func (s *server) mergeChannels(chs []chan *prb.Tick) {
	for _, ch := range chs {
		go s.sendTick(ch)
	}
}

func (s *server) sendTick(ch chan *prb.Tick) {
	defer s.wg.Done()

	for {
		select {
		case t, ok := <-ch:
			if !ok {
				return
			}

			s.ticks <- t
		}
	}
}

func (s *server) sendToSubscribers() {
	for t := range s.ticks {
		for _, sub := range s.subscribers {
			if sub.HasName(t.Name()) {
				sub.Channel() <- t.Error()
			}
		}
	}
}
