package checker_test

import (
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/stretchr/testify/require"
)

func FuzzHTTPCheckerRequestAndStatus(f *testing.F) {
	f.Add("https://example.com/health", 200, "X-Health-Check", "readyz")
	f.Add("https://example.com/health", 204, "Authorization", "Bearer token")
	f.Add("https://example.com/health", 399, "X-Health-Check", "livez")
	f.Add("https://example.com/health", 400, "X-Health-Check", "readyz")
	f.Add("https://example.com/health", 500, "X-Health-Check", "readyz")
	f.Add("://bad-url", 200, "X-Health-Check", "readyz")
	f.Add("", 200, "X-Health-Check", "readyz")
	f.Add("https://example.com/health", 200, "", "readyz")

	f.Fuzz(func(t *testing.T, url string, status int, headerName, headerValue string) {
		transport := &recordingRoundTripper{status: status}
		check := checker.NewHTTPChecker(
			url,
			time.Second,
			checker.WithRoundTripper(transport),
			checker.WithHeader(headerName, headerValue),
		)

		err := check.Check(t.Context())

		if transport.request == nil {
			require.Error(t, err)
			require.NotErrorIs(t, err, checker.ErrInvalidStatusCode)
			return
		}

		require.Equal(t, http.MethodGet, transport.request.Method)
		if headerName != "" {
			require.Equal(t, headerValue, transport.request.Header.Get(headerName))
		}
		if status >= http.StatusBadRequest {
			require.ErrorIs(t, err, checker.ErrInvalidStatusCode)
			return
		}
		require.NoError(t, err)
	})
}

func FuzzOnlineCheckerURLsAndStatuses(f *testing.F) {
	f.Add("https://example.com/first", 200, "https://example.com/second", 500, "://bad-url", 404)
	f.Add("https://example.com/first", 204, "https://example.com/second", 404, "https://example.com/third", 500)
	f.Add("https://example.com/first", 500, "https://example.com/second", 404, "https://example.com/third", 399)
	f.Add("://bad-url", 200, "https://example.com/second", 500, "https://example.com/third", 500)
	f.Add("", 200, "https://example.com/second", 204, "https://example.com/third", 500)

	f.Fuzz(func(t *testing.T, url1 string, status1 int, url2 string, status2 int, url3 string, status3 int) {
		transport := &statusRoundTripper{
			statuses: map[string]int{
				url1: status1,
				url2: status2,
				url3: status3,
			},
		}
		check := checker.NewOnlineChecker(
			time.Second,
			checker.WithRoundTripper(transport),
			checker.WithURLs(url1, url2, url3),
		)

		err := check.Check(t.Context())

		if transport.healthy() {
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, checker.ErrNotOnline)
	})
}

type recordingRoundTripper struct {
	request *http.Request
	status  int
}

func (t *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.request = req.Clone(req.Context())
	t.request.Header = req.Header.Clone()

	return response(req, t.status), nil
}

type statusRoundTripper struct {
	statuses map[string]int
	mux      sync.Mutex
	online   bool
}

func (t *statusRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	status, ok := t.statuses[req.URL.String()]
	if !ok {
		return nil, errors.New("unexpected URL")
	}
	if onlineStatus(status) {
		t.mux.Lock()
		t.online = true
		t.mux.Unlock()
	}

	return response(req, status), nil
}

func (t *statusRoundTripper) healthy() bool {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.online
}

func response(req *http.Request, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}
}

func onlineStatus(status int) bool {
	return status == http.StatusOK || status == http.StatusNoContent
}
