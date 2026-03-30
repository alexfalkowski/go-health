package checker_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
)

func ExampleNewHTTPChecker() {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	check := checker.NewHTTPChecker(upstream.URL, time.Second)

	fmt.Println(check.Check(context.Background()) == nil)
	// Output: true
}

func ExampleNewOnlineChecker() {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	check := checker.NewOnlineChecker(time.Second, checker.WithURLs(upstream.URL))

	fmt.Println(check.Check(context.Background()) == nil)
	// Output: true
}

func ExampleNewReadyChecker() {
	errNotReady := errors.New("not ready")
	check := checker.NewReadyChecker(errNotReady)

	fmt.Println(errors.Is(check.Check(context.Background()), errNotReady))

	check.Ready()

	fmt.Println(check.Check(context.Background()) == nil)
	// Output:
	// true
	// true
}
