# Frozen integrated mutation meanings

Candidate tests only; probes are never injected into mutation runs. Patch each candidate semantically at its own implementation, verify clean and compiled states, then interpret explicit contract failures. A compiler error, unrelated panic or overall timeout is not a detected regression.

- process-success-on-error: main retains diagnostic but exits 0 after apply/watch failure.
- host-release-before-join: cancellation returns/releases while an already-started refresh is still doing cooperative cleanup.
- direct-snapshot-truncation: replace atomic snapshot publication with in-place destination writes; an already-open old reader sees modified bytes.
- discarded-caller-context: refresh receives background context instead of the caller context.
- start-relative-recurrence: delay is measured from callback start so a long successful cycle has no required post-completion interval.
- omitted-text-validation: a nonempty key with blank text is accepted; useful separate field/state signal.

Targeted patches are controller diagnostics, not author input or a prescribed test form. POSIX snapshot observation is task-scoped; Windows runtime is unverified.
