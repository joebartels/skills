# Isolation patterns

Use these patterns when the actual boundary requires them. They are options,
not a required test architecture or framework.

## Subtest lifetime and focused selection

A parent t.Cleanup runs after its descendants; its defer runs when the parent
body returns. Parallel children pause until that return. Own fixtures at the
test whose remaining work needs them: parent-owned roots when the parent reads
after children, child-owned roots when all use ends inside the child.

A serial grouping subtest waits for its parallel descendants before returning
to outer assertions. It does not guarantee every sibling ran: -run can select
one child or one nested table row. Final checks must reflect expected executed
work, not the entire table. Avoid shared mutable counters without synchronization
when children run in parallel. Test a consequential focused leaf as well as the
whole suite; repeating the full suite alone cannot expose selection dependence.

Cleanup is last-in, first-out. If teardown must release a blocked callback,
cancel/join it and then close a borrowed file, register those actions in the
opposite order or use one owner that explicitly enforces the sequence. Do not
close the file first merely because it was acquired first in the test body.

## Go1.22-compatible event and cleanup illustration

This complete illustration uses only APIs available by Go1.22. The goroutine
stands for cooperative project work borrowing the file; substitute the actual
operation in a real test. It publishes errors to the test goroutine, observes
startup, and installs cancellation/join cleanup before any fatal assertion.
The safety bound diagnoses stalls; it is not a readiness estimate.

```go
package example_test

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestBorrowedFixture(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "borrowed")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close fixture: %v", err)
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan error, 1)
	finished := make(chan struct{})
	t.Cleanup(func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Error("borrower did not finish during cleanup")
		}
	})
	go func() {
		defer close(finished)
		_, err := file.Stat()
		started <- err
		<-ctx.Done()
	}()
	select {
	case err := <-started:
		if err != nil {
			t.Fatalf("borrower fixture use: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("borrower did not start")
	}
	cancel()
	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("borrower did not finish after cancellation")
	}
	if _, err := file.Stat(); err != nil {
		t.Fatalf("caller fixture unavailable after join: %v", err)
	}
}
```

An operation that ignores cancellation cannot be forcibly stopped by this
pattern. A timed-out join reports a failed guarantee; it does not justify claiming
safe resource release. Use a separately terminable process where arbitrary
non-cooperation must be contained. t.Context (Go1.24+) is canceled before cleanup
starts, but the test still owns joining and resource-release order.

## Dependency fidelity

| Fixture | Supports | Requires another observation when |
| --- | --- | --- |
| Local temporary files | Actual path/data/ownership behavior | Cross-platform replacement/permission semantics matter |
| Handwritten HTTP transport/body | Constructed request, statuses, read errors, close counts | Real framing/transport cancellation is promised |
| httptest.ResponseRecorder | Handler response behavior | Network connection behavior matters |
| Local httptest server/client | Real local request/response and context behavior | External DNS/TLS/service behavior matters |
| Actual command with explicit environment | Process status/streams/configuration | Deployment environment or other platforms matter |
| Transaction fixture | Writes performed through that transaction | Another client commits or a worker writes outside it |

Do not make unrelated integration infrastructure mandatory. Keep reduced-scope
results explicit rather than calling a substitute a successful real integration.

## Timing choices

Channels/control events work in Go1.22 projects without a production clock seam.
For recurrence, hold a started callback past the interval, observe its completion
and the later start, and check the documented interval from completion. Allow
late scheduling where the contract permits it; avoid a tight maximum duration
that merely grades machine speed.

Stable synctest.Test is optional for Go1.25+ in-process tests: its bubble waits
for contained goroutines and advances fake time at durable blocking. Real socket
I/O and child processes remain external; fake time cannot make those observations
deterministic. Keep separate real-boundary checks where they establish the claim.
