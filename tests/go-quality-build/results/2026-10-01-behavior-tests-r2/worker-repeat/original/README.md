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
