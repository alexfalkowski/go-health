# Health Monitoring Pattern

This repository solves the health monitoring pattern in go.

## Background

To understand the background please have a read of [Health Endpoint Monitoring pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/health-endpoint-monitoring).

### Rationale

You might be asking yourself why create another health solution as there seems to be a few. You are right these are the ones I could find.

- [docker/go-healthcheck](https://github.com/docker/go-healthcheck)
- [InVisionApp/go-health](https://github.com/InVisionApp/go-health)
- [etherlabsio/healthcheck](https://github.com/etherlabsio/healthcheck)
- [heptiolabs/healthcheck](https://github.com/heptiolabs/healthcheck)
- [hellofresh/health-go](https://github.com/hellofresh/health-go)

So you are free to use any of these awesome solutions, though I had some requirements that I wanted met. These are as follows:

- The solution has to be asynchronous so that we don't DOS our dependencies (some of these solutions have this)
- The solution is free from other dependencies.
- Flexible enough to be able to implement any transport that is needed.
- Not to provide an opinionated way to do heath monitoring across transports.

### Types

The types of monitoring that we want others to build is as follows:
- White/Black box health
- [Liveliness/Readiness](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Health Check API](https://microservices.io/patterns/observability/health-check-api.html)

## Usage

To get going please add the dependency, as follows:

```sh
go get github.com/alexfalkowski/go-health
```

To uses it, please look at this example:

```go
package main

import (
    "time"

    "github.com/alexfalkowski/go-health/pkg/chk"
    "github.com/alexfalkowski/go-health/pkg/svr"
)

func main() {
    timeout := 5 * time.Second
    server := svr.NewServer()

    cc := chk.NewHTTPChecker("https://httpstat.us/200", timeout)
    hr := svr.NewRegistration("http", 0, cc)
    tc := chk.NewTCPChecker("httpstat.us:80", timeout)
    tr := svr.NewRegistration("tcp", 0, tc)

    if err := server.Register(hr, tr); err != nil {
        panic(err)
    }

    ob, err := server.Observe("http", "tcp")
    if err != nil {
        panic(err)
    }

    if err := server.Start(); err != nil {
        panic(err)
    }

    defer server.Stop()

    ob.Error() // This will update with errors. If nil everything is OK.
}
```
