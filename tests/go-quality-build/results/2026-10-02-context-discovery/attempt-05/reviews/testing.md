## Testing — C-
Scope: Complete author suite `candidate/stages_test.go` for the README's requested `FetchAll` budget/lifetime behavior, compared with the original basic sequential test; Go 1.22 minimum.
Coverage: Every author test and assertion was inspected; author tests, race/shuffle/repeated runs, minimum-version tests, and four isolated behavioral mutations were executed in disposable copies. Reviewer-written diagnostics independently checked budget inheritance, stage/body lifetime, closure acceptance, completed success, cancellation causes, and a local real HTTP body-read boundary.
Rationale: Two independent major gaps leave the requested total-budget continuity and body-scope lifetime contracts effectively unverified. Three narrower moderate gaps concern close-only failure, completed-success cancellation, and cancellation-transition classification. Two major issues select C-; no critical/systemic severity is claimed.
Finding counts: critical=0, major=2, moderate=3, minor=0

Good

- [G1] `candidate/stages_test.go:40`, `:116`, and `:138` assert ordered results, retained prefix/no third request, owned-body closure, and inspectable joined read/close/status failures.
- [G2] `candidate/stages_test.go:150` uses `errors.Is` to check both an independent transport error and a synchronous custom cancellation cause/classification. `:50` checks no request starts for an already-canceled caller; `:165` checks empty input performs no calls.
- [G3] Blocking transport tests wait on request context cancellation rather than guessing with sleeps. Reviewer author-suite race/shuffle/count=3, Go 1.26.5 ordinary tests, and Go 1.22.12 cgo-disabled tests passed. Supplied lost-total-parent mutation is also reported as caught; that is a narrower safeguard than all total-budget continuity regressions.

Bad

- [T1][major][existing-in-scope] `candidate/stages_test.go:86` calls the first stage without consuming meaningful budget and compares only each remaining duration to broad bounds. A mutation that creates a fresh total context for every stage passes all author tests. It fails the reviewer diagnostic comparing absolute stage deadlines. Also `:64` gives the parent 250ms and stage only 80ms, so no author assertion exercises an earlier parent deadline, despite test names mentioning it. The requested one-total-scope/earlier-parent budget contract lacks meaningful continuity/inheritance assertions.
- [T2][major][existing-in-scope] `candidate/stages_test.go:17` bodies ignore their request context, and the timeout tests block only inside `RoundTrip`. Canceling the stage immediately before `io.ReadAll` passes all author tests, while direct body-lifetime diagnostics fail in both Read and Close. The explicitly requested live stage scope through body consumption and closure, and cancellation at that body boundary, are effectively unverified.
- [T3][moderate][existing-in-scope] `candidate/stages_test.go:116` combines read and close failures. There is no successful-read/failed-close case. Changing the production failure guard to check only `readErr` passes all author tests and accepts an unsuccessfully closed body; a reviewer close-only diagnostic fails. This leaves the closure-only acceptance branch unchecked.
- [T4][moderate][existing-in-scope] The suite has no cancellation after complete reading/successful final closure. Adding a final total-context cancellation veto passes all author tests; the reviewer completed-success diagnostic fails. A narrowly timed supported success outcome can regress unnoticed.
- [T5][moderate][existing-in-scope] `candidate/stages_test.go:150` finishes cancellation before returning the transport failure, so it cannot exercise a transition between cancellation observations. Real-context and deterministic-transition diagnostics expose F1, but the author suite has no invariant that a returned custom cancellation cause also carries its classification. This is a separate regression-test correction from F1's production fix.

Suggested changes

- [T1] Assert identical absolute deadlines across total-dominated stages and exact inheritance when the caller deadline is earlier than both configured budgets. These are deterministic checks and detect restarting or widening the scope without slow sleeps.
- [T2] Add bodies that observe the request context during Read and Close, assert it remains live until closure, and check cancellation through an exercised body/transport boundary. Assert completed-stage cancel resources are released before the next request.
- [T3] Add a successful-read/close-error case that checks no body is appended and the close error remains inspectable.
- [T4] Add a successful final close that cancels the caller and verify fully completed success remains successful.
- [T5] Add a focused cancellation-transition error invariant that preserves independent failures and rejects custom cause without the standard classification. The saved deterministic diagnostic supplies one approach; retain a real-context stress check only as supplementary evidence.

Limits: `evidence.md` gives commands/outcomes and distinguishes supplied held tests from author tests. Held-test source was unavailable; passing supplied held checks does not demonstrate the author's regression coverage. Reviewer diagnostics do not count as author-suite safeguards. The first local HTTP attempt was sandbox-blocked; the authorized loopback rerun passed. Public FetchAll stress did not reproduce F1, while the helper's real-context stress and deterministic transition did. No test-count, coverage-percentage, fuzzing, benchmark, framework, or CI-policy requirements were imposed.
