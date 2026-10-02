# Value-semantics baseline outcome review

Reviewed 2026-10-01, independently of the three implementation authors and before a value skill draft exists. **All three assessed changesets earn Code Quality A / Correctness A, with no confirmed owned defect.** The current evidence does not support a claim that a new value skill would improve these outcomes. Retain the candidate's ownership/copying research, but defer standalone drafting/promotion unless a separately declared, bounded investigation establishes a real incremental target.

## Scope and provenance

The three `first` trials are `default-wire`, `snapshot-ownership`, and `required-construction`, under `/private/tmp/go-language-trials-3b23/go-values-and-zero-values/baseline/<case>/first/reconstructed/`. Reviewed original fixture code/README contracts, complete reconstructed code/tests/docs, manifests, author reports and private probes. The [fixture suite](../../tests/go-quality-build/go-values-and-zero-values/evals/evals.json) and [baseline archive](../../tests/go-quality-build/results/2026-10-01-language-contracts/go-values-and-zero-values/baseline/) identify the permanent artifacts.

All manifests record fresh `gpt-6-luna` medium authors and the same three-skill architecture catalog. All three author reports say they opened `go-api-contracts`; this is an author-reported selection fact, not independent tracing of skill reads. Baselines therefore measure behavior with existing guidance available, not a no-skill condition. This reviewer applied the unchanged Code Quality & Go Idioms and Correctness & Compatibility skills and their decision references. Private checks were visible to the reviewer; the assessment is independent of authorship, not blind to expected contracts.

Independently verified all 12 reconstructed output-file hashes and all three archived patch hashes against manifests. The reviewer did not independently reapply patches or check every input/catalog hash. The scope is introduced or worsened behavior relative to the original inputs; existing implementation details are context unless the changes expose them.

| Trial | Main decisions assessed | Code Quality & Go Idioms | Correctness & Compatibility |
| --- | --- | --- | --- |
| `default-wire/first` | Supported zero/default, explicit unlimited, nil/empty JSON, borrowed input, exact factory type | A | A |
| `snapshot-ownership/first` | Owned mutable graph fork, shared/cyclic topology, collection nilness, borrowed view, value method sets | A | A |
| `required-construction/first` | Required key validation, owned key, pointer receiver, existing mutex, concurrent observation | A | A |

The report cards below cover all three changesets. Each case independently has zero counted findings and relevant verified strengths. Passing expected behavior is sufficient for A; the number of test assertions does not establish exceptional safeguards or require A+.

## Code Quality & Go Idioms — A

Scope: Three small Go 1.22 library changesets, including new constructor/method behavior and corresponding tests/docs.

Coverage: Read all original and resulting Go files and README contracts. Assessed state representation, local value semantics, error returns, copying depth, nilness, method receivers, mutex use, documentation claims and newly exported complexity. Source uses APIs compatible with its declared Go version; actual execution used Go 1.26.5.

Rationale: Implementations directly encode the required distinctions without unnecessary public abstractions. No actionable introduced quality problem is substantiated. Optional syntax, naming, helper or documentation-volume preferences are not findings.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `default-wire/view.go:10–12,20–37` represents explicit unlimited separately from the numeric default, preserving the established zero value. Formatting only reslices the local slice descriptor before encoding; it neither normalizes nil into empty nor writes into borrowed elements.
- [G2] `snapshot-ownership/graph.go:23–48` memoizes original node pointers before following links, then separately copies maps, byte payloads and link slices. The helper stays private, and its nil checks preserve observable collection state. The comments at `graph.go:10–15` distinguish the borrowed view from the owned fork accurately.
- [G3] `required-construction/counter.go:18–24` uses a pointer receiver and the same mutex as `Sign`. It does not copy the lock or introduce an independent synchronization mechanism. Construction/key validation and ownership stay in the existing constructor.
- [G4] Added public surface is limited to the requested `NewWithLimit`, `Graph.Fork` and `(*Counter).Count`. There are no new exported state wrappers, clone interfaces, constructors for the graph, error taxonomies, packages or dependencies.

Bad

- None found.

Suggested changes

- None needed for the graded scope.

Limits: This is not a full Architecture, Testing, Security or Performance assessment. Formatter field choices, `fmt.Errorf` for a constant diagnostic, recursive traversal and retained task-oriented README wording do not establish a concrete maintenance defect here. The original missing general public-type comments are not introduced changes. No generic requirement for a useful zero counter, uniform receiver kinds or copying every borrowed view was imposed.

## Correctness & Compatibility — A

Scope: The same three changesets against the supplied public contracts, original method signatures and representative external consumers.

Coverage: Verified default/configured wire output, absent versus known-empty state, borrowed-input preservation, value-method usability on non-addressable values, graph node identity and cycles, mutable-data isolation in both directions, independently captured forks, required construction, key ownership, signature behavior, and concurrent count observation.

