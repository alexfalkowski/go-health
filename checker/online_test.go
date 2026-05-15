package checker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/stretchr/testify/require"
)

func TestOnlineCheckerClonesURLs(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(healthy.Close)

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(unhealthy.Close)

	urls := []string{healthy.URL}
	check := checker.NewOnlineChecker(time.Second, checker.WithURLs(urls...))

	urls[0] = unhealthy.URL

	require.NoError(t, check.Check(t.Context()))
}

func TestOnlineCheckerReturnsOnFirstHealthyURL(t *testing.T) {
	slowStarted := make(chan struct{})
	slowCanceled := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(slowStarted)
		<-r.Context().Done()
		close(slowCanceled)
	}))
	t.Cleanup(slow.Close)

	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-slowStarted
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(healthy.Close)

	check := checker.NewOnlineChecker(5*time.Second, checker.WithURLs(slow.URL, healthy.URL))

	started := time.Now()
	require.NoError(t, check.Check(t.Context()))

	require.Less(t, time.Since(started), 500*time.Millisecond)
	require.Eventually(t, func() bool {
		select {
		case <-slowCanceled:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}
