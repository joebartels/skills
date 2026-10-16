# Evaluate client-call guidance

Requires Python 3.11+, Git, and Go 1.22+. Run from the repository root.
The six cases cover HTTP replay, retry hints, gRPC native policy, breaker health,
a calculation that should skip the skill, and gRPC transfer.

## Prepare comparable author trials

Use `client_call_trials.py prepare` with the same case and separate trial IDs
for `baseline` and `skill-on`. It copies the fixture and neighboring guidance;
only `skill-on` includes `go-client-calls`. Keep generated work outside the repo:

```sh
python3 tests/go-quality-build/client_call_trials.py prepare \
  --case http-write-replay --arm baseline --trial-id t01 \
  --scratch /tmp/client-call-trials --archive /tmp/client-call-results
```

Give the author only that trial's `prompt.txt`, `catalog/`, and `source/`.
Use fresh contexts with the same explicit model, reasoning effort, tool access,
and time budget. Record those settings and the tested guidance revision beside
temporary results; the helper cannot discover model settings. Withhold
`expected_output`, probes, other trials and reviewer conclusions from authors.

## Check completed candidates

The checker copies the candidate, adds all applicable private probes and runs
its Go tests without modifying the original. Nonzero exit means a failed check:

```sh
python3 tests/go-quality-build/go-client-calls/check.py http-write-replay \
  /tmp/client-call-trials/t01/source
```

Repeat for both arms on the minimum and current supported toolchains, using
`--go /path/to/go` to select a toolchain and `--race` for concurrent cases.
Dependencies stay pinned by each fixture's `go.mod` and `go.sum`.

Assess preserved contracts, effects, actual request/RPC counts, caller-visible
errors, resource ownership, and unnecessary complexity. Compare the same checks
on both arms; a passing control or tied result is not evidence of improvement.
Fresh tasks test transfer. Model outcomes can vary even with matched settings.

The standalone check characterizes resolver-config behavior used by the guidance:

```sh
cd tests/go-quality-build/go-client-calls/checks/grpc-service-config
go test -mod=readonly -count=1 -race ./...
```

Repository validation checks packaging and fixture structure, not model behavior.
Keep generated trials, logs and comparison results outside the checkout.
