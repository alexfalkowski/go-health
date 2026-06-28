package subscriber_test

import (
	"errors"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/alexfalkowski/go-sync"
	"github.com/stretchr/testify/require"
)

const (
	dbProbeName    = "db"
	cacheProbeName = "cache"
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
	t.Parallel()

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

func TestSubscriberDropsTicksWhenBufferIsFull(t *testing.T) {
	t.Parallel()

	s := subscriber.NewSubscriber([]string{"a"})
	first := probe.NewTick("a", nil)
	second := probe.NewTick("a", nil)

	s.Send(first)

	sent := make(chan struct{})
	go func() {
		s.Send(second)
		close(sent)
	}()

	select {
	case <-sent:
	case <-time.After(time.Second):
		require.Fail(t, "send blocked while subscriber buffer was full")
	}

	require.Equal(t, first, <-s.Receive())
	select {
	case tick := <-s.Receive():
		require.Fail(t, "unexpected tick after best-effort drop", tick.Name())
	default:
	}
}

func TestObserverStartsWithNilErrorsForTrackedNames(t *testing.T) {
	t.Parallel()

	s := subscriber.NewSubscriber([]string{dbProbeName, cacheProbeName})
	ob := subscriber.NewObserver([]string{dbProbeName, cacheProbeName}, s)
	s.Close()
	ob.Wait()

	errs := ob.Errors()

	require.NoError(t, ob.Error())
	require.Contains(t, errs, dbProbeName)
	require.NoError(t, errs[dbProbeName])
	require.Contains(t, errs, cacheProbeName)
	require.NoError(t, errs[cacheProbeName])
}

func TestObserverErrorsReturnsCopy(t *testing.T) {
	t.Parallel()

	s := subscriber.NewSubscriber([]string{dbProbeName, cacheProbeName})
	ob := subscriber.NewObserver([]string{dbProbeName, cacheProbeName}, s)
	errDB := errors.New("db failed")

	s.Send(probe.NewTick(dbProbeName, errDB))
	s.Close()
	ob.Wait()

	errs := ob.Errors()
	errs[dbProbeName] = nil
	delete(errs, cacheProbeName)

	fresh := ob.Errors()
	require.ErrorIs(t, fresh[dbProbeName], errDB)
	require.Contains(t, fresh, cacheProbeName)
	require.NoError(t, fresh[cacheProbeName])
	require.ErrorIs(t, ob.Error(), errDB)
}

func TestObserverWatchSendsCurrentErrorAndUpdates(t *testing.T) {
	s := subscriber.NewSubscriber([]string{dbProbeName})
	ob := subscriber.NewObserver([]string{dbProbeName}, s)
	t.Cleanup(func() {
		s.Close()
		ob.Wait()
	})

	watcher := ob.Watch()
	defer watcher.Close()
	updates := watcher.Receive()

	require.NoError(t, receiveError(t, updates))

	errDB := errors.New("db failed")
	s.Send(probe.NewTick(dbProbeName, errDB))

	require.ErrorIs(t, receiveError(t, updates), errDB)
}

func TestObserverWatchClosesWhenStopped(t *testing.T) {
	s := subscriber.NewSubscriber([]string{dbProbeName})
	ob := subscriber.NewObserver([]string{dbProbeName}, s)
	t.Cleanup(func() {
		s.Close()
		ob.Wait()
	})

	watcher := ob.Watch()
	updates := watcher.Receive()
	require.NoError(t, receiveError(t, updates))

	watcher.Close()
	_, ok := receive(t, updates)
	require.False(t, ok)
}

func TestObserverWatchSendsTicksWithoutStatusChange(t *testing.T) {
	s := subscriber.NewSubscriber([]string{dbProbeName})
	ob := subscriber.NewObserver([]string{dbProbeName}, s)
	t.Cleanup(func() {
		s.Close()
		ob.Wait()
	})

	watcher := ob.Watch()
	defer watcher.Close()
	updates := watcher.Receive()
	require.NoError(t, receiveError(t, updates))

	s.Send(probe.NewTick(dbProbeName, nil))

	require.NoError(t, receiveError(t, updates))
}

func TestErrorsErrorJoinsAnnotatedErrors(t *testing.T) {
	t.Parallel()

	errDB := errors.New("db failed")
	errCache := errors.New("cache failed")
	err := subscriber.Errors{
		dbProbeName:    errDB,
		cacheProbeName: errCache,
		"search":       nil,
	}.Error()

	require.ErrorIs(t, err, errDB)
	require.ErrorIs(t, err, errCache)
	require.ErrorContains(t, err, "db:")
	require.ErrorContains(t, err, "cache:")
}

func receiveError(t *testing.T, updates <-chan error) error {
	t.Helper()

	err, ok := receive(t, updates)
	require.True(t, ok)

	return err
}

func receive(t *testing.T, updates <-chan error) (error, bool) {
	t.Helper()

	select {
	case err, ok := <-updates:
		return err, ok
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for observer update")
		return nil, false
	}
}
