// Package checker provides reusable health check implementations.
//
// A Checker is a small unit of work that reports health by returning nil when the
// dependency is healthy or a non-nil error when it is not. The package keeps the
// checkers focused on "how to check" rather than "when to check", which makes
// them easy to reuse directly or through higher-level orchestration such as the
// probe and server packages.
//
// # Common implementations
//
// The package includes:
//
//   - HTTPChecker for HTTP GET health endpoints.
//   - TCPChecker for simple TCP connectivity checks.
//   - DBChecker for SQL-like dependencies that expose PingContext.
//   - OnlineChecker for best-effort external connectivity checks across a list
//     of URLs.
//   - ReadyChecker for application-managed readiness gates.
//   - NoopChecker for always-healthy checks.
//
// # Timeouts and defaults
//
// HTTPChecker, TCPChecker, DBChecker, and OnlineChecker accept a timeout
// duration. Passing 0 uses a default timeout of 30 seconds.
// DBChecker and TCPChecker derive per-call contexts using [context.WithTimeoutCause].
// If their underlying dependency returns [context.Cause] after that timeout
// expires, the resulting error matches [ErrTimeout].
//
// OnlineChecker uses a built-in list of connectivity endpoints unless you
// override it with WithURLs. HTTPChecker and OnlineChecker use
// http.DefaultTransport by default, and TCPChecker uses net.DefaultDialer by
// default.
//
// # Options
//
// Some constructors accept functional options so shared dependencies can be
// injected without wrapping the checker:
//
//   - WithRoundTripper customizes the HTTP transport.
//   - WithDialer customizes TCP dialing.
//   - WithURLs replaces the OnlineChecker URL list.
//
// # Example
//
// End-to-end usage usually looks like:
//
//	httpChecker := checker.NewHTTPChecker("https://example.com/health", 5*time.Second)
//	readyChecker := checker.NewReadyChecker(errors.New("starting up"))
//
//	if err := httpChecker.Check(context.Background()); err != nil {
//		log.Printf("dependency unhealthy: %v", err)
//	}
//
//	readyChecker.Ready()
//
// All exported checker implementations are safe for concurrent use.
package checker
