// Package net contains small network-related interfaces used by the health
// checkers.
//
// The package exists to decouple checker implementations from the concrete types
// in the standard library. Callers can pass the default adapter in production or
// provide a small fake in tests.
package net
