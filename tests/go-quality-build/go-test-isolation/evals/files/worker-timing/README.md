# Poller test work

`Poll(ctx context.Context, interval time.Duration, callback func(context.Context) error) error`
is synchronous and owns sequential callbacks until caller cancellation or
callback failure. The first callback starts immediately unless ctx is already
canceled. Later callbacks start no earlier than interval after the preceding
callback finishes. Callback errors are returned with their identity preserved,
including errors concurrent with cancellation. Cancellation alone is normal
completion (nil). An invalid nonpositive interval returns an error without
calling the callback. Callbacks receive the caller's context and cooperate
with its cancellation. Poll returns only after the active callback and its
cleanup finish; callers own any resources used by those callbacks.

The implementation already supports these contracts. Replace the guessed
readiness in existing test setup and add reliable recurrence, cancellation,
fixture teardown and independent-invocation tests. Preserve useful old checks.
Keep public API, Go 1.22 minimum and standard-library dependencies unchanged.
The project does not require a clock framework or exported testing hooks.