Rationale: All archived private probes and six additional external-package reviewer tests pass, with source inspection supporting their findings. No confirmed contract defect or meaningful unresolved contract ambiguity affects the implemented task behavior.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G5] Default and factory-created formatters retain ten items, explicit zero retains all items, positive limits truncate, and negative limits return an error. Private probes verify the exact `func() Formatter` factory type. Additional external tests check nil and non-nil empty inputs for zero/default/unlimited/positive/max-int configurations, borrowed backing-array contents including spare capacity, and stability of encoded output after later input edits.
- [G6] Graph tests preserve original node sharing and cycles while assigning distinct nodes to the fork. The reviewer additionally exercises a diamond with equal-named but distinct intermediate nodes, a shared leaf, back edges, a self-loop and a nil link. Corresponding topology survives; equal content does not cause distinct nodes to collapse.
- [G7] Map entries, byte elements, empty payloads with original spare capacity and link slots are isolated across the source and two independently captured forks. Mutations in both directions preserve the other graph's captured values. Author tests independently cover nil versus empty maps, slices and byte payloads.
- [G8] External interface assignment and calls on `map[string]Graph` index expressions compile for `Name`, `RootView` and `Fork`. `RootView` returns the exact original pointer and supports in-place edits; a zero graph retains its nil root, empty name and usable `Fork` behavior.
- [G9] External constructor checks reject 0- and 15-byte keys and accept a 16-byte key. Mutating the caller's key after construction does not change HMAC output; mutating a returned signature does not affect later signatures. The new `Count` remains correct while signing: the reviewer executes 1,024 calls across eight goroutines, checks monotonic bounded observations, waits for completion and checks the exact final count under the race detector.

Bad

- None found.

Suggested changes

- None needed for the graded scope.

Limits: Actual Go 1.22, other platforms and real downstream repositories were not executed. The graph explicitly has no concurrent mutation guarantee; no such guarantee was tested or invented. Race-enabled counter tests exercise relevant interleavings but are not exhaustive. Very deep/adversarial graph workloads, heap profiles, counter overflow and copying a counter after first use were not tested; no supported workload or task change makes these a new defect. Copying after use is explicitly outside the supplied counter contract. The cryptographic check verifies preserved deterministic behavior and key ownership, not a cryptographic security audit. Tests were not mutation-tested or assigned a Testing grade.

## Ambiguities and non-findings

No confirmed owned defect was found. The graph contract explicitly preserves sharing of **nodes through links**. It does not require preservation of incidental backing-array/map aliasing within one source graph, slice capacity, or pointer addresses across capture. The implementation preserves node topology and independent editing of all required mutable fields; demanding additional intra-fork payload aliasing would add a new contract. Similarly, nil versus non-nil empty state is required and tested, whereas retaining spare capacity is not.

The unsupported zero counter remains unsupported. `Count` repeats the existing constructor-precondition panic while holding the same lock and releasing it through `defer`; a forced zero-value redesign would contradict the task. Conversely, the lightweight graph must remain a usable value, and the implementation correctly retains value receivers. These contextual choices are positive counterexamples to blanket rules about constructors, pointer receivers or copying.

## Independent verification

Copied the three reconstructed modules into `/private/tmp/value-outcome-review-3b23/<case>/`, added the repository private probe as `language_contract_probe_test.go`, and added two external-package reviewer tests per case in `reviewer_test.go`. Trial outputs, fixture files and runtime skills were not changed.

Set `GOCACHE=/private/tmp/value-outcome-review-3b23/gocache` and `GOWORK=off`. In each case ran:

```sh
rtk proxy go test -race -count=1 -timeout=45s -v ./...
rtk proxy go vet ./...
rtk proxy gofmt -l <all author Go files>
```

All three test commands and all three vet commands exited 0. All three format checks exited 0 with empty output. `rtk proxy go version` reported `go1.26.5 darwin/arm64`.

Scratch evidence retained for archival: `identity-checks.json`, three `<case>-verification.json` files, and each case's `reviewer_test.go` beneath `/private/tmp/value-outcome-review-3b23/`. Those JSON files preserve the exact command arguments and stdout/stderr. The original private probes and baseline verification are already in the repository; the controller may archive the additional reviewer artifacts without altering author results.

## Distinct content and effectiveness opinion

A value skill could have a coherent distinct center: tracing ownership through reference-containing values; choosing the required copy depth; preserving shared graph topology; separating view APIs from capture APIs; protecting supported nil/empty and explicit-default semantics; and deciding receivers from mutation, method sets and copy-sensitive state. The [source audit](language-contracts-source-audit.md) remains useful research for those decisions. Nil/empty syntax preferences or universal constructor/copy rules would weaken that content.

The current outcomes already implement those distinctions correctly, including the most demanding graph ownership case. Default/wire compatibility also substantially overlaps the existing [API-contract guidance](../../plugins/go-quality-build/skills/go-api-contracts/SKILL.md), which all authors report opening. No unnecessary exported complexity or copying cost attributable to a bad decision was found. The graph's deep copies fulfill the explicit owned-fork contract; the formatter correctly avoids copying borrowed data it only reads. Author effort, tokens and comparative performance were not independently measured, so no efficiency improvement is claimed.

Under the [recorded promotion gate](README.md#evaluation-and-promotion), useful potential content is insufficient without evidence of better outcomes. **Defer standalone value-skill drafting/promotion on these baselines.** This is a limit of the available improvement evidence, not proof that value-semantics guidance is useless generally. If further investigation is chosen, declare its task/search budget before results and preserve these successful baselines. Keep the reserved transfer task reserved; do not reinterpret correct aliasing or construction decisions as failures to manufacture a target.

Next action for the controller: record the independent A/A findings and the draft/defer decision, preserve the reviewer evidence, then follow the approved sequence for the conditional names/docs assessment. This reviewer changed only this assessment file and leaves the canonical progress record to the controller.
