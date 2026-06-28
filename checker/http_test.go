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
		{name: "not modified", status: http.StatusNotModified},
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
