// Package net contains small network-related interfaces used by the health
// checkers.
//
// The package exists to decouple checker implementations from the concrete types
// in the standard library. Callers can pass the default adapter in production or
// provide a small fake in tests.
//
// # Example
//
//	import stdnet "net"
//
//	type fakeDialer struct{}
//
//	func (fakeDialer) DialContext(ctx context.Context, network, address string) (stdnet.Conn, error) {
//		return nil, errors.New("dial failed")
//	}
//
//	checker.NewTCPChecker("cache:6379", time.Second, checker.WithDialer(fakeDialer{}))
package net
