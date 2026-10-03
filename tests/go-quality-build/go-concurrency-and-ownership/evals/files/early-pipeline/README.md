# Consumer completion and callback ownership

Implement runPipeline and focused tests; retain signature, package, standard-library-only dependencies and Go 1.22 minimum. ctx is nonnil and capacity positive. produce alone sends ordered values and consume alone receives. Both callbacks cooperate with context-aware channel operations. Cleanup may be held after cancellation and remains part of callback lifetime. Callbacks do not close the channel. The pipeline owns closing it only after production finishes.

Run both callbacks concurrently. Normal consumer completion can occur early after enough values; stop and join production and succeed when no independent failure or caller cancellation occurred. Producer failure must stop a blocked consumer. Join both actual callback returns on every outcome, including held cleanup. Accepted values/effects stay accepted. An already-canceled caller invokes neither callback.

Preserve independent producer and consumer failures, including wrapped/joined context errors and bare DeadlineExceeded from an independent operation. Only the exact shared-context standard error returned solely to acknowledge coordinated stopping may be suppressed when the caller did not cancel. Caller cancellation classification and custom cause remain inspectable. A custom cause can be noncomparable. Do not use broad cancellation classification to infer error origin.

The capacity bounds the channel buffer; it does not prove completion or limit arbitrary external work. No arbitrary blocking I/O, signals/process host, framework, new dependency or panic recovery is needed.
