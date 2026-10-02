I selected **go-behavior-tests** because the requested fix changes behavior at a boundary and calls for a focused regression test. I opened its `SKILL.md` and its local *behavior observations* reference for guidance on boundary assertions.

I declined **go-api-contracts**, **go-context-and-deadlines**, **go-interfaces-and-composition**, **go-package-boundaries**, and **go-test-isolation**: the supplied task establishes no external contract, cancellation, dependency or package change, or test state to isolate.

I did not open the referenced `/Users/jb/.codex/RTK.md` because it is outside the workspace. I made no edits and ran no tests.