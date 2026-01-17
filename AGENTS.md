# AGENTS.md

## Repository overview

This is a Go library implementing a health monitoring pattern (async probes + observers). The public API is primarily under:

- `checker/`: check implementations (HTTP/TCP/DB/online/ready/noop/timeout)
- `probe/`: periodic probe runner that emits ticks
- `subscriber/`: subscriber/observer that tracks probe errors
- `server/`: orchestration (services, registrations, observers)
- `net/`: dialer abstraction used by checkers
- `sql/`: DB “pinger” helper

Module: `github.com/alexfalkowski/go-health/v2` (`go.mod:1`).

## First-time setup (important)

This repo depends on the `bin/` git submodule for most `make` targets.

```sh
git submodule sync
git submodule update --init
```

The top-level `Makefile` only includes `bin/build/make/go.mak` and `bin/build/make/git.mak` (`Makefile:1-2`).

## Essential commands

All commands below are *observed* in `bin/build/make/go.mak` and CI (`.circleci/config.yml`).

### Dependencies

```sh
make dep
```

- Runs `go mod download`, `go mod tidy`, `go mod vendor` (`bin/build/make/go.mak:9-26`).
- Many targets run with `-mod vendor`; run `make dep` after changing dependencies.

### Tests

```sh
make specs
```

- Runs tests via `gotestsum` with `-race` and coverage output to `test/reports/` (`bin/build/make/go.mak:61-64`).

Direct Go test (when you don’t need the CI-style reports):

```sh
go test ./...
```

### Lint

```sh
make lint
```

- Runs a “field alignment” check (`bin/build/make/go.mak:39-56`).
- Runs `golangci-lint` via `bin/build/go/lint` (`bin/build/make/go.mak:45-53`).
- Formatting is configured via `.golangci.yml` (enables `gci`, `gofmt`, `gofumpt`, `goimports`).

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

CI runs:

```sh
make coverage
```

- Generates HTML + function coverage outputs (`bin/build/make/go.mak:76-86`).
- Coverage thresholds configured in `.codecov.yml`.

### Misc (available via Make targets)

From `bin/build/make/go.mak`, these targets exist but may require extra tools installed:

- `make benchmark package=<pkg>` (`bin/build/make/go.mak:65-72`)
- `make create-diagram package=<pkg>` (uses `goda` + `dot`) (`bin/build/make/go.mak:118-121`)
- `make money` (uses `scc`) (`bin/build/make/go.mak:126-128`)
- `make analyse` (uses `gsa`) (`bin/build/make/go.mak:122-125`)
- `make create-certs` (uses `mkcert`) (`bin/build/make/go.mak:113-117`)

## CI behavior (CircleCI)

CircleCI runs these steps in order (`.circleci/config.yml:23-59`):

- `make clean`
- `make dep`
- `make lint`
- `make sec`
- `make specs`
- `make coverage`
- `make codecov-upload`

CI also starts a “status” service container (`alexfalkowski/status:latest`) on `tcp://:6000` with a base64-encoded config (`.circleci/config.yml:6-10`).

## Testing gotchas

### Local status service dependency (`localhost:6000`)

Some tests expect an HTTP status service on `http://localhost:6000` (e.g. `server/server_test.go:132-165`, `268-323`). If nothing is listening on `:6000`, those tests fail with `connect: connection refused`.

A config file for that service exists at `test/status.yml`.

CI satisfies this by starting a container (`alexfalkowski/status:latest`) with `command: server -i env:CONFIG` and a base64-encoded config in `CONFIG` (`.circleci/config.yml:6-10`).

There is also a `make start`/`make stop` pair (`bin/build/make/go.mak:130-136`), but it uses `bin/build/docker/env` which clones/pulls `git@github.com:alexfalkowski/docker.git` (`bin/build/docker/env:10-19`) and may require GitHub SSH access and Docker.

### External network dependency

Some tests hit external endpoints (e.g. `https://www.google.com/`, `httpstat.us`) (`server/server_test.go:28-33`, `291-295`), so running the full suite may require outbound network access.

## Code organization & patterns

### High-level flow

- A `server.Server` holds named `Service`s and controls start/stop (`server/server.go:16-99`).
- A `Service` registers probes, merges their tick channels, and broadcasts ticks to subscribers (`server/service.go:14-117`).
- A `probe.Probe` runs a `checker.Checker` periodically and sends `*probe.Tick` on a buffered channel (`probe/probe.go:11-67`).
- A `subscriber.Observer` maintains an `Errors` map and updates it by receiving ticks from a subscriber (`subscriber/observer.go:5-47`).

### Interfaces & options

- Core interface: `checker.Checker` with `Check(ctx context.Context) error` (`checker/checker.go:5-8`).
- Checkers use an options pattern (`checker/options.go`) for injecting dependencies:
  - `WithRoundTripper` for HTTP checkers
  - `WithDialer` for TCP checkers
  - `WithURLs` for online checker URL list

### Error wrapping style

Checkers generally wrap errors with a component prefix, e.g. `fmt.Errorf("http checker: %w", err)` (`checker/http.go:33-53`) and `fmt.Errorf("tcp checker: %w", err)` (`checker/tcp.go:28-39`).

## Lint/style conventions

- `golangci-lint` is used with “default: all” and a set of disabled linters (`.golangci.yml:1-20`).
- Formatters are enabled via golangci-lint (`.golangci.yml:31-36`).

Follow existing style:

- Tabs for indentation (standard gofmt output).
- Keep APIs small and dependency-free; most packages use only stdlib.

## Submodule notes

- `bin/` is a git submodule (`.gitmodules:1-3`).
- If `make` targets fail due to missing scripts, initialize/update the submodule first.

## Where to look first when changing behavior

- Add a new check type: implement `checker.Checker` under `checker/` and wire it via `server.Registration` helpers (see `server/registration.go:9-21`).
- Probe scheduling and tick emission: `probe/probe.go:27-67`.
- Aggregation and fan-out: `server/service.go:58-117`.
- Observer semantics (first error vs all errors): `subscriber/observer.go:25-39` and `subscriber/errors.go`.
