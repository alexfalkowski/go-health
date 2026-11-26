# Accepted design for lifecycle

Part of the [accepted-design index](../accepted-design.md).

## Service and server setup

Complete registration before calling `Start`. A repeated service name in
`Server.Register` or probe name in `Service.Register` replaces the earlier
registration. For an existing observer kind within a service, `Observe` keeps
the original probe set. When creating a new kind, `Observe` validates the
requested names and returns an error matching `server.ErrProbeNotFound` for
unknown probes.

[Service registration](../../../server/service.go) ·
[Server registration](../../../server/server.go) ·
[Observer registration tests](../../../server/service_test.go)

## Service and server startup and shutdown

The supported lifecycle is setup, one `Start`, then one `Stop` after `Start`
returns. Repeat-call behavior is documented for specific cases, such as
calling `Start` while already running with a live context. Keep lifecycle
expectations scoped to those documented cases. Use a shutdown context that
can remain valid until cleanup completes.

`Service.Start` runs initial probe checks concurrently and waits for them to
finish. Checkers shared across registrations must support concurrent `Check`
calls. The service fans in the probe ticks and forwards them to subscribers,
so initial results may still be pending at observers when `Start` returns.

[Service lifecycle](../../../server/service.go) ·
[Server lifecycle](../../../server/server.go) ·
[Startup tests](../../../server/service_test.go)

## Direct probe use

During normal startup, `Probe.Start` waits for the initial check before
returning the tick channel. Keep receiving while the probe runs. When the
buffer fills, the worker waits for a receiver or cancellation. A non-positive
period produces a single tick wrapping `probe.ErrInvalidPeriod` and closes
the channel. `Probe.Stop` cancels in-flight checks and waits for the worker to
exit, subject to the caller's shutdown context.

[Probe lifecycle](../../../probe/probe.go) ·
[Probe tests](../../../probe/probe_test.go)
