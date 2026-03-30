// Package subscriber provides fan-out and state tracking for probe ticks.
//
// Subscriber is the transport layer for ticks inside a service: it accepts probe
// results for a configured set of names and forwards matching ticks to a channel.
// Observer is the stateful layer on top of that channel: it consumes ticks in a
// goroutine and remembers the latest error for each tracked probe.
//
// This package is typically used indirectly through the server package, but the
// types are also useful when you want to build custom aggregation or expose
// health state yourself.
//
// # Error views
//
// Observer exposes two views of the current state:
//
//   - Error returns a joined summary error of all unhealthy probes.
//   - Errors returns a copy of the per-probe error map.
//
// Joined errors are annotated with the probe name so they are still useful when
// logged or returned directly.
//
// # Delivery semantics
//
// Subscriber.Send is best-effort and non-blocking. If the buffer is full, the
// tick is dropped rather than back-pressuring the producer. This keeps probe
// execution decoupled from slow observers.
//
// # Example
//
//	sub := subscriber.NewSubscriber([]string{"db"})
//	ob := subscriber.NewObserver([]string{"db"}, sub)
//
//	sub.Send(probe.NewTick("db", errors.New("ping failed")))
//	sub.Close()
//	ob.Wait()
//
//	fmt.Println(ob.Error() != nil)
package subscriber
