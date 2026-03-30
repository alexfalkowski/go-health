![Gopher](assets/gopher.png)
[![CircleCI](https://circleci.com/gh/alexfalkowski/go-health.svg?style=shield)](https://circleci.com/gh/alexfalkowski/go-health)
[![codecov](https://codecov.io/gh/alexfalkowski/go-health/graph/badge.svg?token=Q7B3VZYL9K)](https://codecov.io/gh/alexfalkowski/go-health)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexfalkowski/go-health)](https://goreportcard.com/report/github.com/alexfalkowski/go-health)
[![Go Reference](https://pkg.go.dev/badge/github.com/alexfalkowski/go-health/v2.svg)](https://pkg.go.dev/github.com/alexfalkowski/go-health/v2)
[![Stability: Active](https://masterminds.github.io/stability/active.svg)](https://masterminds.github.io/stability/active.html)

# Health Monitoring Pattern

`go-health` is a small Go library for building asynchronous health monitoring.
It separates the problem into four layers:

1. `checker`: how to test one dependency.
2. `probe`: when to run that test.
3. `subscriber`: how to keep the latest probe state.
4. `server`: how to orchestrate probes and observers for a service.

This keeps the pieces reusable. You can use a single checker directly, wire
your own probe loop, or use the `server` package for a complete "register,
observe, start, stop" workflow.

## Why this library

This repository focuses on a few constraints:

- Health checks should run asynchronously so a health endpoint does not trigger a fresh dependency check for every request.
- The public API should stay small and transport-agnostic.
- The core packages should be easy to compose into custom liveness, readiness, or dependency-specific health views.
- The implementation should stay lightweight and friendly to tests.

Related background:

- [Health Endpoint Monitoring pattern](https://learn.microsoft.com/azure/architecture/patterns/health-endpoint-monitoring)
- [Kubernetes liveness and readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Health Check API pattern](https://microservices.io/patterns/observability/health-check-api.html)

## Installation

```sh
go get github.com/alexfalkowski/go-health/v2
```

The module path includes `/v2`:

```go
import "github.com/alexfalkowski/go-health/v2/server"
```

## Package overview

| Package | Purpose |
| --- | --- |
| `checker` | Reusable health checks for HTTP, TCP, SQL ping, online connectivity, readiness gates, and no-op checks. |
| `probe` | Periodically runs a checker and emits ticks. |
| `subscriber` | Tracks the latest error for a selected set of probe names. |
| `server` | Registers probes per service, creates observers, and manages lifecycle. |
| `net` | Small interfaces used to customize network dialing. |
| `sql` | Small interfaces used to customize database pinging. |

## Key behaviors

- `probe.Start` performs an immediate check before the periodic loop continues.
- A probe with an invalid period emits a single error tick and closes.
- `HTTPChecker`, `TCPChecker`, `DBChecker`, and `OnlineChecker` use a default timeout of `30s` when you pass `0`.
- `OnlineChecker` reports healthy if any configured URL returns `200 OK` or `204 No Content`.
- `subscriber.Observer` starts with `nil` errors for every tracked probe name and updates as ticks arrive.
- `server.Service` and `server.Server` keep observer instances across stop/start cycles, so existing observers continue receiving ticks after a restart.

## End-to-end example

The `server` package is the usual entry point when you want one or more health
views for a service:

```go
package main

import (
	"errors"
	"log"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/server"
)

func main() {
	const (
		timeout = 5 * time.Second
		period  = 30 * time.Second
	)

	s := server.NewServer()

	readyChecker := checker.NewReadyChecker(errors.New("starting up"))
	apiChecker := checker.NewHTTPChecker("https://example.com/health", timeout)
	cacheChecker := checker.NewTCPChecker("cache.internal:6379", timeout)

	apiRegistration := server.NewRegistration("api", period, apiChecker)
	cacheRegistration := server.NewRegistration("cache", period, cacheChecker)
	readyRegistration := server.NewRegistration("ready", time.Second, readyChecker)

	s.Register("payments", apiRegistration, cacheRegistration, readyRegistration)

	if err := s.Observe("payments", "livez", apiRegistration.Name, cacheRegistration.Name); err != nil {
		log.Fatal(err)
	}

	if err := s.Observe("payments", "readyz", apiRegistration.Name, cacheRegistration.Name, readyRegistration.Name); err != nil {
		log.Fatal(err)
	}

	s.Start()
	defer s.Stop()

	readyChecker.Ready()

	livez, err := s.Observer("payments", "livez")
	if err != nil {
		log.Fatal(err)
	}

	readyz, err := s.Observer("payments", "readyz")
	if err != nil {
		log.Fatal(err)
	}

	if err := livez.Error(); err != nil {
		log.Printf("livez unhealthy: %v", err)
	}

	for name, err := range readyz.Errors() {
		log.Printf("readyz %s = %v", name, err)
	}
}
```

Notes:

- `Observe` validates that the service exists and every named probe was registered.
- The first meaningful observer state arrives after the initial probe tick is processed.
- Calling `Observe` again with the same service and kind is a no-op; it does not replace the existing observer.

## Direct checker examples

### HTTP checker

```go
check := checker.NewHTTPChecker("https://example.com/health", 5*time.Second)

if err := check.Check(context.Background()); err != nil {
	log.Printf("endpoint unhealthy: %v", err)
}
```

Use `checker.WithRoundTripper` when you need a custom transport:

```go
transport := &http.Transport{}
check := checker.NewHTTPChecker(
	"https://example.com/health",
	5*time.Second,
	checker.WithRoundTripper(transport),
)
```

### Online checker

`OnlineChecker` is useful when the question is "can this process reach the
outside world?" rather than "is this exact upstream healthy?"

```go
check := checker.NewOnlineChecker(
	5*time.Second,
	checker.WithURLs(
		"https://google.com/generate_204",
		"https://cp.cloudflare.com/generate_204",
	),
)

if err := check.Check(context.Background()); err != nil {
	log.Printf("no outbound connectivity: %v", err)
}
```

### Readiness gate

`ReadyChecker` lets your application control readiness explicitly:

```go
ready := checker.NewReadyChecker(errors.New("warming caches"))

if err := ready.Check(context.Background()); err != nil {
	log.Printf("not ready yet: %v", err)
}

// Later, once startup work is complete.
ready.Ready()
```

## Manual probe usage

If you do not need the `server` package, you can work directly with a probe:

```go
p := probe.NewProbe("api", 10*time.Second, checker.NewNoopChecker())
ticks := p.Start()
defer p.Stop()

tick := <-ticks
log.Printf("%s healthy=%t", tick.Name(), tick.Error() == nil)
```

## Working with observers

Observers give you both a summary and a detailed view:

```go
summary := observer.Error()  // joined error across unhealthy probes
details := observer.Errors() // copy of the latest per-probe error map
```

`Errors()` returns a copy, so callers can inspect or mutate the returned map
without affecting the observer's internal state.

## Development

This repository uses a `bin/` git submodule for shared build tooling. Before
running the `make` targets, initialize the submodule:

```sh
git submodule sync
git submodule update --init
```

Common commands:

```sh
make dep
make lint
make sec
make specs
make coverage
```

Local test notes:

- Some tests make real network calls.
- Some tests expect a local status service on `STATUS_PORT`, defaulting to `6000`.
- CircleCI starts `alexfalkowski/status:latest` for that service during CI.

## Documentation

- Package documentation lives in `doc.go` files and exported symbol comments.
- Runnable pkg.go.dev examples live in `example_test.go` files.
- README examples should always import `github.com/alexfalkowski/go-health/v2/...`.
