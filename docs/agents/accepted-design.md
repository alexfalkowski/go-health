# Accepted design

This index records go-health's accepted design choices, tradeoffs, and support
boundaries. The area files carry the same weight as [AGENTS.md](../../AGENTS.md).

Read the relevant area files before proposing a change or recording a finding
or audit entry. Revisit a decision when the owner asks or new evidence changes
its documented contract or tradeoff.

- [Checkers](accepted-design/checkers.md) covers SQL query execution, HTTP
  redirects, external connectivity checks, and the readiness gate.
- [Lifecycle](accepted-design/lifecycle.md) covers setup-time registration,
  startup, and shutdown for probes, services, and servers.
- [Observers](accepted-design/observers.md) covers tick delivery, initial
  health state, aggregate errors, and watch subscriptions.
- [Testing](accepted-design/testing.md) covers intentional live network
  integration tests.
