package sql_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
)

var errPingFailed = errors.New("ping failed")

func Example() {
	check := checker.NewDBChecker(failingPinger{}, time.Second)

	fmt.Println(errors.Is(check.Check(context.Background()), errPingFailed))
	// Output: true
}

type failingPinger struct{}

func (failingPinger) PingContext(context.Context) error {
	return errPingFailed
}
