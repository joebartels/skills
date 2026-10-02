# Frozen concurrency mutation meanings

These four targets are frozen before any author exposure; concrete compiling substitutions are recorded during controller conformance. They are relevant only where the authored implementation provides the corresponding mechanism. An unrelated build error or already failing original target does not qualify.

1. library: split compound update/publication into independent atomics; TestControllerCoherentSnapshots must detect Count/Sum incoherence. A scheduler yield exposes the otherwise legal interleaving; no race-detector-only credit.
2. service: skip Run join before cohort resource release; TestControllerPartialStartAndBlockedAdmission must observe Close/host return before held cleanup completion.
3. service: wait for input/admission without observing stopping; TestControllerCancellationWithOpenInput must observe failure to return before the caller closes its channel. Cleanup closes input only after the failed assertion, avoiding a hung test.
4. CLI: omit coordinated stopping on normal early consumer completion; TestControllerEarlyConsumerAndProducerFailure must observe a blocked producer that is not stopped/joined. Cleanup cancels the caller after the failed assertion, avoiding a hung test.

Pure sum regression is a control check, not a fifth candidate target. Later discoveries are supplementary and cannot replace these efficacy targets.
