// Package checker provides health check implementations.
//
// A checker is a small unit of work that reports health by returning nil (healthy)
// or a non-nil error (unhealthy). All checkers implement the Checker interface:
//
//	Checker.Check(ctx) error
//
// The implementations in this package are intentionally transport-agnostic and
// dependency-light so you can compose them into higher-level orchestration
// (for example, via the server and probe packages).
//
// # Timeouts
//
// Most checkers accept a timeout duration. A zero timeout uses a sensible default.
//
// # Options
//
// Some checkers accept functional options (see Option) to configure shared
// dependencies such as HTTP transports or dialers.
//
// # Common implementations
//
// The package includes, among others:
//
//   - HTTPChecker: performs an HTTP GET and fails on 4xx/5xx responses.
//   - TCPChecker: dials a TCP address to verify connectivity.
//   - DBChecker: pings an SQL-like dependency via the sql.Pinger interface.
//   - OnlineChecker: periodically validates external connectivity using HTTP.
//   - ReadyChecker: reports an error until explicitly marked ready.
//   - NoopChecker: always reports healthy.
//
// Checkers should be safe to call concurrently unless documented otherwise.
package checker
