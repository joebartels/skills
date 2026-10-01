# Trial report: API contract skill applicability

## Skill selection

Selected `go-api-contracts`. The requested change is to cache expiry semantics: entries become unavailable at the deadline instead of one clock tick later. Although the helper must remain private, callers can observe which cached entries are still returned at that boundary, so the requested behavior changes a consumer-visible contract. This is not a private implementation-only edit without observable behavior change.

## Fixture provenance

Before editing, compared every file recursively and byte-for-byte against `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-api-contracts/evals/files/not-public-contract`. The disposable copy contains the same four paths and all bytes match:

- `README.md` — SHA-256 `db815043de5c2878343167e62d02a0b0605253e9f4f2f4965577d83b846b5c1c`
- `go.mod` — SHA-256 `b358e19a75e6027b4076c0c3c750f180164c4236910c2527d4de8afdb6f03074`
- `internal/cache/expiry.go` — SHA-256 `5c57765c332bf283ab2d829f4338e9655afceb54b2f74b2dec806fc7adee75c3`
- `internal/cache/expiry_test.go` — SHA-256 `86743bb0b723c57621a7e555f72b096c1a153c964528d2c7a5d761da17b2cc98`

## Selection rationale and limits

The selection was based on the requested semantic boundary change: callers can observe whether an entry remains live exactly at its deadline. The fixture confirms the helper stays private and has no persisted or serialized representation. This trial has no downstream consumer repository or separately documented cache policy, so compatibility beyond the local call boundary was not evaluated.

## Files changed

- `internal/cache/expiry.go` — changed the deadline comparison from `>` to `>=` and clarified the comment.
- `internal/cache/expiry_test.go` — added the equality-at-deadline regression assertion; retained coverage for future, past, and zero deadlines.

## Checks and results

- `go test ./...` initially failed before compilation because the default Go build cache under the user's Library directory was not writable in this sandbox (`operation not permitted`).
- `GOCACHE=/private/tmp/go-quality-build-api-eval/skill-on-luna/not-public-contract/.gocache go test ./...` — passed (`example.com/cacheworker/internal/cache`).
- `GOCACHE=/private/tmp/go-quality-build-api-eval/skill-on-luna/not-public-contract/.gocache go vet ./...` — passed with no diagnostics.

## Limits

Verification covers the supplied private package only. The fixture README states Go 1.22 support; checks ran with Go 1.26.5 on darwin/arm64. No external module consumer or production worker integration was available in this fixture.
