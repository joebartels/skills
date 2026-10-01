# Candidate C review addendum — newly surfaced CLI compatibility evidence

This follow-up was prompted by newly surfaced evidence: an existing valid file named `-jobs.json` may no longer be accepted by the same 1.x invocation. The original report remains unchanged. I inspected neither the skill-on candidate nor the build skill.

## Reproduction

Copied original `wire-contract` and baseline `candidate-c` into `/private/tmp/cli-dashfile-review-7edahr4f/original` and `/private/tmp/cli-dashfile-review-7edahr4f/baseline`. Each directory contains a real file named `-jobs.json` with exactly `[]`. Built each copy using the following command, with its directory as cwd:

```sh
GOCACHE=/private/tmp/cli-dashfile-review-7edahr4f/gocache go build -o joblist .
```

Both builds exited 0 with empty stdout and stderr. The identical invocation in each directory was:

```sh
./joblist -jobs.json
```

Original: exit 0, stdout `{"jobs":null}\n`, stderr empty.

Baseline candidate C: exit 2, stdout empty, stderr exactly:

```text
flag provided but not defined: -jobs.json
Usage of joblist:
  -state value
    	include only records with this exact state

panic calling String method on zero main.stateFlag for flag state: runtime error: invalid memory address or nil pointer dereference
```

The indentation before `include` is four spaces followed by a tab. The final line is diagnostic text from flag usage formatting, not an uncaught process panic: the process returned its normal error exit 2.

Additional comparisons:

| Invocation | Original exit / stdout | Baseline exit / stdout |
| --- | --- | --- |
| `./joblist ./-jobs.json` | 0 / `{"jobs":null}\n` | 0 / `{"jobs":null}\n` |
| `./joblist -- -jobs.json` | 2 / empty | 0 / `{"jobs":null}\n` |

Successful invocations had empty stderr. Original `-- -jobs.json` emitted `usage: joblist FILE\n`. Escaping or rewriting the filename is a workaround for changed callers; it does not preserve the existing invocation.

## Architecture & Design — A (unchanged)
Scope: The same original-to-baseline CLI changeset reviewed in `review-cli.md`, with this additional argument-compatibility evidence.
Coverage: Reassessed parser boundary, optional state representation, and impact of the reproduction on the previous architectural assessment.
Rationale: The confirmed issue is a concrete accepted-input compatibility regression, owned by Correctness & Compatibility below. Repair can remain local to argument parsing; the private model/wire separation and explicit filter-presence design remain appropriate. No additional architectural defect is substantiated.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate-c/main.go:20` and `:56` distinguish filter presence from its value, preserving the required empty-string distinction.
- [C-G2] `candidate-c/main.go:12` preserves the public `id` JSON name independently of the internal `Key` field.

Bad

- None found within this topic. Cross-reference [C-F1] below for the confirmed behavioral regression.

Suggested changes

- None needed independently of [C-F1].

Limits: This targeted follow-up adds no broader architecture coverage beyond the original report.

## Correctness & Compatibility — B (revised from A)
Scope: Same supplied joblist 1.x diff; documented one-file invocation and its supported upgrade path.
Coverage: Fresh builds and identical real-file invocation in original and baseline; two alternate argument spellings. Prior independent JSON/filter/error checks remain valid for their exercised inputs.
Rationale: One moderate introduced compatibility issue. A valid filename accepted by the documented original one-file interface now produces a usage failure under the same invocation. This is a contained accepted-input regression requiring caller changes, with simple filename-escaping workarounds; there is no data loss or broad failure justifying major severity. One moderate issue selects B. The original report's unspecified-conventions limit did not establish preservation of this existing supported case, so its A conclusion was too strong.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [C-G3] Prior fresh checks established exact and explicit-empty state filtering and retained JSON schema/order/empty-result behavior for ordinary file arguments.

Bad

- [C-F1][moderate][introduced] `candidate-c/main.go:25` passes every argument to `flag.FlagSet.Parse`, reinterpreting a formerly valid sole filename beginning with `-` as a flag. Original `main.go` checked only argument count and directly read `args[0]`; the unchanged README documents one JSON file argument without a leading-dash exclusion. Verified `./joblist -jobs.json` changes from exit 0 with `{"jobs":null}` to exit 2 and no stdout. Existing scripts using that valid spelling fail after the 1.x upgrade. Primary remediation owner: Correctness & Compatibility.

Suggested changes

- [C-F1] Preserve the existing sole-file argument behavior when introducing state-option parsing; explicitly define and handle the ambiguity between filenames and recognized options. Add a regression that creates `-jobs.json` containing `[]` and exercises the identical old invocation. Verify ordinary filtering, explicit empty filtering, and malformed new-option handling still meet the intended CLI contract. A documentation-only escape recommendation does not restore existing invocations.

Limits: Checks used Go 1.26.5 darwin/arm64 and disposable copies; original source and baseline candidate were not modified. This follow-up covers the specifically surfaced filename case, not an exhaustive CLI grammar audit. Candidate D grades are unchanged. The extra panic wording during usage is observed but not counted as a separate compatibility finding: diagnostic wording is explicitly unstable, and the supported error exit/stdout behavior is preserved there.

Original report SHA-256 before and after this investigation: `f8395017e0d8da0f053f326e4f435ee24e0a381d0ac7dec993f6f1cb7424d0f9`.
