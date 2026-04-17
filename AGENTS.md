# AGENTS.md

## Shared skill

Use the shared `coding-standards` skill from `./bin/skills/coding-standards`
for cross-repository coding, review, testing, documentation, and PR
conventions. Treat this `AGENTS.md` as the repo-specific companion to that
skill.

## Project overview

This repository is a Go library for asynchronous health monitoring. It lets you
build reusable health checks, schedule them periodically, aggregate their latest
results, and expose multiple health views such as `livez` and `readyz`.

- Module path: `github.com/alexfalkowski/go-health/v2`
- Go version: `1.26.0`
- Primary CI: CircleCI via `.circleci/config.yml`

## Repository layout

Top-level packages:

- `checker/`: concrete health check implementations and checker options.
- `probe/`: periodic execution of a `checker.Checker`, producing `probe.Tick` values.
- `subscriber/`: best-effort tick fan-out plus observer state tracking.
- `server/`: orchestration for registering probes per service, creating observers, and managing start/stop lifecycle.
- `net/`: small network-related interfaces used by checkers.
- `sql/`: small SQL-related interfaces used by checkers.
- `internal/test/`: test helpers such as building URLs for the local status service.

Notable non-code paths:

- `README.md`: user-facing overview and examples.
- `Makefile`: includes targets from the `bin/` submodule.
- `bin/`: git submodule with shared build tooling.
- `.circleci/config.yml`: CI workflow and service containers.
- `.golangci.yml`: lint configuration, formatter configuration, and gosec exclusions.
- `test/`: test inputs, certificates, and generated reports under `test/reports/`.

## Runtime model

The library is layered. This is the mental model to keep in sync when editing
docs or examples:

1. A `checker.Checker` runs one synchronous health check and returns `nil` or an error.
2. A `probe.Probe` runs a checker immediately on `Start`, then periodically on its configured interval.
3. A `subscriber.Subscriber` forwards matching probe ticks on a best-effort, non-blocking channel.
4. A `subscriber.Observer` tracks the latest error for each configured probe name.
5. A `server.Service` wires probes to subscribers and observers for one service.
6. A `server.Server` manages multiple services.

Important behavior details:

- `probe.Start` waits for the initial check to finish before returning the tick channel.
- A probe with a non-positive period emits one error tick wrapping `probe.ErrInvalidPeriod` and then closes.
- `subscriber.Observer` initializes tracked probe names with `nil` errors until ticks arrive.
- `server.Service` and `server.Server` preserve observer instances across stop/start cycles.
- `server.Server.Observe` and `server.Service.Observe` are idempotent for an existing observer kind; they do not replace the original probe set.

## Build, test, and lint

Most workflows go through `make`. The root `Makefile` includes:

- `bin/build/make/go.mak`
- `bin/build/make/git.mak`

### Submodule setup

Many `make` targets rely on scripts in `bin/`, so initialize the submodule first:

```sh
git submodule sync
git submodule update --init
```

CircleCI runs those commands before any build steps.

### Common commands

```sh
make dep          # go mod download + tidy + vendor
make clean-dep    # go clean -cache -testcache -fuzzcache -modcache
make clean-lint   # clear golangci-lint cache through the bin helper
make clean        # repo-specific clean helper from bin/build/go/clean

make format       # go fmt ./...
make lint         # fieldalignment + golangci-lint
make fix-lint     # auto-fix what can be fixed
make sec          # govulncheck -show verbose -test ./...
make specs        # gotestsum + go test -race -mod vendor + coverage
make coverage     # generate HTML and function coverage summaries
```

`make specs` writes reports to:

- `test/reports/specs.xml`
- `test/reports/profile.cov`
- `test/reports/final.cov` as part of the coverage post-processing flow used by the build tooling

### Useful focused verification

When you only change docs, examples, or small package behavior, targeted `go test`
commands are often faster than the full `make specs` run. Examples:

```sh
go test ./checker ./probe ./subscriber ./net ./sql
go test ./probe -run Example -count=1
go test ./server -run TestRestartKeepsObserverReceivingTicks -count=1
```

Use the vendored or CI-style flow when you want parity with the main build.

## Testing gotchas

- Some tests make real network calls, including requests to `https://www.google.com/`.
- Some tests expect a local status service on `STATUS_PORT`; if unset, the default is `6000`.
- `internal/test.StatusURL` builds URLs like `http://localhost:<port>/v1/status/<code>`.
- CircleCI starts `alexfalkowski/status:latest` to provide that service during the `build` job.

If tests fail with connection errors to `localhost:6000`, the status service is
probably not running locally.

## CI notes

The main CircleCI `build` job does the following, in order:

1. Checks out the repository.
2. Syncs and initializes the `bin/` submodule.
3. Runs `make source-key`.
4. Restores caches.
5. Runs `make clean`.
6. Runs `make dep`.
7. Runs `make lint`.
8. Runs `make sec`.
9. Runs `make specs`.
10. Runs `make coverage`.
11. Uploads coverage and stores test reports.

There are also `sync`, `version`, and `wait-all` jobs in `.circleci/config.yml`.

## Code conventions and patterns

### Errors

- Checkers wrap underlying failures with context, usually using `fmt.Errorf("<checker>: %w", err)`.
- `checker.DBChecker` and `checker.TCPChecker` use `checker.ErrTimeout` as the cause for their derived timeout contexts.
- Aggregated observer errors use `errors.Join`.
- `subscriber.Errors.Error` annotates each joined error with the probe name.

### Options pattern

- Checker options are implemented through the `checker.Option` interface and `optionFunc`.
- `checker.WithRoundTripper` affects `HTTPChecker` and `OnlineChecker`.
- `checker.WithDialer` affects `TCPChecker`.
- `checker.WithURLs` replaces the `OnlineChecker` default URL list.

### Concurrency model

- `probe.Probe` runs its scheduling loop in a goroutine and cancels in-flight checks on `Stop`.
- `subscriber.Subscriber` is intentionally best-effort and may drop ticks if its buffer is full.
- `subscriber.Observer` consumes subscriber ticks in a goroutine and protects reads with an RW mutex.
- `server.Service` fans in probe tick channels and then fans ticks out to subscribers.

### Formatting

- `.editorconfig` uses tabs for `*.go` files and `Makefile`, spaces elsewhere.
- Follow standard Go doc comment style: start comments with the identifier name and describe observable behavior, not implementation trivia.

## Documentation maintenance

When changing code, keep these documentation surfaces aligned:

1. Package docs in `doc.go` files.
2. Exported symbol comments in package source files.
3. Runnable examples in `example_test.go`.
4. Top-level usage guidance in `README.md`.
5. Agent-specific maintenance guidance in this file.

Documentation expectations:

- Use the module path with `/v2` in all README and Go doc examples.
- Prefer examples that compile and run locally without external services when possible.
- If an example depends on asynchronous state, make the wait explicit in the example.
- Call out important defaults such as the `30s` timeout and `STATUS_PORT=6000`.
- Mention behaviors that are easy to miss, such as observer state starting at `nil` and restart-safe observers.

## External tooling notes

- `make start` and `make stop` shell out to `bin/build/docker/env`.
- That helper clones or updates the sibling `../docker` repository via SSH: `git@github.com:alexfalkowski/docker.git`.
- Expect those commands to require GitHub SSH access and a writable parent directory.

## Scope note

The `bin/` directory is a git submodule. Treat changes there as changes to shared
build tooling, not to this library itself. Update the top-level docs in this
repository by default unless the task explicitly includes submodule maintenance.
