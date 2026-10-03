# Behavior observations

Read the relevant section for the behavior at risk. These examples guide
assertion choices, not mandatory test forms or production abstractions.

## Failure stage and retained state

An error alone rarely proves a stateful failure contract. Start from what
the caller is promised to retain, then place the failure at a point that can
damage it.

| Failure point | Relevant observations |
| --- | --- |
| Validation before mutation | Rejection and unchanged prior state |
| Read/write after some progress | Promised prefix/rollback policy and exact remaining state |
| Replacing an already accepted value | Prior value survives rejected replacement, if promised |
| After acceptance of earlier operations | Earlier effects remain, later effects do not occur |
| Publication or cleanup | Complete visible output; promised result/error still reaches the caller |

A filesystem obstruction that prevents opening a file does not exercise
truncation followed by a partial write. [os.WriteFile](https://pkg.go.dev/os#WriteFile)
can fail after changing its target; [io.Writer](https://pkg.go.dev/io#Writer)
can report both progress and error. A double returning an error before any
side effect cannot support a broader retention claim.

Use a controlled failure at the actual boundary. Depending on the project,
that may be a faithful existing I/O seam or a supported-platform child process
with a finite file-size limit. Bound the child and keep limits local to it;
do not exhaust disk space or alter the parent process's limits. Check prior
bytes, rejected/later effects and error/process status together. Platform
skips and substituted dependencies narrow the verified claim.

Tests may expose a production retention defect. Correct it at the responsible
operation and keep its regression test; do not change the promised failure
policy just to make the test pass. No particular storage interface, transaction
or commit mechanism is required by this writing skill.

## Independent work and cleanup errors

When callers are promised both causes, asserting a joined-error case alone
can miss code that reports success whenever cleanup succeeds. Each independent
combination protects a different branch. The following complete illustration
uses standard-library APIs available in Go 1.22 (errors.Join itself requires
Go 1.20). Adapt the tests to the real operation; keep the actual production
contract and supported error exposure.

```go
package finish_test

import (
	"errors"
	"testing"
)

func finish(work, closeResource func() error) error {
	workErr := work()
	return errors.Join(workErr, closeResource())
}

func TestFinishErrorCombinations(t *testing.T) {
	workErr := errors.New("work failed")
	closeErr := errors.New("close failed")
	for _, tc := range []struct {
		name        string
		work, close error
	}{
		{"success", nil, nil},
		{"work fails, close succeeds", workErr, nil},
		{"work succeeds, close fails", nil, closeErr},
		{"both fail", workErr, closeErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			closes := 0
			got := finish(
				func() error { return tc.work },
				func() error { closes++; return tc.close },
			)
			if tc.work == nil && tc.close == nil && got != nil {
				t.Errorf("finish = %v, want nil", got)
			}
			for _, want := range []error{tc.work, tc.close} {
				if want != nil && !errors.Is(got, want) {
					t.Errorf("finish = %v, want cause %v", got, want)
				}
			}
			if closes != 1 {
				t.Errorf("close calls = %d, want 1", closes)
			}
		})
	}
}
```

Changing finish to return nil after a successful close must fail the
work-fails/close-succeeds case. Match error text exactly only when that text
is a supported contract. Use [errors.Is/As](https://pkg.go.dev/errors) for
promised identity/type, not for every incidental implementation error.

An API accepting error values does not imply they are comparable pointer
sentinels. When the changed path classifies callback or cancellation errors,
use a valid custom value (for example one containing a slice) to challenge
comparison assumptions. Include wrapped/joined forms when those decisions are
promised, and retain an independently identifiable failure alongside cancellation.
Use the API's supported matching/type semantics and check release/return effects;
no universal custom-error matrix is required for unrelated paths.

## Boundaries and independent oracles

- **Library:** use consumer-visible outputs and independently derived values.
  An external test package is useful for exported usage; same-package tests
  suit private contracts. Choose access deliberately.
- **Wire/file:** compare fixed known representations and complete meaningful
  fields. Round trips supplement an oracle; matching bugs on both sides can
  preserve them. Decode with an independent standard format reader where that
  is the consumer contract, rather than the same application codec.
- **CLI:** a run helper is convenient for case matrices, while the real built
  executable establishes exit status and stream routing. Inspect effects
  after failure, including accepted/rejected records, not only a diagnostic.
- **Service:** choose the boundary claimed. A handler recorder tests handler
  outputs; real HTTP client/server checks establish protocol behavior. Check
  complete success output, independent rejected fields and retained state.
  Distinguish the exact accepted set from conventional categories: if only
  status 200 is accepted, a valid-body 206 rejection checks a different branch
  than a 503 or malformed body. Select discriminating neighbors from the
  actual contract rather than enumerating every possible status.
- **Worker/host:** cancellation-only cases miss successful recurrence. Observe
  the later successful callback and the event from which its interval starts.
  Verify join/release through the host, after controlled callback cleanup.

Use [testing](https://pkg.go.dev/testing), [Go subtests](https://go.dev/blog/subtests)
and [fuzzing](https://go.dev/doc/security/fuzz/) for API details. Keep each check's
claim within what its fixtures and assertions actually observe.
