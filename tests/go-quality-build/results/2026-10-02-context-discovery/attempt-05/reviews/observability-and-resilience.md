## Observability & Resilience — B
Scope: Completed requested `FetchAll` caller-budget and failure-signaling behavior in a Go 1.22 library, excluding unrelated legacy policies and unspecified service telemetry.
Coverage: Total and stage budgets, earlier caller deadline, active request/body cancellation, independent failure/cause identity, successful completion decisions, stage cleanup, and owned-body cleanup were assessed. No retries are introduced and no service/probe/logger configuration is implicated.
Rationale: F1 is one moderate shared diagnostic-contract failure: an observed custom cancellation can lose its standard classification. Confirmed budget containment and cleanup otherwise support the scoped operation.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [G1] `candidate/stages.go:14` and `:18` bound the whole sequence and each stage without extending an earlier caller deadline; absolute-deadline diagnostics passed.
- [G2] `candidate/stages.go:24` passes the live stage context to the HTTP request. A real local HTTP response-body read is interrupted on stage deadline, and the local handler observes cancellation.
- [G3] `candidate/stages.go:38` and `:45` preserve independent status/read/close failures rather than discarding cleanup errors. Inspectable error checks pass for stable cancellation states.

Bad

- [F1][moderate][existing-in-scope] `candidate/stages.go:61` can report a custom cancellation cause without the standard cancellation classification because the two observations occur at different times. A caller relying on `errors.Is(err, context.Canceled)` can treat a cancellation-associated failure as ordinary dependency failure. The reached helper interleaving is verified; retry or operator behavior in external consumers is not asserted. Primary owner: Correctness & Compatibility.

Suggested changes

- [F1] Sample the classification once, then join its stable cause only when cancellation was observed. Verify the transition invariant alongside the existing synchronous cause test; the disposable correction passed.

Limits: Direct cancellation claims are confined to the exercised custom bodies/transport and local standard HTTP boundary. Arbitrary client transports/body implementations may not cooperate with context; no remote cessation guarantee is made. No supplied service SLO, telemetry, or retry policy is necessary for this local contract. Exact executed/supplied results are in `evidence.md`.
