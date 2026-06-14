![Gopher](assets/gopher.png)
[![CircleCI](https://circleci.com/gh/alexfalkowski/go-health.svg?style=shield)](https://circleci.com/gh/alexfalkowski/go-health)
[![codecov](https://codecov.io/gh/alexfalkowski/go-health/graph/badge.svg?token=Q7B3VZYL9K)](https://codecov.io/gh/alexfalkowski/go-health)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexfalkowski/go-health)](https://goreportcard.com/report/github.com/alexfalkowski/go-health)
[![Go Reference](https://pkg.go.dev/badge/github.com/alexfalkowski/go-health/v2.svg)](https://pkg.go.dev/github.com/alexfalkowski/go-health/v2)
[![Stability: Active](https://masterminds.github.io/stability/active.svg)](https://masterminds.github.io/stability/active.html)

# 🩺 Health Monitoring Pattern

`go-health` is a small Go library for building asynchronous health monitoring.
It separates the problem into four layers:

1. `checker`: how to test one dependency.
2. `probe`: when to run that test.
3. `subscriber`: how to keep the latest probe state.
4. `server`: how to orchestrate probes and observers for a service.

This keeps the pieces reusable. You can use a single checker directly, wire
your own probe loop, or use the `server` package for a complete "register,
observe, start, stop" workflow.

## 🎯 Why this library

This repository focuses on a few constraints:

- Health checks should run asynchronously so a health endpoint does not trigger a fresh dependency check for every request.
- The public API should stay small and transport-agnostic.
- The core packages should be easy to compose into custom liveness, readiness, or dependency-specific health views.
- The implementation should stay lightweight and friendly to tests.

Related background:

- [Health Endpoint Monitoring pattern](https://learn.microsoft.com/azure/architecture/patterns/health-endpoint-monitoring)
- [Kubernetes liveness and readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Health Check API pattern](https://microservices.io/patterns/observability/health-check-api.html)

## 📦 Installation

```sh
go get github.com/alexfalkowski/go-health/v2
```

The module path includes `/v2`:

```go
import "github.com/alexfalkowski/go-health/v2/server"
```

## 🧭 Package overview

| Package | Purpose |
| --- | --- |
| `checker` | Reusable health checks for HTTP, TCP, SQL ping, online connectivity, readiness gates, and no-op checks. |
| `probe` | Periodically runs a checker and emits ticks. |
| `subscriber` | Best-effort tick fan-out and latest-error tracking for selected probe names. |
| `server` | Registers probes per service, creates observers, and manages lifecycle. |
| `net` | Small interfaces used to customize network dialing. |
| `sql` | Small interfaces used to customize database pinging. |

## 🔑 Key behaviors

### ⏱️ Probe and server lifecycle

- `probe.Start` performs an immediate check before the periodic loop continues.
- `probe.Start` is idempotent while running.
- `probe.Stop` is safe before start, safe to call multiple times, cancels
  in-flight checks, and waits for the probe worker to exit.
- `server.Start` waits for each service's initial checks before returning.
- `server.Start` and `server.Stop` are idempotent.
- Call `Stop` after `Start` returns, typically during process shutdown.
- A probe with an invalid period emits a single error tick and closes.

### 🧩 Registration and observers

- `server.Register` and `server.Observe` are setup-time calls; finish
  registration before calling `Start`.
- `server.Register` replaces an existing service with the same name.
- Duplicate probe names within a service replace earlier registrations.
- `server.NewOnlineRegistration` builds a registration named `online`; use
  `server.NewRegistration` with `checker.NewOnlineChecker` if you need another
  probe name.
- `subscriber.Subscriber` sends ticks on a best-effort, non-blocking channel; a
  slow observer can miss intermediate ticks.
- `subscriber.Observer` starts with `nil` errors for every tracked probe name
  and updates as ticks arrive.
- `server.Service` and `server.Server` keep observer instances across
  stop/start cycles, so existing observers continue receiving ticks after a
  restart.

### ✅ Checker defaults

- `HTTPChecker`, `TCPChecker`, `DBChecker`, and `OnlineChecker` use a default
  timeout of `30s` when you pass `0`.
- All exported checker implementations are safe for concurrent use.
- If you inject a shared transport, dialer, or pinger, that dependency should
  also be safe for the concurrent calls you expect.
- `HTTPChecker` treats HTTP status codes below `400` as healthy and wraps
  `checker.ErrInvalidStatusCode` for `4xx` and `5xx` responses.
- `DBChecker` and `TCPChecker` use `checker.ErrTimeout` as the timeout cause for
  their derived per-call contexts.
- `OnlineChecker` reports healthy if any configured URL returns `200 OK` or
  `204 No Content`.
- `ReadyChecker.Ready` is a one-way transition; once ready, that checker stays
  healthy.

## 🚦 End-to-end example

The `server` package is the usual entry point when you want one or more health
views for a service:

```go
package main

import (
	"context"
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

	if err := s.Observe(
		"payments",
		"livez",
		apiRegistration.Name,
		cacheRegistration.Name,
	); err != nil {
		log.Fatal(err)
	}

	if err := s.Observe(
		"payments",
		"readyz",
		apiRegistration.Name,
		cacheRegistration.Name,
		readyRegistration.Name,
	); err != nil {
		log.Fatal(err)
	}

	if err := s.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer s.Stop(context.Background())

	livez, err := s.Observer("payments", "livez")
	if err != nil {
		log.Fatal(err)
	}

	readyz, err := s.Observer("payments", "readyz")
	if err != nil {
		log.Fatal(err)
	}

	readyChecker.Ready()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if readyz.Error() == nil {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if err := livez.Error(); err != nil {
		log.Printf("livez unhealthy: %v", err)
	}

	for name, err := range readyz.Errors() {
		log.Printf("readyz %s = %v", name, err)
	}
}
```

> [!NOTE]
> - `Observe` validates that the service exists and every named probe was registered.
> - The first meaningful observer state arrives after the initial probe tick is processed.
> - Calling `Observe` again with the same service and kind is a no-op; it does not replace the existing observer.

## 🧪 Direct checker examples

### 🌐 HTTP checker

`HTTPChecker` performs a `GET` request. Redirects and other status codes below
`400` are considered healthy; `4xx` and `5xx` responses are unhealthy.

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

### 🌍 Online checker

`OnlineChecker` is useful when the question is "can this process reach the
outside world?" rather than "is this exact upstream healthy?"

By default it checks:

- `https://google.com/generate_204`
- `https://cp.cloudflare.com/generate_204`
- `https://connectivity-check.ubuntu.com`

Use `checker.WithURLs` to replace that default list:

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

### 🗄️ Database checker

`DBChecker` depends on the small `sql.Pinger` interface. A standard
`*database/sql.DB` satisfies that interface, so you can use it directly:

```go
var db *sql.DB

check := checker.NewDBChecker(db, 5*time.Second)

if err := check.Check(context.Background()); err != nil {
	log.Printf("database unhealthy: %v", err)
}
```

### 🚪 Readiness gate

`ReadyChecker` lets your application control readiness explicitly:

```go
ready := checker.NewReadyChecker(errors.New("warming caches"))

if err := ready.Check(context.Background()); err != nil {
	log.Printf("not ready yet: %v", err)
}

// Later, once startup work is complete.
ready.Ready()
```

## ⏱️ Manual probe usage

If you do not need the `server` package, you can work directly with a probe:

```go
p := probe.NewProbe("api", 10*time.Second, checker.NewNoopChecker())
ticks, err := p.Start(context.Background())
if err != nil {
	log.Fatal(err)
}
defer p.Stop(context.Background())

tick := <-ticks
log.Printf("%s healthy=%t", tick.Name(), tick.Error() == nil)
```

`Stop` cancels any check still running and waits for the probe goroutine to
exit, or returns the supplied context error if that wait is canceled.

## 👀 Working with observers

Observers give you both a summary and a detailed view:

```go
summary := observer.Error()  // joined error across unhealthy probes
details := observer.Errors() // copy of the latest per-probe error map
```

`Errors()` returns a copy, so callers can inspect or mutate the returned map
without affecting the observer's internal state.

When you manage multiple services, `Server.Observers(kind)` iterates the
services that have that observer kind and skips services that do not:

```go
for name, observer := range s.Observers("readyz") {
	log.Printf("%s ready=%t", name, observer.Error() == nil)
}
```

## 🛠️ Development

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
- For local `make specs` runs that need the status service, run `make start`
  before the tests and `make stop` afterward. These commands use the shared
  `bin/` Docker helper, which may require Docker, GitHub SSH access, and a
  writable sibling `../docker` checkout.

## 📚 Documentation

- Package documentation lives in `doc.go` files and exported symbol comments.
- Runnable pkg.go.dev examples live in `example_test.go` files.
- README examples should always import `github.com/alexfalkowski/go-health/v2/...`.
