# Warehouse reservations

This package backs a small fulfillment worker. `Reserve` records a reservation,
decrements each requested SKU, and returns remaining stock for the reserved
SKUs. The original worker sent one line at a time. A new caller sends complete
multi-line baskets, and retries known rejected jobs with the same reservation ID.

Make `Reserve` safe for that caller while keeping the exported API, Go 1.22,
the pinned dependency, and schema compatible:

- A successful reservation durably records one header and its exact requested
  lines, decrements stock by those quantities, and returns the correct ID,
  total units, and remaining quantities.
- All errors return a zero `Receipt`. Known rejected or rolled-back attempts,
  including stock shortages, statement rejections, and deferred foreign-key
  commit failures with the supplied SQLite driver, must leave no effects from
  that attempt. Existing reservations and stock must be preserved. A known
  rejected attempt can be retried.
- If an I/O or commit error leaves the outcome uncertain, preserve its cause
  and return no usable receipt. Do not report success or automatically retry
  the operation as though rollback were certain. Reconciliation of uncertain
  outcomes is outside this task.
- Concurrent callers sharing stock must never sell the same unit twice. With
  one unit available, at most one reservation may succeed, and the recorded
  reservation lines must agree with the stock decrement. Ordinary SQLite
  contention errors may be returned to the worker for retry.
- Preserve `ErrInvalid` for empty IDs/customers, empty baskets, empty SKUs,
  nonpositive quantities, or duplicate SKUs; `ErrInsufficient` for a known SKU
  with too little stock; and errors wrapping `sql.ErrNoRows` for unknown SKUs.
  Preserve underlying context and database causes for `errors.Is`/`errors.As`.
  Duplicate reservation IDs are errors, not successful replays.
- `Available` remains usable independently. The caller owns the pool and may
  limit it to one connection. Operations must honor their context and release
  database resources on every return, so later calls can reuse the pool.

Keep this a small change to the existing operation and its stock-reading
helper. Private helper signatures may change. Do not replace SQLite, change
the schema, alter caller pool settings, or add a process-global coordination
mechanism. Equivalent correct SQL and internal organization are welcome.

The service opens file-backed SQLite databases using
`?_foreign_keys=on&_busy_timeout=1000&_txlock=immediate`. Customer records are
managed separately; a reservation for a missing customer must fail. The
deferred customer foreign key is part of the deployed schema. Stock quantities
and basket totals in this task fit in a Go `int`.

Add focused regression tests. Run `go test ./...` and `go test -race ./...`.
The driver uses CGO, so a C compiler is required. `schema.sql` and the current
test demonstrate setup; the package does not own schema migration or seeding.
