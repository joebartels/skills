# Required finalization

Update Run and its tests, preserving its signature and the supplied command host. Go 1.22, standard library only. Callbacks are nonnil and cooperative; the parent context is nonnil and finalizationBudget positive.

Jobs run sequentially. Check cancellation before starting work; each nil apply result increments accepted. Keep accepted count and inspectable independent failures. Stop processing on cancellation/error; observed cancellation retains standard classification and cause. A nil final apply result completes processing successfully without a late cancellation veto.

When accepted is nonzero, call finalize exactly once with that count, under a separate bounded scope that survives caller cancellation. Its owner waits for cooperative completion before Run returns; detachment alone supplies no bound/completion. Release derived cancellation resources. Retain processing failure and independent finalization failure together. Empty/unaccepted work requires no finalization. A finalization failure is a failed operation.

The supplied command already owns parsing/signals/callbacks. Positive integer jobs print `accepted N`; job 0 signals `waiting` and blocks cooperatively until cancellation. `--receipt PATH` writes `accepted=COUNT` plus newline at finalization. Exit 0 means complete success; processing/finalization failure exits 2 with stderr. Do not replace host lifecycle or add test-only command options. Test actual child output, status and receipt after interruption, and a real receipt-write failure. The interrupt mechanism is platform-specific; disclose unsupported platforms.
