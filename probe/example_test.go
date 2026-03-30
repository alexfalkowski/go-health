package probe_test

import (
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/probe"
)

func ExampleNewProbe() {
	p := probe.NewProbe("noop", 20*time.Millisecond, checker.NewNoopChecker())
	ticks := p.Start()
	tick := <-ticks

	fmt.Println(tick.Name(), tick.Error() == nil)

	p.Stop()
	// Output: noop true
}
