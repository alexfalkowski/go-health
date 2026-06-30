package checker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/stretchr/testify/require"
)

func TestHTTPCheckerStatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		status            int
		wantInvalidStatus bool
	}{
		{name: "ok", status: http.StatusOK},
		{name: "no content", status: http.StatusNoContent},
		{name: "redirect", status: http.StatusFound, wantInvalidStatus: true},
		{name: "not modified", status: http.StatusNotModified, wantInvalidStatus: true},
		{name: "bad request", status: http.StatusBadRequest, wantInvalidStatus: true},
		{name: "server error", status: http.StatusInternalServerError, wantInvalidStatus: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			t.Cleanup(upstream.Close)

			check := checker.NewHTTPChecker(upstream.URL, time.Second)

			err := check.Check(t.Context())

			if !tt.wantInvalidStatus {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, checker.ErrInvalidStatusCode)
		})
	}
}

func TestHTTPCheckerDoesNotFollowRedirects(t *testing.T) {
	t.Parallel()

	redirected := make(chan http.Header, 1)
	target := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		redirected <- r.Header.Clone()
	}))
	t.Cleanup(target.Close)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL)
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(upstream.Close)

	check := checker.NewHTTPChecker(
		upstream.URL,
		time.Second,
		checker.WithHeader("X-Health-Token", "secret"),
	)

	err := check.Check(t.Context())

	require.ErrorIs(t, err, checker.ErrInvalidStatusCode)
	select {
	case headers := <-redirected:
		require.Failf(t, "redirect target received request", "headers: %v", headers)
	default:
	}
}

func TestHTTPCheckerHeaders(t *testing.T) {
	t.Parallel()

	headers := make(chan http.Header, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		headers <- r.Header.Clone()
	}))
	t.Cleanup(upstream.Close)

	check := checker.NewHTTPChecker(
		upstream.URL,
		time.Second,
		checker.WithHeader("Authorization", "Bearer token"),
		checker.WithHeader("X-Health-Check", "readyz"),
		checker.WithHeader("X-Health-Check", "livez"),
	)

	require.NoError(t, check.Check(t.Context()))

	requestHeaders := <-headers
	require.Equal(t, "Bearer token", requestHeaders.Get("Authorization"))
	require.Equal(t, []string{"readyz", "livez"}, requestHeaders.Values("X-Health-Check"))
}
