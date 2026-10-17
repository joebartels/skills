# SQLite boundaries from Go

Inspect the actual driver, DSN and database topology. A bare :memory: database belongs to a connection; a Go pool can open another connection to a different database. For multi-connection state assertions use an intentionally configured shared database or a temporary file, own its pools and keep required connections alive. Foreign-key enforcement and other connection settings must apply to each relevant connection; executing a PRAGMA on one pool lease does not establish all leases' state.

SQLite permits only one writer at a time. Transaction begin mode, journal mode, busy handling and existing readers affect progress. A deferred read transaction that later upgrades to write can encounter SQLITE_BUSY_SNAPSHOT; retrying only its UPDATE does not refresh its snapshot. Restart the complete unit only when the operation permits it. Configure begin mode through the driver's transaction support; do not mix raw BEGIN/COMMIT with database/sql transaction ownership.

Busy/locked rejection is different from overselling or partially recording a basket. Check both returned outcomes and independently observed persisted rows. Use actual SQLite constraints/triggers to verify statement and deferred commit failures. Check the pinned driver's failed-Commit cleanup rather than assuming a later sql.Tx.Rollback can undo a Tx already completed by database/sql. Do not turn one driver's behavior into a cross-driver guarantee.

Primary sources: [SQLite isolation](https://sqlite.org/isolation.html), [transactions](https://sqlite.org/lang_transaction.html), [foreign keys](https://sqlite.org/foreignkeys.html), [go-sqlite3 driver configuration](https://github.com/mattn/go-sqlite3#connection-string).
