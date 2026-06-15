// Package watcher defines health state watch subscriptions.
package watcher

// Subscription receives current and future health error snapshots.
//
// Receive returns a channel with the current error snapshot followed by future
// snapshots. Close releases the subscription and closes the receive channel.
type Subscription interface {
	Receive() <-chan error
	Close()
}
