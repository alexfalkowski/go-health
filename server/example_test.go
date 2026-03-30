package server_test

import (
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

	observer, err := s.Observer("payments", "readyz")
	if err != nil {
		fmt.Println(err)
		return
	}

	s.Start()
	defer s.Stop()

	fmt.Println(waitForCondition(func() bool {
		return errors.Is(observer.Error(), errNotReady)
	}))

	ready.Ready()

	fmt.Println(waitForCondition(func() bool {
		return observer.Error() == nil
	}))
	// Output:
	// true
	// true
}

func waitForCondition(fn func() bool) bool {
	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}

		time.Sleep(5 * time.Millisecond)
	}

	return fn()
}
