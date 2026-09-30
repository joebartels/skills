## Correctness & Compatibility — C
Scope: Review umbrella-live; complete existing code area for catalog/catalog.go, catalog/catalog_test.go, go.mod and REVIEW.md; export/export.go inspected only to trace the Snapshot consumer. Paths relative to golang/go-quality-report/evals/files/catalog-export. Repository HEAD 28c092c44f4a0709c4132dac3b3947dabfae5a84; untracked fixture identities match all six SHA256 values in /tmp/go-umbrella-eval/live/manifest.json. Sequential library; go.mod declares Go 1.21.
Coverage: Constructor copies; empty/nil/zero-value catalogs; append and element-write behavior; multiple snapshots and export consumer; effective declared language and lack of build tags; concurrency excluded by explicit sequential contract. No historical upgrade claim.
Rationale: C: one contained major breach of the central Snapshot contract on ordinary nonempty input; direct mutation reproduces the state change.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [G1] catalog/catalog.go:6 copies constructor input; TestReviewConstructorIsolation passes, verifying callers cannot mutate catalog state through their original slice.

Bad

- [F1][major][existing-in-scope] catalog/catalog.go:9 returns a capacity-limited view of the owned backing array. Writing first[0] changes the catalog and a previously obtained second Snapshot; reproduced in catalog-repro.log. export/export.go:10-12 assigns uppercase strings through this same view. This breaks the explicitly promised independent mutability/ownership contract on ordinary nonempty input. It is one contained major issue; no severe broad or irreversible reach is established. Primary remediation owner: catalog Snapshot implementation. Same root cause in both Architecture and Correctness cards; do not add their counts.

Suggested changes

- [F1] Return a copy of the stored string slice (the existing append-copy idiom is compatible with the declared language) so element writes cannot mutate catalog-owned storage. Add mutation-isolation assertions for two snapshots and verify the export caller preserves labels. The scratch repro supplies the direct regression assertion; proposed fix and regression additions have not been applied.

Limits: Executed in disposable /tmp/go-umbrella-eval/live/catalog-check with GOWORK=off, GOTOOLCHAIN=local, GOCACHE=/tmp/go-umbrella-eval/live/go-cache, GOPROXY=off, GOSUMDB=off: `rtk proxy go env GOVERSION GOOS GOARCH GOMOD GOWORK` reports go1.26.5 darwin/arm64, local module, workspace off; `rtk proxy go list -m all` reports only example.com/catalog-export; `rtk proxy go build -mod=readonly ./...` passes; `rtk proxy go test -mod=readonly -count=1 ./catalog` passes. Exact outputs: /tmp/go-umbrella-eval/live/catalog-checks.log. Added scratch-only review_test.go and executed `rtk proxy go test -mod=readonly -count=1 -v ./catalog -run TestReview`: constructor isolation and nil/empty/zero pass; Snapshot independence fails with stored label and second snapshot both changed. Exact output: /tmp/go-umbrella-eval/live/catalog-repro.log. No fixture edits. No Go 1.21 executable, race scan, benchmarks, profile or vulnerability scan run; no concurrent use promised.

Invocation provenance: Read and applied golang/go-correctness-and-compatibility/SKILL.md and golang/go-correctness-and-compatibility/references/correctness-compatibility-decisions.md; also reporting.md and grading.md from go-quality-report/references. No expectations or evaluation results read.
