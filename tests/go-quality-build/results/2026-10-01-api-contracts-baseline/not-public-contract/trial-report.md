# Private helper baseline control

The initial fresh gpt-6-luna author read only this fixture and identified the local `now > deadline` comparison, but an automatic approval review rejected its edit to this disposable copy as unrelated to the authorized skills workspace. It made no source changes. The controller then verified that the four files in this directory were byte-for-byte identical to the tracked `go-api-contracts` fixture, and retried the same direct patch successfully. This control is therefore controller-completed, not an independent model implementation.

`expired` remains private in `internal/cache`. The correction is `now >= deadline` for nonzero deadlines, with an equality regression test. No exported API, version, wrapper, migration guide, package boundary, or dependency changed.

Verification: `rtk proxy env GOCACHE=/private/tmp/go-quality-build-api-eval/go-cache go test ./...` passed; `rtk proxy env GOCACHE=/private/tmp/go-quality-build-api-eval/go-cache go vet ./...` passed. The initial author's attempts without a writable Go cache were blocked by the sandbox. The controller's patch and checks are recorded separately from the three true blind author baselines.
