# Report materialization worker

This small package stores the first bounded sample of an account's ledger. The worker shares its caller-owned `*sql.DB` with other work. Production uses drivers supporting context-aware database calls and transactions. The pool can have one connection; driver result sets can occupy that connection until closed.

The established schema is:

```sql
ledger(account TEXT, sequence INTEGER, amount BIGINT)
accounts(account TEXT PRIMARY KEY, currency TEXT)
reports(account TEXT PRIMARY KEY, currency TEXT, entry_count INTEGER, total BIGINT)
```

`Materialize(ctx, account, maxEntries)` selects ledger amounts ordered by sequence, keeps at most `maxEntries`, reads the account currency, deletes the old report, and inserts the replacement. Existing SQL constants encode the established statements and bind order. `maxEntries < 1` is rejected before database work. Amount conversion and missing accounts are errors; a zero-row sample is a valid report.

## Task

Make Materialize safe for the report worker's new deployment. It samples at most maxEntries ledger amounts in sequence order, reads the account currency, and replaces that account's saved report. Keep the public API, SQL schema, parameter order, and Go 1.22 support. The caller owns the DB and may configure a pool of one connection.

A successful result must correspond to a committed replacement. Read the sample and currency and replace the report as one transaction, including helper work. If any read, conversion, write, or commit fails, return a zero Report and an error preserving the underlying cause for errors.Is; never publish an incomplete sample. A failed pre-commit operation must leave the previous report intact. A commit error must be reported, without assuming whether a real database committed. Release row and transaction resources on every exit and before dependent database operations; a failed call must leave the pool usable. Carry the caller context through database work so cancellation can stop an operation already in flight.

The existing maxEntries validation and bounded-sample behavior must remain. Keep the change focused; add useful regression tests without introducing third-party dependencies. Run go test ./... .

The checked-in tests cover the original happy path. They use a minimal local driver, not a SQL engine. No external service or dependency is needed.
