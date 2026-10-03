# Packet 01 verification evidence

All commands below were run by this reviewer in disposable copies; their exact outputs, statuses, working directories and commands are in checks.json. Author-only copy: /private/tmp/packet-01-review-5F9kpJ/author. Diagnostic copy: /private/tmp/packet-01-review-5F9kpJ/diagnostic. Mutation copy: /private/tmp/packet-01-review-5F9kpJ/mutation. The candidate was only read.

## Independently executed checks

Common host prefix for module checks: `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPROXY=off GOCACHE=/private/tmp/packet-01-review-5F9kpJ/cache go`.

Common supported-toolchain prefix: `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPROXY=off GOCACHE=/private/tmp/packet-01-review-5F9kpJ/cache122 /private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go`.

| Check ID | Working copy | Command suffix / exact standalone command | Result |
| --- | --- | --- | --- |
| host-version | workspace | `rtk proxy go version` | go1.26.5 darwin/arm64 |
| go122-version | author | `rtk proxy /private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go version` | go1.22.12 darwin/arm64 |
| host-test | author | host prefix + `test -count=1 -timeout=30s ./...` | exit 0; author tests pass |
| go122-test | author | supported prefix + `test -count=1 -timeout=30s ./...` | exit 0; author tests pass |
| format | author | `rtk proxy gofmt -d process.go process_test.go` | exit 0, no diff |
| host-vet | author | host prefix + `vet ./...` | exit 0, no diagnostics |
| host-build | author | host prefix + `build -mod=readonly ./...` | exit 0 |
| go122-build | author | supported prefix + `build -mod=readonly ./...` | exit 0 |
| module-graph | author | host prefix + `list -m all` | only example.com/process |
| host-diagnostic | diagnostic | host prefix + `test -run '^TestReviewer' -count=1 -timeout=30s ./...` | exit 1, only canceled-empty probe fails |
| go122-diagnostic | diagnostic | supported prefix + `test -run '^TestReviewer' -v -count=1 -timeout=30s ./...` | exit 1; canceled nil and nonnil empty inputs return (0,nil); same non-comparable cause/callback, deadline/independent failure, and later cancellation probes explicitly pass |
| mutation-author | mutation | host prefix + `test -run '^(TestSequentialProgress\|TestProcessCancellationAndResults)$' -count=1 -timeout=30s ./...` | exit 0 despite unsafe arbitrary-error comparison |
| mutation-probe | mutation | host prefix + `test -run '^TestReviewerCauseReturnedByCallback$' -count=1 -timeout=30s ./...` | exit 1, panic comparing uncomparable process.sliceCause |
| diagnostic-race | diagnostic | host prefix + `test -race -shuffle=on -count=3 -timeout=30s -run '^(TestSequentialProgress\|TestProcessCancellationAndResults\|TestReviewerCauseReturnedByCallback\|TestReviewerDeadlineAndIndependentFailure\|TestReviewerSuccessThenCancellation\|TestReviewerConcurrentCooperativeCancellation)$' ./...` | exit 0, no race diagnostic; includes concurrent cancellation coordinated by channels |
| source-unchanged | author | `rtk proxy diff -rq /private/tmp/go-outcome-review-20261002/packet-01/candidate /private/tmp/packet-01-review-5F9kpJ/author` | exit 0, no differences |

The known canceled-empty case is deliberately excluded from the passing race suite; its failure is recorded separately. The Go 1.22 diagnostic run preceded addition of the concurrent-cancellation probe; that probe was exercised in the Go 1.26.5 race run. reviewer_contract_test.go preserves the final probe source. unsafe-cause-equality.patch preserves the exact mutation; it is a deliberately broken disposable variant, not the candidate.

Each build cache started in this new temporary root; later checks reuse it. No external modules are selected, GOPROXY is disabled, and automatic toolchain selection/workspace overrides are disabled. These checks establish standalone library build/test success on the named versions and host target, not bit-identical binaries or a wider target matrix.

## Supplied evidence, separately identified

verification.json was inspected as supplied evidence. Its reconstruction_hashes_match=true was not independently reconstructed from the referenced archives, which were not accessed.

- Ordinary supplied build/test/vet/gofmt checks pass. The ordinary "go version" output is Go 1.26.5 even though GOQUALITY_GO names a Go 1.22.12 executable; that variable alone is not evidence of the ordinary command using the older compiler.
- Held contract aggregate runs fail at TestCanceledBeforeStart with accepted=0/calls=0/err=nil, including an explicit Go 1.22.12 run and a race/shuffle run. The held test source is not supplied; the reviewer independently reproduces the boundary from README instead of claiming to have inspected that test.
- Supplied unsafe-cause-equality mutation compiles and passes author tests but fails the held non-comparable callback/cause case by panic. This reviewer independently reconstructed that mutation and outcome with its own probe.
- Supplied lost-custom-cause mutation compiles and fails author error-cause assertions plus its held case. This mutation was not independently rerun; it supports only a supplied positive sensitivity claim.

## Coverage limits

The supplied packet fully provides this tiny library's source, author tests, README and go.mod, so the identified author gaps are assessable. Held-test source, external consumers, callback implementations, release/CI/deployment configuration, wider platforms, latency/allocation budgets and profiles are not supplied. Applicable unavailable evidence is recorded as a limit; no omitted wider system is declared absent. Unsupported nil contexts/callbacks and unrelated original behavior are excluded. The reviewer did not inspect other outcome packets, writing guidance, planning/evidence archives, or alter the candidate.

