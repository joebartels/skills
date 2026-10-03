Selected and opened:

- **go-concurrency-and-ownership** — directly applies to bounded admission, dispatching available work, supervising failures during input/acquisition, and waiting for worker completion before releasing resources.
- **go-interfaces-and-composition** — applies if implementing this changes host composition or resource lifecycle ownership.
- **go-behavior-tests** — applies if implementing the requested behavioral changes and their supported contract.
- **go-test-isolation** — applies if tests for those changes involve workers, channels, resource lifetimes, or controlled concurrency.

I read the workspace’s required `/Users/jb/.codex/RTK.md` and used `rtk cat` for the instruction and skill files. I decline **go-api-contracts** because no API or external contract change was requested, and **go-package-boundaries** because no package responsibility or import direction change was requested. I made no edits, ran no tests, and did not inspect task files.