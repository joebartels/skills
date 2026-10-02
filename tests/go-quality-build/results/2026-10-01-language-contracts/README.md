# Language-contract trials, 2026-10-01

The three architecture skills are available identically in both arms. Authors see
only task inputs and catalog descriptions, choose relevant skills explicitly,
and record opened files. This measures explicit exposure, not automatic routing.
Each manifest records input/catalog/output SHA-256 values; `source.patch` applies
to the tracked fixture to reconstruct the completed source exactly. Two independently reviewed candidate drafts now exist outside runtime. One reader
pair shows an observed improvement under the later blind review; repeated causal
benefit and runtime promotion remain unproven.

## Errors: reviewed draft, promotion unproven

The independent [outcome review](../../../../docs/go-quality-build/error-contracts-baseline-review.md)
initially found no confirmed owned defect in four gpt-6-luna medium outcomes (A/A). The
single bounded supplementary cursor task also passes. The reserved CSV case
was consumed prospectively after drafting as a matched compatibility control.
It is not baseline failure search. Successful
baselines outside the reader co-failure finding show no observed incremental benefit.
The later arm-blinded review found a caller-cause loss in the reader baseline,
which the draft-exposure implementation avoids; see the reconciliation below.

Two early inherited-model trials are supplemental: their exact model IDs are
unavailable. The explicit gpt-6-luna medium cohort was declared before examining
completed outputs; inherited trials are not silently treated as matched samples.
The pre-draft private-calculation control opened no skills. Final-byte draft
exposure trials are complete, with one fresh author per case/arm.

| Trial | Author model / effort | Archived verification |
| --- | --- | --- |
| backend-errors/first | gpt-6-luna / medium | verification-final.json: PASS |
| cli-completion/first | inherited root; exact id not surfaced / inherited root, no override | verification-corrected.json: PASS; verification.json: INVALID/FAIL; see correction |
| cli-completion/named | gpt-6-luna / medium | verification-final.json: PASS |
| cursor-iteration/named | gpt-6-luna / medium | verification-final.json: PASS |
| not-error-work/named | gpt-6-luna / medium | verification-final.json: PASS |
| reader-errors/first | inherited root; exact id not surfaced / inherited root, no override | verification.json: PASS |
| reader-errors/named | gpt-6-luna / medium | verification.json: PASS |

The first inherited CLI `verification.json` is invalid: its writer double exposed
a promoted `WriteString` that bypassed failure injection, and default cache access
was blocked. `verification-corrected.json` preserves the rerun with the corrected
double and writable cache. Neither failure is scored as an author defect. Other
`verification.json` files predate the cache fix but contain successful checks;
main new records use `verification-final.json`. Reader `probe-validation.json`
records a deliberately broken private copy losing a read cause; it is probe
validation, not a model failure.

Independent reviewer raw checks and extra probes are under
[expert-review](go-error-contracts/expert-review/), with SHA-256 identities.
The review is independent of authoring, but sees expected contracts/private checks.
It is not blinded to arm labels. Real Go 1.22 execution, other platforms and broad
usefulness remain unverified; checks used Go 1.26.5 darwin/arm64.

## Values

Five original cases are fixed before drafting, including a reserved borrowed-frame
transfer. Four fresh gpt-6-luna medium baselines are archived (three applicable and one private control). All pass the private checks; the control opens no skills. Independent [value outcome review](../../../../docs/go-quality-build/value-semantics-baseline-review.md) grades all applicable changes A/A with no confirmed owned defect. The user's content condition authorizes the reviewed value draft; promotion remains unproven; this is not evidence that value guidance is useless on other tasks. Reserved transfer is now a prospectively consumed matched control after drafting. Additional reviewer checks/probes and identities are under [expert-review](go-values-and-zero-values/expert-review/). No measured added benefit is claimed.

## Draft and reference review

