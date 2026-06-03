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

func TestOnlineCheckerStatusCodes(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		wantNotOnline bool
	}{
		{name: "ok", status: http.StatusOK},
		{name: "no content", status: http.StatusNoContent},
		{name: "not modified", status: http.StatusNotModified, wantNotOnline: true},
		{name: "server error", status: http.StatusInternalServerError, wantNotOnline: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			t.Cleanup(upstream.Close)

			check := checker.NewOnlineChecker(time.Second, checker.WithURLs(upstream.URL))

			err := check.Check(t.Context())

			if !tt.wantNotOnline {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, checker.ErrNotOnline)
		})
	}
}

func TestOnlineCheckerReturnsNotOnlineWhenEveryURLFails(t *testing.T) {
	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(unhealthy.Close)

	missing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(missing.Close)

	check := checker.NewOnlineChecker(time.Second, checker.WithURLs(unhealthy.URL, missing.URL, "://bad-url"))

	err := check.Check(t.Context())

	require.ErrorIs(t, err, checker.ErrNotOnline)
}
