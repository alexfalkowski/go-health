package subscriber_test

import (
	"testing"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-sync"
	"github.com/stretchr/testify/require"
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

func TestSubscriberClonesNames(t *testing.T) {
	names := []string{"a"}
	s := subscriber.NewSubscriber(names)

	names[0] = "b"

	s.Send(probe.NewTick("a", nil))

	select {
	case tick := <-s.Receive():
		require.Equal(t, "a", tick.Name())
	default:
		require.Fail(t, "expected tick for the original probe name")
	}

	s.Send(probe.NewTick("b", nil))

	select {
	case tick := <-s.Receive():
		require.Fail(t, "unexpected tick for mutated probe name", tick.Name())
	default:
	}
}
