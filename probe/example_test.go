package probe_test

import (
	"context"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/probe"
)

func ExampleNewProbe() {
	p := probe.NewProbe("noop", 20*time.Millisecond, checker.NewNoopChecker())
	ticks, err := p.Start(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	tick := <-ticks

	fmt.Println(tick.Name(), tick.Error() == nil)

	_ = p.Stop(context.Background())
	// Output: noop true
}
