---
name: go-concurrency-and-ownership
description: Use when Go work changes concurrent mutable state, goroutines, channels, bounded admission, asynchronous completion, or resources used by concurrent work. Skip private serial calculations and local expression fixes with no such changes.
---

# Go concurrency and ownership

Read the affected code and its contract. Identify the shared invariants, admission policy, blocking operations, and owner of each goroutine, channel and acquired resource. Preserve supported APIs, Go versions, dependencies and accepted effects. Keep serial work serial when it meets the contract.

Protect the observation the caller needs. One lock can guard a compound invariant; independently atomic fields do not make their combined snapshot coherent. Choose locks, handoff, bounded workers or an existing group from that invariant and workload. A channel transfers a value, not necessarily exclusive access to its referenced storage. Publish fully initialized state and establish who may mutate each remaining alias. Do not copy synchronization values after use. Keep arbitrary callbacks and blocking I/O outside a lock when they do not need that invariant; when they do, account for reentry, lock ordering and blocked progress.

For example, a count and sum that must describe the same accepted updates need one observation boundary:

```go
type totals struct {
	mu         sync.Mutex
	count, sum int64
}

func (t *totals) add(delta int64) {
	t.mu.Lock()
	t.count++
	t.sum += delta
	t.mu.Unlock()
}

func (t *totals) snapshot() (count, sum int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.count, t.sum
}
```

This uses `sync` and has a usable zero value. Scalar results are independent observations. Referenced results would need their own alias/lifetime policy. An RWMutex or lock-free redesign needs a concrete benefit rather than a preference.

Make admission and completion supervision progress together. If the contract requires available work to start, start it while capacity is available; do not wait for a full batch or input closure before processing the first job. Batching is valid when it preserves that progress. Observe worker failures while awaiting further input, capacity or acquisition; collecting every startup result before reading any worker result can deadlock when an acquisition awaits the stop that a worker failure should request. If acquisition can block, its cooperative stop path and completion must remain observable without bypassing the declared bound.

Define what capacity counts: running callbacks, admitted work, queued values or acquired resources. A lease-count bound lasts through completed release, including acquisitions whose work never starts. Buffers change where waiting occurs; they do not establish cancellation, safe closure or bounded total work by themselves. A limited group's admission call can itself block; the failure/stop controller must remain able to run. Add a dependency only when it improves the actual mechanism.

For each exit, account for admission, stopping, completion and release. On partial startup failure, worker failure or observed caller cancellation, stop admitting affected work, request its cooperative stop, observe every owned completion, then release resources after their users finish. Join the whole cohort first when resources share that lifecycle boundary. An acquired resource whose work never starts is still owned. One join error does not complete another worker; one close error does not excuse another owned release. Collect promised independent outcomes and accepted effects before returning. Cancellation requests stopping; a timeout does not kill a goroutine or authorize closing a resource still in use. If completion cannot be established, report the retained ownership required by the host contract.

Give blocking sends, receives and admission a termination path when the contract requires one. Normal early consumer completion may need to stop a blocked producer without becoming failure. Producer failure may need to stop a blocked consumer. Preserve independent errors according to the result contract; an error matching cancellation can still contain an independent failure. Assign channel closure to an owner that can establish all sends have finished. Receivers cannot safely close merely because they stop reading; not every channel needs closing. Do not add automatic panic recovery without a defined boundary and outcome policy.

Use version-supported completion mechanics. With WaitGroup, register work before launch, signal completion on every worker exit and wait before returning owned resources. Newer helpers require the effective Go version. Check the actual failure paths with held work/acquisition/release events, including sparse open input and normal early completion; arrange failure cleanup and bound waits so a regression yields a useful assertion. Race detection supports these observations but does not prove coherence or completion.

Keep host/API design with composition and compatibility; cancellation/time-budget policy with the operation contract; general assertion and fixture strategy with their testing owners. This skill supplies concurrent execution mechanics without requiring another skill to be installed.
