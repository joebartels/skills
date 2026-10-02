For the described Go change, I selected and opened:

- **go-context-and-deadlines** — it directly covers caller cancellation, one operation-wide budget, narrower request budgets, and keeping each request context live through body read and close.
- **go-behavior-tests** and its *behavior observations* reference — the change affects observable behavior, including which request starts next and what error reaches the caller.
- **go-test-isolation** and its *isolation patterns* reference — HTTP cancellation tests may need controlled requests, response bodies, and timing.

I declined **go-api-contracts**, **go-interfaces-and-composition**, and **go-package-boundaries**: the description does not establish an exported contract change, dependency-wiring change, or package change. I read no other task, made no edits, and ran no tests.