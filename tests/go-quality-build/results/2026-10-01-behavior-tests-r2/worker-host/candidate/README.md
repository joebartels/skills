# Host-owned sweeper

The old host does one sweep. Extend Run(ctx, interval, sweep, release) to own a
periodic worker. The first sweep starts immediately unless ctx is already
canceled. Each successful cycle waits the full positive interval after
callback completion; calls never overlap. A callback error ends work and is
returned with its error identity preserved. Parent cancellation stops future
cycles and is a normal nil worker result. Nonpositive intervals reject.

The host owns release: cancel and wait for the worker and its callback cleanup
before invoking release, and before returning. Invoke release exactly once
on normal cancellation and callback failure; join any release error with the
worker error so callers can inspect both. An invalid interval is rejected
before starting work and does not acquire/release resources. The callback
may block until its context is canceled, then finish its cleanup. Preserve
Run's public signature, Go 1.22, and standard-library dependencies.

Cancellation requests stopping and Run waits for the active callback to finish;
callbacks must cooperate with their context for shutdown to complete. A callback
returning nil after cancellation gives a normal nil worker result. Any error the
callback returns, even while cancellation is happening, remains a worker error
and is joined with a release error. The worker's owned context is canceled before
release runs, including when a callback fails.
