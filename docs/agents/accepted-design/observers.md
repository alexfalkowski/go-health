# Accepted design for observers

Part of the [accepted-design index](../accepted-design.md).

## Tick delivery and initial state

`Subscriber` forwards matching probe ticks without blocking. When its buffer
is full, it drops the incoming tick so a slow consumer cannot stall fan-out.
An observer processes ticks in a background goroutine and protects state
reads with a read/write mutex. It starts with `nil` errors for its configured
probe names and tracks the last tick it processed for each name. An initial
healthy snapshot can therefore precede any observed check result. Consumers
that need evidence of probe execution must account for that startup window.

[Subscriber delivery](../../../subscriber/subscriber.go) ·
[Observer state](../../../subscriber/observer.go) ·
[Subscriber and observer tests](../../../subscriber/subscriber_test.go)

## Aggregate health

`Server.Error` joins the errors of services that have the requested observer
kind. When no service has that kind, `Server.Error` and `Server.Watch` return
`server.ErrObserverNotFound` so a misconfigured health view fails visibly.
`Server.Observers` supports enumeration and can yield zero observers.

[Aggregate APIs](../../../server/server.go) ·
[Server tests](../../../server/server_test.go)

## Watch subscriptions

`Watch` publishes the current error snapshot and offers further snapshots as
ticks are processed, including when health is unchanged. Pending snapshots
may be coalesced when the receiver is slow. The caller owns the subscription
and closes it when updates are no longer needed. Stopping observation, a
service, or the server leaves existing watch channels open for the caller to
close.

[Watch contract](../../../watcher/subscription.go) ·
[Observer watch](../../../subscriber/observer.go) ·
[Server watch](../../../server/server.go)
