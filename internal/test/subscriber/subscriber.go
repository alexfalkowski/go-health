package subscriber

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Waiter waits for a subscriber observer loop to finish.
type Waiter interface {
	Wait()
}

// ErrorObserver reports the current observer error.
type ErrorObserver interface {
	Error() error
}

// Fataler reports fatal test or benchmark failures.
type Fataler interface {
	Helper()
	Fatal(args ...any)
}

// RequireObserverNoError waits until observer reports no error.
func RequireObserverNoError(t *testing.T, observer ErrorObserver) {
	t.Helper()

	WaitObserverNoError(t, observer)
}

// RequireObserverError waits until observer reports an error.
func RequireObserverError(t *testing.T, observer ErrorObserver) {
	t.Helper()

	require.Eventually(t, func() bool {
		return observer.Error() != nil
	}, time.Second, 10*time.Millisecond)
}

// RequireObserverStopped requires observer.Wait to return.
func RequireObserverStopped(t *testing.T, observer Waiter) {
	t.Helper()

	stopped := make(chan struct{})
	go func() {
		observer.Wait()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		require.Fail(t, "observer did not stop")
	}
}

// WaitObserverNoError waits until observer reports no error.
func WaitObserverNoError(tb Fataler, observer ErrorObserver) {
	tb.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if observer.Error() == nil {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	if err := observer.Error(); err != nil {
		tb.Fatal(err)
	}
}
