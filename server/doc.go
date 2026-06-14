// Package server orchestrates probes and observers for one or more services.
//
// This is the highest-level package in the module. It wires together the lower
// layers:
//
//   - checker defines how to check a dependency
//   - probe schedules those checks
//   - subscriber tracks the latest state for selected probes
//
// A service is configured with one or more Registration values. Each
// registration gives the probe a name, a schedule, and a checker. You then
// create one or more observers, usually named after an endpoint such as "livez"
// or "readyz", that watch a subset of those probe names.
//
// Register services and observers during setup, then call Start to run the
// orchestration. Start waits for initial checks to finish or the supplied context
// to be canceled before returning. Observer state is updated asynchronously from
// probe ticks, so Start completion does not mean observers have processed the
// initial result. Call Stop after Start has returned, typically from process
// shutdown. Existing observers continue to work across service restarts and
// retain their previous state until new ticks arrive.
//
// # Example
//
//	s := server.NewServer()
//
//	httpChecker := checker.NewHTTPChecker("https://example.com/health", 5*time.Second)
//	httpReg := server.NewRegistration("http", 500*time.Millisecond, httpChecker)
//
//	s.Register("payments", httpReg)
//
//	if err := s.Observe("payments", "livez", httpReg.Name); err != nil {
//		log.Fatal(err)
//	}
//
//	if err := s.Start(context.Background()); err != nil {
//		log.Fatal(err)
//	}
//	defer s.Stop(context.Background())
//
//	ob, err := s.Observer("payments", "livez")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err := ob.Error(); err != nil {
//		log.Printf("payments unhealthy: %v", err)
//	}
package server
