## Architecture & Design — B
Scope: `FetchAll`'s completed requested public error, cancellation, dependency, and ownership boundaries; small Go 1.22 library. Unrelated legacy policies are excluded.
Coverage: Signature preservation, caller-supplied client/context, total/stage lifetime composition, borrowed client/transport, owned bodies, partial-result boundary, and inspectable error contract were assessed. No unrelated package layering or release architecture is assumed.
Rationale: One moderate shared issue F1 makes the public cancellation error representation inconsistent during a supported concurrent transition. The design otherwise fits the small synchronous operation and needs no new interface or package structure.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [G1] `candidate/stages.go:13` preserves the exact exported signature and visible caller-configured dependencies. It borrows the existing client/transport, avoids hidden globals, and launches no unowned background work.
- [G2] `candidate/stages.go:14`, `:18`, `:43`, and `:50` align budget and body ownership with each stage. Direct diagnostics verify lifecycle and prefix semantics at the caller boundary.

Bad

- [F1][moderate][existing-in-scope] `candidate/stages.go:61` can expose a custom cancellation cause without `context.Canceled`, weakening the README's meaningful error-classification boundary for callers using `errors.Is`. Real timer-context and deterministic transition checks substantiate the representation mismatch. Primary owner: Correctness & Compatibility.

Suggested changes

- [F1] Make a single cancellation-state decision before exposing its cause; retain the existing API and independent error exposure. The disposable correction demonstrates this can be repaired locally without redesign.

Limits: No external consumer or release-version policy is supplied; signature equality and the packet's explicit README contract define this assessment. The public stress check did not observe F1; no production frequency is claimed. See `evidence.md`.
