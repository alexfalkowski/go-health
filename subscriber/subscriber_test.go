package subscriber_test

import (
	"sync"
	"testing"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/subscriber"
)

func TestCloseWhileSendingDoesNotPanic(t *testing.T) {
	for range 5000 {
		s := subscriber.NewSubscriber([]string{"a"})
		tick := probe.NewTick("a", nil)
		start := make(chan struct{})

		var ready sync.WaitGroup
		var done sync.WaitGroup

		for range 8 {
			ready.Add(1)
			done.Go(func() {
				ready.Done()
				<-start
				s.Send(tick)
			})
		}

		ready.Wait()
		close(start)
		s.Close()
		done.Wait()
	}
}
