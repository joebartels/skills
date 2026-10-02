# Standalone indexer author report

## Changes

- Implemented `Index.ApplyCSV` with two-field rows, key/nonblank-text validation, original text preservation, ordered duplicate replacement and accepted-prefix effects. Validation errors preserve `ErrInvalidRecord`; field-count errors preserve both validation and CSV error causes. Existing `Put`, `Get`, `Open`, exported signatures and arbitrary Put text behavior remain intact.
- Implemented `Refresh` using the supplied HTTP client and context: GET, status-200-only acceptance, exactly one JSON array, nonblank text/key validation, ordered compact JSON plus newline and an atomic same-directory temporary-file rename. Empty arrays publish `[]\n`; acquired response bodies close on success and rejection. Publication failures remove temporary files and retain the prior destination.
- Implemented `Serve` as a synchronous callback loop. Positive intervals are checked before callbacks or release; canceled startup skips refresh; each next interval begins after callback completion. Cancellation waits for callback cleanup, then releases exactly once. Callback and release errors are joined without losing identity. A callback-returned error remains a callback error even if parent cancellation also occurs; ordinary context cancellation when the callback returns nil yields nil.
- Preserved the existing positional `put` grammar, including leading-dash directory/text arguments. Added stdin CSV `apply` and validated HTTP/HTTPS `watch` flags. The process host owns signal cancellation and its cloned HTTP transport, joins refresh through `Serve`, then closes client idle connections. Watch treats a Refresh error caused by its canceled host context as normal shutdown. Success has no stdout; errors/usage diagnose on stderr and exit 2.
- Kept the existing library and command packages, concrete supplied HTTP client and function callbacks; added no interfaces, options framework, dependencies or shared mutable lifecycle state. `go.mod` remains Go 1.22.

## Regression evidence

`index_test.go` checks original Put replacement/empty-and-whitespace text, invalid keys/Get error identity, independent roots, quoted and multiline CSV text, duplicate order, retained prefixes/rejected replacements, no later-row writes and read/publication failures. Public function/method assignments compile in an external consumer test package.

`refresh_test.go` checks fixed independently known lowercase JSON bytes, exact compact snapshot output, record order/original whitespace/duplicates, empty-array bytes, exactly one supplied-client GET and caller-context forwarding, body closure, malformed/null/object/trailing JSON/status/record rejection, fetch/read/publication failures, context cancellation and exact prior-byte retention. A reader opened before POSIX rename observes the entire old snapshot; a later opener observes the entire new snapshot.

`write_failure_unix_test.go` launches bounded disposable test children with local `RLIMIT_FSIZE=4096` and ignored SIGXFSZ. 128 KiB Put, CSV replacement and Refresh publications fail with actual EFBIG after partial temporary-file writing. Assertions protect exact old or accepted-prefix values, absence of later CSV effects and temporary-file cleanup. Parent process limits and signal policy stay untouched.

`serve_test.go` checks invalid intervals without release, canceled startup, independent work/release error combinations, exact caller context, held cancellation cleanup before release/return, a held first callback longer than interval followed by a next call that still waits from completion, absence of overlap, successful recurrence and independently stoppable active invocations. Synchronization waits are finite.

`cmd/indexer/main_test.go` builds and invokes the actual command with bounded subprocess deadlines, asserting exit status, both streams and persisted Put/CSV success/failure effects. `watch_unix_test.go` uses real localhost HTTP servers and real commands to cover failed first refresh, repeated complete snapshot publication, SIGINT and SIGTERM termination, no work after process return, and cancellation of an in-flight partial JSON response with exact destination retention.

## Actual verification

All verification uses `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`; shell commands are prefixed with `rtk`, using `rtk proxy` for raw output. Exact command arrays, working directory, non-secret environment, stdout, stderr, exit codes and deadlines are retained in `checks.json`.

- `gofmt -s -w .`: exit 0; subsequent `gofmt -l .` produced no output.
- Initial sandbox `go test -timeout=60s ./...`: exit 1. Library tests passed. Actual watch HTTP process tests failed specifically at localhost bind with `operation not permitted`. The failure is preserved in `checks.json`.
- Approved unchanged-source rerun outside the sandbox, `go test -timeout=60s ./...`: exit 0, both packages passed. Its separate evidence is in `listener-rerun.json`.
- `go vet ./...`: exit 0.
- `go test -race -count=1 -timeout=60s ./...`: exit 0, both packages passed.
- `go test -shuffle=on -count=3 -timeout=60s ./...`: exit 0, both packages passed.
- `go version`: `go1.26.5 darwin/arm64`.

- Staticcheck was unavailable in the existing PATH; no installation was attempted.

## Execution limits and remaining risks

The tested operating system is macOS/POSIX. Real localhost listener/process tests required the approved unsandboxed boundary; the denied sandbox run was retained rather than skipped. Partial-write and signal tests are build-constrained to Darwin/Linux/FreeBSD, but only Darwin was exercised here. Windows replacement and shutdown semantics are outside the execution claim.

The installed local toolchain is Go 1.26.5. The module minimum remains Go 1.22, all added standard-library APIs/language features predate or match 1.22, and vet passed, but no Go 1.22 toolchain execution was possible without installing another toolchain. The race check instruments the Go test binaries; the standalone CLI built by process tests uses ordinary `go build`.

Atomic visibility and preservation on failed writes were exercised; crash-durable fsync guarantees and OS-induced close failures are not asserted. Idle-connection closure is wired in the release callback after synchronous Serve completion; process tests observe joined HTTP cancellation and exit, while the library lifecycle tests directly observe callback-cleanup-before-release ordering.

No other author workspace, controller probes, evaluation JSON, repository/design files or unavailable writing/review guidance were inspected. No delegation, dependency/tool installation, commit or external application change occurred. All implementation files are in the isolated module; reports are in this task's report directory.
