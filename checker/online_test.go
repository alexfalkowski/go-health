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
