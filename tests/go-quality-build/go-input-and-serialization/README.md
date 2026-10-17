# Evaluate input and serialization guidance

Requires Python 3.10+ and Go 1.22+. Use the same explicit model, reasoning
effort, tools and task budget for both conditions. Baselines receive the existing
build-skill catalog; guided trials add only `go-input-and-serialization`.

The cases cover bounded complete JSON documents, patch presence and atomicity,
exact-number relaying, permissive compatibility, a calculation that should skip
the skill, and a held-out CSV transfer. Requirements live in fixture READMEs;
expected outcomes and private probes are withheld from authors.

Prepare trials outside the checkout, with unrelated IDs and fresh author contexts:

```sh
python3 tests/go-quality-build/go-input-and-serialization/trials.py prepare \
  bounded-document baseline /tmp/input-trials/t01
python3 tests/go-quality-build/go-input-and-serialization/trials.py prepare \
  bounded-document skill-on /tmp/input-trials/t02 --skill /tmp/input-guidance
```

Give each author only `prompt.txt`, `source/` and the catalog it describes.
Keep the generated manifest, guidance/input hashes, requested model settings,
author output and checks outside the repository. Confirm matched fixture and
neighboring-catalog hashes before comparing. Author reports establish declared
selection, not automatic harness routing.

Check each completed candidate in a disposable copy:

```sh
python3 tests/go-quality-build/go-input-and-serialization/trials.py check \
  bounded-document /tmp/input-trials/t01/source
```

The checker runs candidate tests and vet, then private probes from a separate
consumer module to avoid test-helper collisions. It emits JSON and exits nonzero
on failure or missing probe execution. Inspect named assertion failures; build failures,
timeouts and missing probes are not evidence of the intended safeguard. Use
`--go /path/to/go` for each supported toolchain and `--race` when relevant.

Before drafting, run two independent baselines for each of the three main cases,
plus one compatibility and one skip control. If baselines show no meaningful
decoding defect, stop rather than invent guidance. Freeze the draft and repeat
the same cases/settings with fresh authors. Hold the transfer case out of draft
development, then run two fresh authors per condition on it.

Retain guidance only when repeated baseline defects are corrected, the other
core cases and controls preserve their contracts, and transfer reveals no new
defect or unjustified complexity. Obtain an independent arm-hidden source review;
check resource bounds, accepted/rejected inputs, exact output values and side
effects, rather than averaging grades or counting tests. Characterize probe
sensitivity with compiling counterexamples in disposable copies. Allow at most
two focused guidance revisions, rerunning affected repeats and controls; abandon
the skill if benefit remains unestablished. Report ties, regressions and tested
toolchain limits plainly. Keep comparisons and run history out of maintained docs
and commits.
