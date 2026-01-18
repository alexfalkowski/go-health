# AGENTS.md

## Project overview

This repository is a Go library implementing a **health monitoring pattern** (asynchronous checks, observers, and status aggregation).

- Module: `github.com/alexfalkowski/go-health/v2` (see `go.mod:1`)
- Go version: `go 1.25.0` (see `go.mod:3`)
- CI: CircleCI (`.circleci/config.yml`)

## Repository layout

Top-level packages:

- `checker/`: health check implementations (HTTP/TCP/DB/online/ready/noop).
- `probe/`: periodic execution of a `checker.Checker` producing ticks.
- `subscriber/`: subscription + observer state tracking of probe ticks.
- `server/`: orchestration layer: register probes per service, create observers, start/stop.
- `net/`: small interfaces/wrappers (e.g., `Dialer`).
- `sql/`: small interfaces/wrappers (e.g., `Pinger`).
- `internal/test/`: test helpers (e.g., building status URLs from env).

Notable non-code:

- `bin/`: **git submodule** with shared build tooling (see `.gitmodules`).
- `.golangci.yml`: golangci-lint configuration.
- `.gosec`: gosec exclusions.
- `test/`: test configs and report outputs (e.g., `test/reports/`).

## Build / test / lint commands

Most workflows are driven via `make`. The root `Makefile` only includes targets from the `bin/` submodule:

- `Makefile:1-2` includes `bin/build/make/go.mak` and `bin/build/make/git.mak`.

### Submodule setup

Many `make` targets depend on scripts under `./bin/...`.

- Initialize and update submodule:

```sh
git submodule sync
git submodule update --init
```

CircleCI runs these steps explicitly (`.circleci/config.yml:15-16`).

### Dependency management

From `bin/build/make/go.mak`:

```sh
make dep          # go mod download + tidy + vendor
make clean-dep    # go clean -cache -testcache -fuzzcache -modcache
make clean-lint   # clear golangci-lint cache via bin script
make clean        # repo-specific clean via ./bin/build/go/clean
```

Notes:

- `make dep` creates/updates `vendor/` and many commands run with `-mod vendor`.

### Testing

From `bin/build/make/go.mak`:

```sh
make specs
```

`make specs` runs `gotestsum` and then `go test` with:

- `-race`
- `-mod vendor`
- coverage output to `test/reports/profile.cov` and junit XML to `test/reports/specs.xml` (`bin/build/make/go.mak:62-63`).

Coverage helpers:

```sh
make coverage      # generates test/reports/coverage.html and prints func coverage
make html-coverage
make func-coverage
```

### Linting / formatting

Lint targets (`bin/build/make/go.mak`):

```sh
make lint          # fieldalignment + golangci-lint
make fix-lint      # attempts auto-fixes (fieldalignment + golangci-lint --fix)
```

Formatting:

```sh
make format        # go fmt ./...
```

`golangci-lint` configuration is in `.golangci.yml` (default: all linters; several disabled).

### Security

```sh
make sec           # govulncheck -show verbose -test ./...
```

Gosec exclusions are listed in `.gosec`.

### Other useful make targets (if tools are installed)

From `bin/build/make/go.mak`:

```sh
make benchmark package=<pkgdir>
make benchmark-pprof
make outdated-dep
make encode-config kind=<name>
make create-certs
make create-diagram package=<pkgdir>
make analyse
make money
make start
make stop
```

Note: `make start/stop` shells out to `bin/build/docker/env` (`bin/build/make/go.mak:130-135`). That script clones/updates `git@github.com:alexfalkowski/docker.git` (`bin/build/docker/env:10-16`), which requires SSH access to GitHub.

## Testing gotchas

- Some tests make real network calls (e.g., `server/server_test.go` uses `https://www.google.com/`).
- Tests also hit a local “status” HTTP service via `internal/test.StatusURL`:
  - URL built from `STATUS_PORT` env var, defaulting to `6000` (`internal/test/test.go:9-13`).
  - CircleCI runs an `alexfalkowski/status:latest` container exposing that service (`.circleci/config.yml:7-10`).

If tests fail locally with connection errors to `http://localhost:6000/...`, you likely need an equivalent status service running or to set `STATUS_PORT` appropriately.

## Code conventions and patterns (observed)

### Package boundaries

- Tests use an external test package name (`package server_test`) to validate public API behavior (see `server/server_test.go:1`).

### Errors

- Errors are wrapped with context using `fmt.Errorf("<context>: %w", err)` in checkers (e.g., `checker/http.go:35-53`, `checker/tcp.go:33-39`, `checker/db.go:29-31`).
- Aggregation uses `errors.Join` for combining probe errors (`subscriber/errors.go:13-26`).

### Options pattern

- Checkers support functional options via an `Option` interface + `optionFunc` (`checker/options.go:9-45`).
- `parseOptions` provides defaults (HTTP transport, dialer, and online-check URLs) (`checker/options.go:47-67`).

### Concurrency model

- `probe.Probe` periodically emits `*probe.Tick` to a buffered channel; it performs an immediate check on startup (`probe/probe.go:27-41`).
- `server.Service` merges tick channels from probes into a single fan-in channel and forwards ticks to subscribers (`server/service.go:58-117`).
- `subscriber.Observer` runs a goroutine reading ticks and updating an internal `Errors` map protected by an RWMutex (`subscriber/observer.go:6-46`).

### Formatting / whitespace

- `.editorconfig` indicates Go files use tabs (`.editorconfig:14-16`); Makefiles use tabs.

## CI notes (CircleCI)

CircleCI job flow (`.circleci/config.yml`):

- Updates submodules
- Runs:
  - `make source-key`
  - `make clean`
  - `make dep`
  - `make lint`
  - `make sec`
  - `make specs`
  - `make coverage`
  - `make codecov-upload`

Test reports are stored under `test/reports/`.

## Documentation mismatch to be aware of

- The README usage snippet shows `go get github.com/alexfalkowski/go-health` and imports `github.com/alexfalkowski/go-health/...` (`README.md:45-59`), while `go.mod` declares the module path with `/v2` (`go.mod:1`). Use the module path observed in `go.mod` when adding imports in this repository.
