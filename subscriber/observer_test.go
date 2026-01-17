package subscriber_test

import (
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/subscriber"
)

func TestObserverStart(t *testing.T) {
	sub := subscriber.NewSubscriber([]string{"p"})
	defer sub.Close()

	ob := subscriber.NewObserver([]string{"p"}, sub)
	defer ob.Stop()

	started := make(chan struct{})
	go func() {
		ob.Start()
		ob.Start()
		close(started)
	}()

	select {
	case <-started:
	case <-time.After(1 * time.Second):
		t.Fatal("observer start timed out")
	}
}

func TestObserverStop(t *testing.T) {
	sub := subscriber.NewSubscriber([]string{"p"})
	defer sub.Close()

	ob := subscriber.NewObserver([]string{"p"}, sub)

	stopped := make(chan struct{})
	go func() {
		ob.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("observer stop timed out")
	}
}

func TestObserverStopIdempotent(t *testing.T) {
	sub := subscriber.NewSubscriber([]string{"p"})
	defer sub.Close()

	ob := subscriber.NewObserver([]string{"p"}, sub)

	ob.Stop()

	stopped := make(chan struct{})
	go func() {
		ob.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("observer second stop timed out")
	}
}
