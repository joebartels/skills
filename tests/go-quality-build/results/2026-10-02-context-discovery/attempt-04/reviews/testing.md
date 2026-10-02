## Testing — C+
Scope: Code-area review of packet-01/candidate/process_test.go for the requested Process evolution in original/README.md; paired original tests establish retained progress coverage. Complete packet test files were inspected. Module go 1.22.
Coverage: Mapped assertions to ordered callbacks, error identity, accepted prefix, already-canceled admission, active empty success, between-job cancellation, final completion, and error/cause composition. Independently ran author tests on Go 1.26.5 and Go 1.22.12, reviewer boundary probes, and an unsafe-equality mutation in disposable copies. Also ran race/shuffle/count=3 on author tests plus the passing reviewer cases, including channel-synchronized external cancellation.
Rationale: Two independently actionable moderate regression-detection gaps select C+. F2 omits the empty/canceled interaction; F3 omits the same-concrete-non-comparable callback/cause interaction. They are contained boundary gaps, not evidence that the entire cancellation or error-preservation contract is unverified.
Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [T-G1] candidate/process_test.go:14-26 and 49-63 assert the exact observed job sequence and accepted count. They catch out-of-order work, extra callbacks after cancellation, and incorrect prefix reporting.
- [T-G2] candidate/process_test.go:74-85 independently asserts callback failure identity, cancellation classification, and the actual custom cause value through errors.Is/As. The supplied lost-custom-cause mutation fails these assertions; this is supplied mutation evidence, not an independent rerun.
- [T-G3] candidate/process_test.go:66-71 directly cancels inside the final successful callback and expects (1,nil), protecting the explicit completion rule without timing sleeps. Independently rerun author tests passed both supported and host toolchains.

Bad

- [F2][moderate][existing-in-scope] candidate/process_test.go:30-47 tests already-canceled input only with one job, and empty input only with an active context. The missing interaction allows actual F1 to pass the entire author suite. The independent TestReviewerCanceledEmpty fails for both nil and nonnil empty jobs while those author tests pass on Go 1.22.12 and Go 1.26.5. Primary remediation owner: Testing; separate from F1 because adding the missing assertion and correcting production admission are independent corrections.
- [F3][moderate][existing-in-scope] candidate/process_test.go:74-81 pairs a non-comparable sliceCause with a distinct errors.New callback failure. It never returns a non-comparable error of the same concrete type as the cause. In a disposable copy, inserting the supplied unsafe cause != callbackErr comparison still passes all author tests, but TestReviewerCauseReturnedByCallback panics with "comparing uncomparable type process.sliceCause". The unchanged candidate passes that probe. The author suite therefore misses a plausible regression against the explicit no-panicking-equality requirement. Primary remediation owner: Testing.

Suggested changes

- [F2] Add already-canceled nil and nonnil empty cases asserting zero callbacks/count, cancellation classification, and custom cause, alongside the existing active empty-success case. Confirm these cases fail before F1 is corrected.
- [F3] Add a callback that cancels with and returns the same slice-bearing concrete error, then assert cancellation classification, cause value, count, and no panic. Confirm the unsafe-equality mutation fails while the candidate passes. No comparable sentinel equality is needed for this legal error value.

Limits: Author tests pass both toolchains; they do not prove the missing interactions. Supplied held-test source is unavailable, so their passing paths were not inferred from the failing aggregate run. The supplied unsafe-equality sensitivity was independently reconstructed; lost-custom-cause sensitivity was only supplied. Reviewer race/shuffle runs exclude the known failing empty-canceled probe and passed the named exercised cases, not every context interleaving. No CI enforcement or unprovided external integration suite was assumed.

Ungraded related finding: [F1] Production currently loses cancellation classification/cause for empty already-canceled calls; see correctness-and-compatibility.md.

