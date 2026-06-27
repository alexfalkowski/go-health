// Package watcher defines health state watch subscriptions.
package watcher

// Subscription receives current and future health error snapshots.
//
// Receive returns a channel with the current error snapshot followed by future
// snapshots. Subscriptions expose state snapshots, not an event log; concrete
// implementations may coalesce pending updates to the latest error when the
// receiver is slow. Close releases the subscription and closes the receive
// channel. Callers should close the subscription when they no longer need
// updates.
type Subscription interface {
	Receive() <-chan error
	Close()
}
