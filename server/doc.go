// Package server orchestrates health probes and observers for one or more services.
//
// A service is registered with one or more probe registrations (name + period + checker).
// You then create one or more observers (identified by a "kind", e.g. "livez" or "readyz")
// that track a subset of those probe names.
//
// Typical usage:
//
//	s := server.NewServer()
//
//	httpChecker := checker.NewHTTPChecker("https://example.com/health", 5*time.Second)
//	httpReg := server.NewRegistration("http", 500*time.Millisecond, httpChecker)
//
//	s.Register("myservice", httpReg)
//	_ = s.Observe("myservice", "livez", httpReg.Name)
//	s.Start()
//	defer s.Stop()
package server
