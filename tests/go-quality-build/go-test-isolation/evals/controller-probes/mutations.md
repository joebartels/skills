# Frozen isolation mutation meanings

These meanings are controller-only and must not be dispatched to authors.
Adapt a compiling single-defect patch to the submitted implementation/test
fixture, preserving unrelated semantics. Run candidate tests without these
probes; distinguish test-fixture damage, missing assertions and production
defects. Do not count unrelated panic, compile rejection or whole-tool timeout.

- `store-shared-root`: independent Store instances accidentally use one shared
  filesystem root, losing same-key independence.
- `fixture-parent-teardown`: a grouped parent removes a resource at parent-body
  return rather than after its parallel descendants complete; verify useful
  I/O/lifetime assertions report the failure.
- `environment-leak`: a test's environment helper changes a variable without
  restoration; inspect before/after process state and individual/shuffled runs.
  This mutates test-owned state, not LoadFromEnv's read-only contract.
- `http-ignore-cancellation`: the request uses a background context rather than
  the caller; started in-flight work must observe cancellation promptly.
- `http-body-not-closed`: an acquired response body's Close call is omitted;
  the submitted observable-lifetime test should report it.
- `poll-discard-context`: Poll supplies a background context to callbacks,
  discarding the caller's cancellation/lifetime.
- `poll-start-relative`: replace completion-relative timers with one ticker
  started before callbacks; hold the first callback beyond the interval.

Worker failure reporting is assessed through test-goroutine ownership and
bounded joined completion, not by causing an unrelated testing.Fatal crash.
