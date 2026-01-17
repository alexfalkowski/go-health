package server_test

import (
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/server"
)

func TestServiceStop(t *testing.T) {
	s := server.NewService()
	r := server.NewRegistration("noop", 10*time.Millisecond, checker.NewNoopChecker())
	s.Register(r)
	s.Observe("livez", r.Name)

	stopped := make(chan struct{})
	go func() {
		s.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("service stop timed out")
	}
}

func TestServiceStopDoesNotBlockWithObservers(t *testing.T) {
	s := server.NewService()
	r := server.NewRegistration("noop", 10*time.Millisecond, checker.NewNoopChecker())
	s.Register(r)
	s.Observe("livez", r.Name)
	s.Start()

	stopped := make(chan struct{})
	go func() {
		s.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("service stop timed out")
	}
}
