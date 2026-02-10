// Package subscriber provides subscription and observer state tracking for probe ticks.
//
// A subscriber consumes ticks (typically produced by the probe package) and maintains
// the latest health state per probe name. The primary type is Observer, which runs a
// goroutine that reads ticks and updates an internal error map.
//
// Observers are commonly used by higher-level orchestration (see the server package)
// to aggregate health across multiple probes into a single "kind" (e.g. "livez" or
// "readyz").
//
// # Errors and aggregation
//
// An Observer exposes:
//   - Error(): a single error representing the current unhealthy state (if any),
//     typically produced by joining individual probe errors.
//   - Errors(): a snapshot copy of the current per-probe errors.
//
// The exact aggregation behavior is defined by this package; callers should treat
// Error() as a convenient summary and Errors() as the detailed view.
//
// # Concurrency
//
// Observers are safe for concurrent use: tick processing happens in a dedicated
// goroutine, and reads of the current state are protected internally.
//
// # Lifecycle
//
// Observers are usually started by orchestration code. Ensure they are stopped when
// no longer needed to avoid leaking goroutines.
package subscriber
