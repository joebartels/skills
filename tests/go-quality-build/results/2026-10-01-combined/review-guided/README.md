# Combined review-guided repair — first repair reviewed

This separate repair follows the [first-pass reviews](../trial-summary.md). The author saw those findings; it is not a fresh blind skill-on result and cannot establish blind skill uplift. All first-pass archive files remain unchanged.

The unedited [repair report](review-guided-report.md), source-only patch against the original fixture, complete source hashes and fresh [verification](verification.json) are preserved. The patch reconstructs the entire repaired candidate byte-for-byte, including the new server test file. Author reports and generated caches are excluded from source patches. [manifest.json](manifest.json) records identity and limits.

Fresh ordinary tests, uncached verbose tests, uncached race tests and vet passed with a writable GOCACHE. Exact command arrays, stdout/stderr and exit codes are saved. Archive checks did not rerun author-reported mutations or executable bind-failure reproduction. The report discloses that live-listener shutdown was blocked by socket permissions, and its large-file concurrent-publication observation does not inject filesystem failure. Independent reviews below assess the actual coverage and remaining risk.

Independent repair review reports Architecture A, Correctness A and Testing B; a complete combined-gate claim is not made. Go checks used 1.26.5 darwin/arm64; actual Go 1.22 and other platforms remain unverified. No skill, canonical docs or first-pass evidence was changed, and no commit was made.

## Independent repair review

Unedited reports: [Architecture/Correctness](review-repair-arch-correct.md) and [Testing](review-repair-testing.md). Both were copied byte-for-byte and are hashed in manifest.json. These reviews saw first-pass findings and candidate reports; they are independent repair verification, not blind skill-uplift evidence.

Architecture is **A** and Correctness is **A**. The server now returns startup failure only after cleanup and joining, and an independent executable bind-denial check exited 1. Controlled lifecycle tests verify the runner's ordering; successful live-listener signal shutdown remains unverified.

Testing is **B** for remaining moderate publication-test sensitivity finding T4. Independent rechecks make the no-join, CLI exit-zero and removed-text-validation mutations fail in all ten repetitions each. The direct-write mutation still passes intermittently under default scheduling (2/10 initially, 3/20 in a separate recorded run) and passes **20/20 with GOMAXPROCS=1**, reproduced in two runs. The observer is not synchronized with the publication window; a complete truncating write can occur between its reads.

The author report's claim that all four mutations failed is retained unedited as historical evidence, but the independent recheck limits that claim: three have reliable demonstrated signal; publication does not. Current production still uses atomic replacement, so the finding concerns regression detection rather than a demonstrated production direct-write defect. No filesystem failure was injected and cross-platform behavior remains unverified.

A second test-only repair is underway separately. This directory freezes the first repair and its A/A/B assessment; it does not contain or grade the subsequent repair.

Review SHA-256 `review-repair-arch-correct.md`: `0e5c6af5846ed4355b15aae3803601807257826a08b138ef1add05e569fe3167`.

Review SHA-256 `review-repair-testing.md`: `0e8f1e103ca13fab7c43dd0bd5fff7d0cb0af09f4806e4b5d93db914b9063709`.
