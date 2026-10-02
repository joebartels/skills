---
name: go-error-contracts
description: Use when Go work changes error production, propagation, inspection, translation, or aggregation, including partial results and completion failures. Skip ordinary unchanged error returns and unrelated calculations or formatting.
---

# Go error contracts

Trace failure from its producer to the consumer that acts on it. Decide what the caller can inspect, what work remains usable, and what outcome the returned error promises.

Read the changed operation, callers, tests, documentation and supported Go versions. List its failure points, including failures after the apparent main work: iteration, buffered writes, flush, commit, close or joining work. Distinguish required completion from best-effort cleanup. If simultaneous failures have no defined precedence, choose and document a policy from the operation's contract; do not assume every diagnostic must be aggregated.

| Consumer need | Representation decision |
| --- | --- |
| Only a readable explanation | An ordinary error with useful operation context may suffice. |
| Stable classification | Reuse an established sentinel or type; introduce one only when callers need the distinction. |
| Structured facts | A small error type can expose actionable fields without forcing callers to parse text. |
| Caller-owned cause must remain inspectable | Return or wrap it using `%w`; inspect chains with `errors.Is` or appropriate typed inspection. |
| Private implementation cause must stay private | Translate to the promised domain error. Choose safe public text separately from whether unwrapping is allowed. |

Wrapping exposes a cause as part of the API. `%v` removes unwrapping but retains the cause's message; it does not redact secrets. Avoid both automatic wrapping at every layer and discarding a cause that callers legitimately inspect. Add context where it identifies the failed operation or input without leaking sensitive data or duplicating a chain of identical messages.

Preserve required direct sentinel equality as well as `errors.Is` behavior. A wrapper or `errors.Join(err, nil)` can change identity even when classification still succeeds. Aggregates also affect text and traversal: `errors.Unwrap` follows `Unwrap() error`, while `Is`/`As` also traverse `Unwrap() []error`. Check actual consumers and the supported toolchain; `Join` needs Go 1.20, while `AsType` needs Go 1.26.

Process usable data returned alongside an error before deciding the outcome. An `io.Reader` may return bytes and a failure together. EOF, malformed data and incomplete input need the operation's actual policy. Return the promised valid prefix or accepted count; do not claim all work succeeded merely because some output exists. Check terminal iterator errors and buffered completion errors. Close only resources the operation owns; whether cleanup failure invalidates the outcome is a separate decision from ownership.

When the contract requires both operation and completion failures, preserve lone errors directly and join only simultaneous failures:

```go
func completionError(primary, finish error) error {
    if primary == nil {
        return finish
    }
    if finish == nil {
        return primary
    }
    return errors.Join(primary, finish) // import "errors"; Go 1.20+
}
```

A successful `error` result must be a nil interface. Returning a nil concrete error pointer through `error` creates a non-nil interface; return nil explicitly on that success path. Do not normalize arbitrary interfaces with reflection or forbid legitimate nil receiver behavior.

Verify the changed decisions at the consumer boundary: classification/identity, partial data, successful nil errors, required simultaneous failures and finalization. For CLIs, exercise actual stdout, stderr and exit status after failures. Document only the supported outcome and usable results. Logging, retries and panic/recover policy follow the application's boundary and invariants; no universal log-or-return or recovery rule is required.

This skill owns failure representation and propagation. API contracts own whether supported behavior may change; composition owns acquisition/lifetime; concurrency owns stopping/joining work; telemetry and resilience own reporting systems and retries. It runs independently of those skills.
