package net_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
)

var errDialFailed = errors.New("dial failed")

func ExampleDialer() {
	check := checker.NewTCPChecker(
		"cache:6379",
		time.Second,
		checker.WithDialer(failingDialer{}),
	)

	fmt.Println(errors.Is(check.Check(context.Background()), errDialFailed))
	// Output: true
}

type failingDialer struct{}

func (failingDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	return nil, errDialFailed
}
