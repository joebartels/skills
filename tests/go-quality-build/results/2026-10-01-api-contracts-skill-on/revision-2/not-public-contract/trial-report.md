# Revised skill applicability trial

## Selection

- Fixture: `not-public-contract`, a private Go cache expiry helper.
- Provenance: compared each of the four fixture files (`README.md`, `go.mod`, `internal/cache/expiry.go`, and `internal/cache/expiry_test.go`) byte-for-byte against `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-api-contracts/evals/files/not-public-contract`; all four matched.
- Applicability decision: the supplied skill does not apply. The task changes a private package helper and establishes no external API or supported CLI, wire, file, or error contract; the skill description explicitly says to skip private helper fixes without an established external contract.
- Skill file opened: no. Applicability was decided from the supplied description before opening the skill.

## Changes made in the fixture

- `internal/cache/expiry.go`: changed the deadline comparison from `now > deadline` to `now >= deadline`; retained the zero-deadline never-expires guard.
- `internal/cache/expiry_test.go`: added an equality-boundary regression assertion while retaining the past-deadline, future-deadline, and zero-deadline cases.

## Checks and results

- After adding the regression assertion and before changing the helper, `rtk go test ./...` failed at `entry at deadline is live`, confirming the test detects the bug.
- After the fix, `rtk proxy env GOCACHE=/private/tmp/go-quality-build-api-eval/go-build-cache go test ./...` passed (`example.com/cacheworker/internal/cache`).
- `rtk proxy env GOCACHE=/private/tmp/go-quality-build-api-eval/go-build-cache go vet ./...` completed successfully (no diagnostics).
- The temporary `GOCACHE` was needed because the default Go build-cache location was not usable in this sandbox.

## Limits

This trial covers only the supplied four-file private-helper fixture and focused package checks. It does not establish or evaluate a public compatibility contract, release migration, integration behavior, or the skill's performance on other fixture types. The repository worktree was not edited.
