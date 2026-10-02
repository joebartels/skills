# First-pass four-arm comparison

All four fresh authors used identical frozen task, eight-file inputs, architecture/style guidance and inherited settings. Only the two testing skills' availability differed. Outcomes below preserve raw reviews separately from controller reconciliation; no grades are averaged. Both writing skills remain revision 1 in this run.

| Arm | Architecture | Testing | Correctness | Confirmed result |
| --- | --- | --- | --- | --- |
| Architecture only | A | B | B | Default client follows initial 302, makes two GETs and publishes final 200; no test for the promised initial-status/count boundary. |
| Behavior only | A+ | B (reconciled; raw A+) | B (reconciled; raw A+) | Identical concrete-client probe independently reproduces the same defect/omission. Raw reviewer treated redirect policy as caller-owned; root applies the literal frozen single-request contract. |
| Isolation only | A+ | B | B | Neutral reviewer independently reproduces the same redirect/status/count cause. |
| Both | A+ | A+ | A+ | Author independently disables redirects on a client copy and tests count/configuration preservation. Identical post-review probe passes without author feedback; neutral reviewer independently confirms no actionable findings and verifies publication/lifecycle regressions. |

Every candidate reconstructs exactly and passes ordinary test/vet/format, race, ten shuffled repetitions and held contract probes on Go1.26.5 darwin/arm64 with approved loopback. Initial author/reviewer listener denials remain raw. All four candidate suites reject the six frozen compiled semantic regressions: success process exit on rejection, release before callback join, direct publication, discarded caller context, start-relative recurrence and omitted text validation. Named assertions establish detection; compiler errors/overall timeouts are not credited. These six checks did not expose the subsequently reviewed redirect gap.

The identical post-review diagnostic uses a real concrete http.Client with an observable RoundTripper, initial 302 and final 200. It observes initial rejection, request/body counts, prior bytes and borrowed configuration. Three arms fail and both passes. This is client-policy evidence, not network cancellation evidence or a pre-frozen seventh mutation. Frozen inputs/probes are unchanged.

The observed combined safeguard is useful, but a single successful arm does not establish general A-grade tests or prove causal attribution to either skill. Conditional isolation revision 2 will be evaluated separately on fresh all-case/control/repeat and affected integrated trials. No first-pass author sees this finding or receives a repair. No architectural conflict, new dependency or needless production abstraction has been observed; the both review confirms proportional structure and ownership.

Outcome reviews launch from neutral temporary paths and neutral agent names and receive only original/candidate sources, task and applicable review guidance. Earlier individual baseline/skill-on reviews had a path-label blinding limitation; that remains disclosed in the canonical record. Exact inherited model ID/reasoning, actual Go1.22 runtime, other platforms, automatic routing and broader effectiveness remain unverified.
