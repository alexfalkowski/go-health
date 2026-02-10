// Package net contains small network-related interfaces and adapters.
//
// This package exists to decouple health check implementations from the standard
// library's concrete networking types. By depending on small interfaces (for
// example, Dialer) other packages can accept custom implementations in tests or
// alternate transports without pulling in heavier dependencies.
//
// The checker package uses these interfaces to perform TCP connectivity checks
// and to allow injecting custom dialing behavior.
//
// # Defaults
//
// Where appropriate, this package provides default implementations backed by the
// Go standard library (for example, DefaultDialer).
package net
