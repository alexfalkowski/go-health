# Accepted design for checkers

Part of the [accepted-design index](../accepted-design.md).

## SQL query execution

`DBChecker` runs `SELECT 1` and reads the result using the application's
existing `*database/sql.DB`. A successful connection ping can still leave an
application unable to execute a query or read a row. The caller manages the
database handle's lifetime. Databases that require a different query should
use a custom `checker.Checker`.

[Database checker](../../../checker/db.go) ·
[Server integration tests](../../../server/server_test.go)

## HTTP redirects

`HTTPChecker` evaluates the direct response to a GET. It stops at redirects
to keep configured headers, such as health tokens, scoped to the configured
endpoint. Only `2xx` responses are healthy. Redirects and other non-`2xx`
responses return an error wrapping `checker.ErrInvalidStatusCode`.

[HTTP checker](../../../checker/http.go) ·
[Redirect and status tests](../../../checker/http_test.go)

## External connectivity

`OnlineChecker` checks public connectivity by default. It checks the configured
URLs concurrently. The first `200 OK` or `204 No Content` response triggers
cancellation of the remaining requests and is enough to report healthy.
Services that require specific endpoints can replace the default list with
`checker.WithURLs`.

[Online checker](../../../checker/online.go) ·
[URL option](../../../checker/options.go) ·
[Online checker tests](../../../checker/online_test.go)

## Readiness gate

`ReadyChecker` reports its configured error until `Ready` is called. A `nil`
initial error means it starts healthy. `Ready` is a one-way transition for the
lifetime of the checker. Use a custom checker for readiness that must change in
both directions.

[Readiness checker](../../../checker/ready.go) ·
[Readiness example](../../../checker/example_test.go)