The [exact-byte expert review](../../../../docs/go-quality-build/language-contracts-final-review.md#2026-10-01-follow-up-exact-draft-content-review)
approves the error/value drafts and names/docs reference for candidate evaluation.
Drafts remain outside runtime; content sufficiency is separate from measured
incremental effectiveness. Final hashes and reading cost are recorded. Raw
compiled-example/semantic checks and source are in [draft-review](draft-review/).
Names/docs has no standalone skill or behavior suite.

## Combined use

A fresh library/CLI case tests partial caller input errors, nested independent
snapshots, intentional nil/empty results and actual process status/prefix output.
The fixed three-skill catalog is identical; the draft-exposure arm adds both
reviewed candidates. This checks interaction and possible harm; it does not
establish automatic routing or a broad effectiveness estimate. Both arms pass the current library/private checks and five executable CLI scenarios. The draft arm reports opening API contracts plus both candidates; the baseline opens API contracts. No interaction failure is observed in these probes. The draft author also read the canonical design record, potentially priming selection/expectations. Its manifest records this deviation; the pair is excluded from clean context-withheld evidence. Expert code grades remain A/A.

## Final-byte exposure archive

All 27 authored trials are archived and reconstructable: 14 existing-catalog
baselines and 13 draft-exposure runs. These cover 12 paired tasks, including 11
clean matched pairs and one context-primed combined pair; the
extra first value control is non-blind supplemental evidence. Two inherited-model
error baselines remain supplemental because their exact model IDs are unavailable.
Every clean pair uses gpt-6-luna medium and identical original inputs/existing
catalog bytes. There are 105 successful current recorded check commands; the
invalid first CLI verification remains excluded beside its corrected result.

| Draft exposure | Case/repeat | Output files | Current checks | Provenance |
| --- | --- | ---: | --- | --- |
| combined | language-contracts/first | 6 | PASS | Explicit exposure; canonical-record read |
| go-error-contracts | backend-errors/first | 4 | PASS | Fresh explicit exposure |
| go-error-contracts | cli-completion/first | 6 | PASS | Fresh explicit exposure |
| go-error-contracts | csv-transfer/first | 4 | PASS | Fresh explicit exposure |
| go-error-contracts | cursor-iteration/first | 5 | PASS | Fresh explicit exposure |
| go-error-contracts | not-error-work/first | 4 | PASS | Fresh explicit exposure |
| go-error-contracts | reader-errors/first | 5 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | borrowed-transfer/first | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | default-wire/first | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | not-value-work/first | 4 | PASS | Non-blind supplemental control |
| go-values-and-zero-values | not-value-work/isolated | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | required-construction/first | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | snapshot-ownership/first | 5 | PASS | Fresh explicit exposure |

The isolated error and value controls report no opened skills. Applicable authors
report opening their candidate; the combined author opens both. Selection remains
self-reported explicit exposure, not automatic routing. Frozen reviewed hashes
match every catalog. One prepared-only error calculation directory was never
dispatched and is not counted as an author attempt; [inventory.json](inventory.json)
records this distinction. Exact task/catalog prompts and author reports are
preserved; outer dispatch messages/environment history are not full transcripts.

A fresh reviewer assessed all 12 paired candidate trees with arm identities withheld
until its initial report was frozen. Original inputs/task contracts and private probes
were visible, so this is arm blinding rather than expectation blinding.

## Frozen blind findings and reconciliation

The [frozen blind report](blind-review/frozen-report.md) grades the reader baseline
B/B for losing a caller read cause when the same bytes are malformed, and the
draft-exposure reader A/A. The original also exposes that cause. All other reviewed
candidates earn A/A; no harmful draft behavior is confirmed. Initial baseline
review excluded this combination as a precedence ambiguity. A separate independent
[contract adjudication](../../../../docs/go-quality-build/reader-contract-adjudication.md)
supports the supplied inspection promise; the
[effectiveness review](../../../../docs/go-quality-build/language-contracts-effectiveness-review.md)
reconciles judgments and provenance without changing either raw report or task. This finding was
identified after drafting, in one pair; it is not a pre-draft failure, a repeated
removal result, or a broad causal-benefit estimate.

[Reviewer checks and reproducer bytes](blind-review/artifact-hashes.json), the
post-freeze [arm map](blind-review/arm-map.json), and root's independent
[original/baseline/skill-on observation](blind-review/root-cofailure-observations.json)
are archived. Additional reviewer checks pass except the baseline's co-failure
assertion. Runtime promotion remains unproven.

## Delivery verification

[Final artifact verification](final-artifact-verification.json) records the original
local audit: 27 reconstructed patches and 113 input, 122 output, 129 catalog,
181 blind-source and 23 upstream identities. Its
[captured script](final-artifact-verification.py.txt) retains the exact author-specific
checkout and temporary paths used then; it is historical evidence and cannot be
run unchanged from another checkout. It also writes the historical result file,
so do not use it to replay or overwrite archived results.
[Final structural checks](final-structural-checks.json)
pass root/build/layout/grade tests, repository validation, both draft validators
and Claude plugin/marketplace validation. The independent recommendation is two
reviewed drafts plus the names/docs reference, with runtime promotion deferred.

## Reconstructing archived sources from another checkout

The fixtures, patches and identity manifests are sufficient to reconstruct the
completed code. This does not reproduce model authoring, full tool transcripts,
past review judgments or the original machine's environment. The historical audit
capture above is not a portable verifier or a fresh-checkout test command.

From the repository root, with Python 3, Git and RTK installed, this read-only
check reconstructs all trials in an automatically cleaned temporary directory:

```sh
rtk proxy python3 -B - <<'PY'
from pathlib import Path
import hashlib, json, shutil, subprocess, tempfile
root = Path.cwd()
archive = root / 'tests/go-quality-build/results/2026-10-01-language-contracts'
def hashes(folder):
    return {str(p.relative_to(folder)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in folder.rglob('*') if p.is_file() and '.git' not in p.parts}
manifests = sorted(archive.glob('*/*/*/*/manifest.json'))
if len(manifests) != 27:
    raise SystemExit('expected 27 trial manifests; run from the repository root')
with tempfile.TemporaryDirectory(prefix='go-language-replay-') as tmp:
    for n, path in enumerate(manifests):
        m = json.loads(path.read_text())
        original = root / f'tests/go-quality-build/{m["skill"]}/evals/files/{m["case"]}'
        patch = path.with_name('source.patch')
        if hashes(original) != m['input_hashes'] or hashlib.sha256(patch.read_bytes()).hexdigest() != m['patch_sha256']:
            raise SystemExit(f'input or patch mismatch: {path}')
        copy = Path(tmp) / str(n)
        shutil.copytree(original, copy)
        subprocess.run(['rtk', 'proxy', 'git', 'apply', str(patch)], cwd=copy, check=True)
        if hashes(copy) != m['output_hashes']:
            raise SystemExit(f'output mismatch: {path}')
print('PASS: 27 archived source reconstructions')
PY
```

For the wider identity audit, derive catalog files from each manifest's skill name
and compare its recorded SHA-256 values with the runtime architecture skills or
candidate draft directories. Rebuild blind-review inputs using these rules, then
compare them with `blind-review/blind-source-manifest.json`:

| Blind entry | Archived source |
| --- | --- |
| `original/<file>` | The paired skill/case's `evals/files/<case>/<file>` |
| `candidate-a/<file>` or `candidate-b/<file>` | The reconstructed arm/repeat selected by `blind-review/arm-map.json` |
| `task.txt` | The case's `evals.json` prompt, followed by one newline |
| `probe_test.go.txt` | The paired `probes/<case>_test.go.txt` bytes |

The upstream corpus is not included in this archive. Obtain the exact audited
revision explicitly in a new temporary Git repository; fetching it requires
network access. The directory below must be unused:

```sh
rtk proxy git init --quiet /tmp/go-language-upstream-replay
rtk proxy git -C /tmp/go-language-upstream-replay fetch --depth=1 https://github.com/samber/cc-skills-golang.git 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
```

Read each path in `docs/go-quality-build/language-contracts-upstream-manifest.json`
with `git show <revision>:<path>` from that repository. Compare the raw bytes'
SHA-256, Git blob SHA-1, byte count and line count with the manifest; preserve final
newlines. No original `/private/tmp` tree is required for these steps.

The separate [fresh-checkout audit](copilot-review/replay-verification.json)
verified all 27 trials, 181 rebuilt blind identities and 23 newly fetched upstream
files at PR head `97bc0a9`, without reading the original author checkout or scratch
trees. It also confirms the captured original script/result remain unchanged.
