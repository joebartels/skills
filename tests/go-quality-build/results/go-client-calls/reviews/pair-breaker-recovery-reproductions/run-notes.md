# Neutral pair reproduction record

Inputs: `/private/tmp/go-client-campaign/reviews/pairs/breaker-recovery`. All entries in `source-hashes.json` matched before running checks. Supplied source and probes were not edited.

Disposable copies: location recorded in `scratch-path.txt`. Each copy contains its supplied source, `probes/frozen_test.go.txt` as `frozen_review_test.go`, `probes/supplemental_test.go.txt` as `supplemental_review_test.go`, and the identical adjacent `pair_review_test.go`.

Each command ran in the appropriate scratch A or B directory. Common prefix:

```sh
rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off
```

Append these exact command arguments to that prefix:

| Check | Arguments | A exit | B exit |
| --- | --- | --- | --- |
| Host | `go test -ldflags=-linkmode=external -count=1 -timeout=20s -v ./...` | 1 | 0 |
| Race | `go test -race -ldflags=-linkmode=external -count=1 -timeout=20s ./...` | 1 | 0 |
| Minimum | `/private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -count=1 -timeout=20s ./...` | 1 | 0 |
| Vet | `go vet ./...` | 0 | 0 |

Host version: `go version go1.26.5 darwin/arm64`. Minimum: `go version go1.22.12 darwin/arm64`. Race runs reported assertion failures for A but no data-race diagnostics for either label. Local ephemeral HTTP listeners used the authorized escalation.

Full raw test output is retained in `{A,B}-{host,race,minimum}.log`. Vet emitted no output. A's tests and the original frozen probes passed; the failures came from the separately supplied supplemental probes and independent reviewer reproductions. B passed the complete combined suite.

The operation-deadline reproduction intentionally returns a completed HTTP 503 at context expiration; it checks neutral health accounting, not a guarantee that a custom transport always obeys cancellation. The redirect reproduction uses a supplied policy that already stops redirects, so it requires one transport call for either candidate while distinguishing whether the supplied policy executes. The oversized-200 reproduction records the policies and asserts only the common 4096-byte result bound.

An additional identical `body_health_review_test.go` was copied to scratch A and B. Appending `go test -ldflags=-linkmode=external -count=1 -timeout=20s -v -run=TestPairBodyReadErrorIsNeutral ./...` to the common prefix passed for both candidates. It checks that a failed body read on HTTP 200 preserves closed-state failure history and releases half-open admission without proving recovery.

A separate disposable `B-mutant` copy changed only `status == http.StatusOK && err == nil` to `status == http.StatusOK` in `client.go`. The existing candidate suite, both frozen probe sets, and the initial reviewer tests all still passed with `go test -ldflags=-linkmode=external -count=1 -timeout=20s -v -skip=TestPairBodyReadErrorIsNeutral ./...`. Running the additional body-health reproduction with `-run=TestPairBodyReadErrorIsNeutral` failed both history and recovery cases. This is evidence of a regression-detection gap in B's supplied tests, not a defect in unmodified B. Raw output is in the four body-health/mutation logs.
