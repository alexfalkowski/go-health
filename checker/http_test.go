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
