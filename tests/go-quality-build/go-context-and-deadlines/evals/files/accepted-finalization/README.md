# Accepted import receipts

Update ApplyBatch and add focused tests without changing its API, dependencies or Go 1.22 minimum. Inputs are finite, callbacks are nonnil and cooperative, and finalizationBudget is positive.

Apply items in order using the caller context. Count an item only when apply succeeds. Empty input succeeds, even with a canceled caller, and invokes neither callback. Check cancellation before each next item; never start a callback after observing cancellation. Retain the accepted count when stopping or on an apply error. A completed final item is not retroactively failed because the caller cancels during its successful callback.

When at least one item was accepted, finish must run exactly once before return, including after an apply failure or caller cancellation. It writes a required receipt for the accepted count. Give it caller metadata, its own finalizationBudget, and a context unaffected by caller cancellation or deadline; release that scope after the callback returns. No receipt is needed if zero items were accepted. Wait for cooperative finish rather than returning while it still uses caller-owned data.

Preserve inspectable independent apply and finish failures, including an error wrapping or joining context.Canceled or DeadlineExceeded. When processing stops on observed caller cancellation, expose the caller's standard cancellation classification and custom cause too. These error promises do not change empty or completed-success policy. Successful complete processing and finalization return nil. Do not add goroutines, a new dependency or a general job framework.
