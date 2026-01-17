# AGENTS.md

## Repository overview

This is a Go library implementing a health monitoring pattern (async probes + observers).

Top-level packages:

- `checker/`: check implementations (HTTP/TCP/DB/online/ready/noop/timeout)
- `probe/`: periodic probe runner that emits ticks
- `subscriber/`: subscriber/observer that tracks probe errors
- `server/`: orchestration (services, registrations, observers)
- `net/`: dialer abstraction used by checkers
- `sql/`: DB “pinger” helper interface
- `internal/test/`: test helpers (e.g. status service URL)

Module: `github.com/alexfalkowski/go-health/v2` (`go.mod:1`).

## First-time setup (important)

This repo depends on the `bin/` git submodule for most `make` targets.

```sh
git submodule sync
git submodule update --init
```

The top-level `Makefile` only includes `bin/build/make/go.mak` and `bin/build/make/git.mak` (`Makefile:1-2`).

## Essential commands

These commands are *observed* in `bin/build/make/go.mak` and CI (`.circleci/config.yml`).

### Dependencies

```sh
make dep
```

- Runs `go mod download`, `go mod tidy`, `go mod vendor` (`bin/build/make/go.mak:9-26`).
- Many targets run with `-mod vendor`; run `make dep` after changing dependencies.

### Tests

CI-style test run:

```sh
make specs
```

- Uses `gotestsum` with `-race` and writes reports/coverage under `test/reports/` (`bin/build/make/go.mak:61-64`).

Local quick run:

```sh
go test ./...
```

### Lint

```sh
make lint
```

- Runs field-alignment check (`bin/build/make/go.mak:39-44`).
- Runs `golangci-lint` via `bin/build/go/lint` (`bin/build/make/go.mak:45-53`).
- Linter/formatter config is in `.golangci.yml`.

Auto-fix (where supported):

```sh
make fix-lint
```

### Security

```sh
make sec
```

- Runs `govulncheck -show verbose -test ./...` (`bin/build/make/go.mak:95-98`).
- `.gosec` suppressions: `G104,G307`.

### Coverage artifacts

```sh
make coverage
```

- Generates HTML + function coverage outputs (`bin/build/make/go.mak:76-86`).
- Codecov thresholds are in `.codecov.yml`.

### Other useful Make targets

From `bin/build/make/go.mak` (may require extra tools installed):

- `make benchmark package=<pkg>` (`bin/build/make/go.mak:65-72`)
- `make create-diagram package=<pkg>` (uses `goda` + `dot`) (`bin/build/make/go.mak:118-121`)
- `make create-certs` (uses `mkcert`) (`bin/build/make/go.mak:113-117`)
- `make encode-config kind=<name>` (base64 encodes `test/<name>.yml`) (`bin/build/make/go.mak:109-112`)

## CI behavior (CircleCI)

CircleCI runs these steps (`.circleci/config.yml:23-59`):

- `make clean`
- `make dep`
- `make lint`
- `make sec`
- `make specs`
- `make coverage`
- `make codecov-upload`

CI also starts a “status” service container (`alexfalkowski/status:latest`) with `command: server -i env:CONFIG` and a base64-encoded config in `CONFIG` (`.circleci/config.yml:6-10`).

## Testing gotchas

### Local status service dependency (`localhost:6000`)

Several tests expect an HTTP service that responds on `http://localhost:<port>/v1/status/<code>`.

- The helper `internal/test.StatusURL` builds this URL and defaults to port `6000`; it can be overridden with `STATUS_PORT` (`internal/test/test.go:9-13`).
- CI provides the service via a container (`.circleci/config.yml:6-10`).
- A sample config exists at `test/status.yml`.

If nothing is listening, `go test ./...` fails with `connect: connection refused` (e.g. `server/server_test.go`).

### External network dependency

Some tests hit external endpoints (e.g. `https://www.google.com/`, `httpstat.us`) (`server/server_test.go:28-33`, `291-295`). Running the full suite may require outbound network access.

## Code organization & patterns

### High-level flow

- `server.Server` holds named `Service`s and controls start/stop (`server/server.go:15-99`).
- `server.Service` registers probes, merges their tick channels, and broadcasts ticks to subscribers (`server/service.go:14-140`).
- `probe.Probe` runs a `checker.Checker` periodically and sends `*probe.Tick` on a buffered channel (`probe/probe.go:16-86`).
- `subscriber.Observer` maintains an `Errors` map and updates it by receiving ticks from a `Subscriber` (`subscriber/observer.go:5-93`).

### Interfaces & options

- Core interface: `checker.Checker` with `Check(ctx context.Context) error` (`checker/checker.go:5-8`).
- Checkers use an options pattern (`checker/options.go`) for injecting dependencies:
  - `WithRoundTripper` for HTTP checkers
  - `WithDialer` for TCP checkers
  - `WithURLs` for online checker URL list

### Error wrapping style

Checkers wrap errors with a component prefix, e.g. `fmt.Errorf("http checker: %w", err)` (`checker/http.go:33-53`).

### Concurrency & shutdown notes

- `subscriber.Observer` starts a goroutine in `Start()` and can be stopped via `Stop()` (`subscriber/observer.go:28-62`).
- `server.Service.Stop()` calls `Observer.Stop()` and closes subscribers (`server/service.go:75-100`).
- `probe.Probe.Start()` is idempotent; `Stop()` is guarded and won’t double-close (`probe/probe.go:29-63`).

## Lint/style conventions

- `golangci-lint` config: `.golangci.yml` (default: all; many disabled).
- Formatters are enabled via golangci-lint (`.golangci.yml:31-36`).

## Submodule notes

- `bin/` is a git submodule (`.gitmodules:1-3`).
- If `make` targets fail due to missing scripts, initialize/update the submodule first.

## Where to look first when changing behavior

- Add a new check type: implement `checker.Checker` under `checker/` and wire it via `server.Registration` helpers (see `server/registration.go:9-21`).
- Probe scheduling and tick emission: `probe/probe.go:29-86`.
- Aggregation/fan-out: `server/service.go:58-140`.
- Observer error semantics: `subscriber/errors.go` and `subscriber/observer.go`.
