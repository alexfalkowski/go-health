// Package probe provides periodic execution of health checkers.
//
// A Probe runs a checker.Checker on a fixed interval and emits a Tick containing
// the probe name and the resulting error (nil meaning healthy). Probes are
// typically used as building blocks for higher-level orchestration (for example,
// the server package), where multiple probes are registered per service and their
// ticks are aggregated by observers.
//
// # Emitted ticks
//
// Each execution produces a Tick:
//
//   - Name identifies the probe that produced the result.
//   - Err is the error returned by the underlying checker (nil if healthy).
//
// Ticks are sent on a channel so consumers can process results asynchronously.
//
// # Scheduling behavior
//
// Probes perform an initial check when started, then continue checking on the
// configured period. The check itself is executed with a context passed in by the
// caller or created by orchestration code.
//
// # Concurrency and lifecycle
//
// Probes run their scheduling loop in a goroutine. Call Stop (directly or via
// orchestration code) when the probe is no longer needed to avoid leaking
// goroutines and timers.
//
// # Usage
//
// Typical usage is to construct probes indirectly via the server package.
// If you use probe directly, you generally:
//
//  1. Create a Probe with a name, period, and checker.
//  2. Start it.
//  3. Read ticks from its channel until you stop it.
package probe
