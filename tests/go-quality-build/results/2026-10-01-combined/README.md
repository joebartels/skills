# Combined evaluation preparation

## Scope and dispatch

The single `combined-evolution` case is in `tests/go-quality-build/combined/evals/evals.json`. A blind author should receive its `prompt` and the five listed fixture files, but not `expected_output` or `assertions`. The fixture is an original Go 1.22 module, `example.com/fielddesk`, with two existing features: check-ins and bulletins. It has no copied upstream code or prose. No model implementation, selected-skill record, skill-on result or independent grade exists at this preparation stage.

The proposed combined run should provide all three current build skills, record which were selected, save the source diff and fresh Go checks, and seek independent Architecture, Correctness and Testing review. Review should trace package/import direction, abstraction need, host-owned lifecycle and preservation of the documented Go/HTTP caller contract. Any conflict belongs to its decision owner. These are run instructions, not an outcome claim.

## Source hashes

SHA-256 from `rtk sha256sum` on 2026-10-01, relative to repository root:

```text
7f3fb0f9079bf752866ad2661fae2a74b6974dae02d1911407145ec0c7fd719e  tests/go-quality-build/combined/evals/evals.json
abfc48ed51b3169f42f4ecba8ba43fd8e71f60b42443bbab8ed96390f9ee7a93  tests/go-quality-build/combined/evals/files/combined-evolution/README.md
4d2a01dd430a5b58229ac516d2a3a87432b755e8a3141c371a223db49fa0f29a  tests/go-quality-build/combined/evals/files/combined-evolution/go.mod
30d0976653caff04cbf3afa12509eacfac76942c0af1df36d5c0562cc8b2756d  tests/go-quality-build/combined/evals/files/combined-evolution/fielddesk.go
357365937c0fe86e58233ad61a0f52a4918fedd43d4402444ef74798989bd2d1  tests/go-quality-build/combined/evals/files/combined-evolution/fielddesk_test.go
685220af9e1f0cb541e3ee3df8bcc44ab0778693697504e520d73e28586e8793  tests/go-quality-build/combined/evals/files/combined-evolution/cmd/server/main.go
```

## Preparation checks

- TDD validator red: `rtk python3 -B -m unittest discover -s tests/go-quality-build -p 'test_*.py'` failed one of ten tests because a present combined evaluation directory without `evals.json` was silently ignored. After adding combined discovery, the same command passed ten tests. The test also verifies a valid combined case counts and an escaping fixture path fails under the existing case rules.
- `rtk gofmt -l fielddesk.go fielddesk_test.go cmd/server/main.go`: no output.
- `rtk go test ./...`: two passed tests in two packages. `rtk go vet ./...`: no issues. Both ran in the fixture module with `go version go1.26.5 darwin/arm64`; declared minimum Go 1.22 execution and other platforms were not checked.
- `rtk go test -count=1 ./...`: two passed tests in two packages on a fresh, uncached fixture run.
- `rtk python3 -B scripts/validate.py`: passed 10 review skills, 120 review cases and 15 build cases.
- `rtk python3 -B -m unittest discover -s tests -p 'test*.py'`: six root tests passed. Build-eval tests are separately discovered from `tests/go-quality-build` as above.
- `rtk git diff --check`: passed after writing this evidence document and updating the design record.

The fixture tests verify existing behavior only. They do not demonstrate quality or correctness of the requested evolution. Next: freeze the input hashes and run a fresh combined implementation with expectations withheld.
