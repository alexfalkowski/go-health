package server_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/server"
)

func ExampleNewServer() {
	errNotReady := errors.New("starting up")
	ready := checker.NewReadyChecker(errNotReady)

	s := server.NewServer()
	registration := server.NewRegistration("ready", 10*time.Millisecond, ready)
	s.Register("payments", registration)

	if err := s.Observe("payments", "readyz", registration.Name); err != nil {
		fmt.Println(err)
		return
	}

	if err := s.Start(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = s.Stop(context.Background())
	}()

	watcher := s.Watch("readyz")
	defer watcher.Close()

	fmt.Println(waitForUpdate(watcher.Receive(), func(err error) bool {
		return errors.Is(err, errNotReady)
	}))

	ready.Ready()

	fmt.Println(waitForUpdate(watcher.Receive(), func(err error) bool {
		return err == nil
	}))
	// Output:
	// true
	// true
}

func waitForUpdate(updates <-chan error, match func(error) bool) bool {
	timeout := time.After(250 * time.Millisecond)
	for {
		select {
		case err := <-updates:
			if match(err) {
				return true
			}
		case <-timeout:
			return false
		}
	}
}
