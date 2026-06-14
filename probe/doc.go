// Package probe provides periodic execution of health checkers.
//
// A Probe runs a checker.Checker on a fixed interval and emits Tick values on a
// channel. Probes are the scheduling layer of the library: the checker package
// defines what to run, while this package defines when to run it.
//
// # Tick model
//
// Each execution produces a Tick with accessor methods:
//
//   - Name() identifies the probe that produced the result.
//   - Error() returns the latest checker error, or nil when healthy.
//
// Ticks are delivered asynchronously on a channel so higher-level code can fan
// them out to observers or aggregate them into service-level health.
//
// # Scheduling behavior
//
// Start uses the supplied context for startup and the initial check before
// returning the channel, then continues on the configured period until Stop is
// called. If the context is canceled before startup completes, Start returns the
// context error and no tick channel. Canceling that context after Start returns
// does not stop the probe; use Stop to end the probe lifecycle. If the period is
// zero or negative, Start returns a closed channel after emitting a single tick
// whose error wraps ErrInvalidPeriod.
//
// # Lifecycle
//
// Start is idempotent while a probe is running: repeated calls with a live
// context return the same channel once startup has completed. Stop is also
// idempotent and cancels any in-flight check before waiting for the probe
// goroutine to exit or the supplied context to expire.
//
// # Example
//
//	p := probe.NewProbe("cache", 10*time.Second, checker.NewNoopChecker())
//	ticks, err := p.Start(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer p.Stop(context.Background())
//
//	tick := <-ticks
//	fmt.Println(tick.Name(), tick.Error() == nil)
package probe
