Selected and opened:

- **`go-context-and-deadlines`** — directly applies to carrying caller cancellation through body read and close, bounding the whole operation once, and narrowing individual requests.
- **`go-behavior-tests`** — applies if this behavior is implemented, since it changes observable cancellation and deadline behavior. Its local reference, `references/behavior-observations.md`, is for choosing test observations; I didn’t open it because this was only a selection probe.

I decline **`go-api-contracts`** as not established by the prompt: it doesn’t say this changes an exported or otherwise supported external contract. The interface, package-boundary, and test-isolation skills also don’t apply based on the described scope. I didn’t inspect task files, edit, test, browse externally, or delegate.