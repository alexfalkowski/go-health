package svr

import (
	"context"
	"errors"
	"time"

	"github.com/alexfalkowski/go-health/pkg/chk"
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
	registry := make(map[string]*Probe)
	subscribers := []*sub.Subscriber{}

	return &server{registry: registry, subscribers: subscribers}
}

type server struct {
	registry    map[string]*Probe
	subscribers []*sub.Subscriber
	ctx         context.Context
	cancel      context.CancelFunc
	ticks       chan *ProbeTick
}

func (s *server) Register(name string, period time.Duration, checker chk.Checker) error {
	if _, ok := s.registry[name]; ok {
		return ErrProbeExists
	}

	s.registry[name] = NewProbe(name, period, checker)

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

	s.ticks = make(chan *ProbeTick)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	chs := []chan *ProbeTick{}
	for _, p := range s.registry {
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
	close(s.ticks)

	return nil
}

func (s *server) mergeChannels(chs []chan *ProbeTick) {
	for _, ch := range chs {
		go s.sendTick(ch)
	}
}

func (s *server) sendTick(ch chan *ProbeTick) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case t := <-ch:
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
