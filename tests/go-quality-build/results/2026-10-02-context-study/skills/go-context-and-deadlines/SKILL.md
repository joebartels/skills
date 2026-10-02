---
name: go-context-and-deadlines
description: Use when Go work changes cancellation propagation, operation or stage time budgets, cancellation error/cause decisions, or required finalization after caller cancellation. Skip pure calculations and synchronization-only changes without a context or budget decision.
---

# Go context and deadlines

Make cancellation and time limits follow the operation's contract, with explicit observation points and owned scopes. Preserve supported signatures and Go versions; a context belongs at a cancellation-aware boundary, not on every helper.

## Decide when cancellation affects the result

Read the affected callers, documentation and tests. Establish which work may start, what counts as accepted progress, which failures remain inspectable, and when success becomes complete. Apply those decisions at their actual boundaries:

| Boundary | Implementation decision |
| --- | --- |
| Entry | If already-canceled calls must fail, check before any work or zero-work success return. If the contract permits empty success, preserve it. |
| Next unit of work | Check before admitting the next callback, request or stage; retain the accepted prefix when stopping. |
| Independent failure | Sample `ctx.Err()` once to decide whether cancellation is observed. When nonnil, read `context.Cause(ctx)` after that sample and preserve the promised classification/cause alongside independent errors. If the sample is nil, return the operation failure under its policy. |
| Completed success | Return the accepted result according to its completion rule. A blanket final cancellation check can reject work that already succeeded. |

`ctx.Err()` supplies cancellation classification; a custom `context.Cause` may have a different identity or type. Assemble only the errors the contract exposes, using `errors.Join` when several must remain inspectable. Legal causes can contain slices or maps: preserve them without direct equality of arbitrary error interfaces. Inspect such types with appropriate `errors.As` assertions. General wrapping/exposure and API evolution remain their own decisions.

## Scope the budget through the real operation

Propagate the caller context when its scope fits. Derive a child when a local budget or cancel owner differs. For a total budget, create one operation context at entry; derive every stage from that same scope. A stage timeout narrows the remaining allowance and inherits an earlier parent deadline. Recreating a total timeout inside each stage resets the allowance.

The creator retains cancel ownership until an explicit contract transfers it; name the recipient and its release boundary. The current owner calls cancel on every owned exit path, rather than canceling in a creator that returns live work. Put a short-lived stage in a helper or cancel explicitly per iteration; loop-level deferred cancels retain stage resources until the whole call returns. Keep the stage context live through all work it governs, including response-body reading and closure after headers arrive. Borrowed clients stay owned by their caller; close acquired bodies. Context cancellation only reaches cooperating work through the boundary that observes it; it does not stop an arbitrary blocking reader or prove a goroutine has joined.

Use context values for needed request metadata crossing API boundaries; supply dependencies and configuration through existing parameters or owned fields.

## Bound required work that survives caller cancellation

Detach only when the contract requires accepted work to be finalized after the caller stops. On Go1.21+, `context.WithoutCancel` retains values but removes the inherited deadline, Done signal, error and cancellation cause. Give the detached scope its own stated limit and cancel owner:

```go
func finalizeAccepted(ctx context.Context, budget time.Duration, finish func(context.Context) error) error {
	finalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), budget)
	defer cancel()
	return finish(finalCtx)
}
```

This waits for cooperative finalization to return. Retain the processing outcome separately and combine any finalization failure according to the declared policy. A deadline requests stopping; define remaining work/ownership if it cannot cooperate. `Cause`/`WithCancelCause` require Go1.20; `WithoutCancel`/`AfterFunc` require Go1.21. Choose APIs supported by the project's minimum rather than silently upgrading it. If using `context.AfterFunc`, stopping its registration does not wait for an already-started callback; observe completion separately.

## Verify the decisions

Exercise the promised entry/empty cases, accepted-prefix and completed-success decisions, and coincident independent failure with a legal custom cause. For shared budgets, compare absolute request deadlines across stages and with an earlier parent. Observe the request context through body consumption/closure. For detached finalization, cancel the caller, observe a live bounded finalization scope, and wait for the actual completion/failure effect. Use existing real callback/client/process boundaries and bounded event waits; select only applicable observations. Synchronization mechanics, general test strategy and telemetry belong outside this skill.
