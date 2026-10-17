# Atomic account adjustments inside a caller transaction

This package is used by a checkout flow that already owns an open native pgx
transaction. The caller stages its audit row before applying these adjustments,
then may keep working and choose whether to commit. Today an error midway through
`ApplyAdjustments` can leave its first updates staged, or leave the caller's
transaction aborted. Fix that isolation boundary.

Keep `ApplyAdjustments(ctx, parent, items)` and its exported types/errors. Use the
native `github.com/jackc/pgx/v5` API, pinned here to **v5.7.2**, and **Go 1.22**.
The table is `balances(account_id text primary key, cents bigint not null check
(cents >= 0))`. Each item adds `Delta` to that account. Preserve the missing-account
and invalid-empty-ID errors and the input order. An empty list is a successful
no-op. You may update the simple existing test double to reflect any new methods
the implementation legitimately needs; do not weaken its assertions.

Required behavior:

- Successful adjustments remain staged for the caller's later decision. This
  function must never commit or roll back the transaction it was given.
- Invalid items, missing accounts, or failed updates must undo every adjustment
  from this call, retain work staged before this call, and leave the parent open
  and usable when sub-operation rollback succeeds. The parent can itself be a
  pgx pseudo-nested transaction.
- Errors must remain discoverable with `errors.Is` / `errors.As`, including the
  original pgx error. If rollback also fails, preserve both errors so the caller
  can decide the parent's fate. Surface begin/release errors as well.
- A canceled caller context must not prevent rollback from being attempted.
  Cleanup must retain caller context values (for tracing), ignore the caller's
  cancellation/deadline, and have its own deadline no more than one second away.
  Do not add a goroutine or a new connection to perform cleanup.

There is no PostgreSQL service in this exercise. Keep the solution compatible
with the pinned pgx API and test its Go behavior with transaction doubles.
