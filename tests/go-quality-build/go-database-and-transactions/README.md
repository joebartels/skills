# Database skill evaluations

Use `evals/evals.json` to give an author only the listed fixture files and task
prompt. Keep `probes/` unavailable during authoring. Compare fresh runs with the
same existing skill catalog, model settings and task; add only
`go-database-and-transactions` to the guided catalog. Let authors select relevant
skills. Include `pure-control` to check that local calculations skip this skill.

After authoring, run tests and controller checks in a disposable copy:

```sh
python3 tests/go-quality-build/go-database-and-transactions/check.py sqlite-basket /path/to/authored/module --go /path/to/go1.22/bin/go --race --count 10
```

Repeat for `sql-lifecycle`, `pgx-owned`, `pgx-savepoint` and `pure-control`, and
verify the project's other supported toolchains. SQLite requires CGO and a C
compiler. Module files pin go-sqlite3 v1.14.22 and pgx v5.7.2; provide a module
cache or permit their download. The runner preserves caller cache settings,
disables workspaces and automatic toolchain downloads, and never edits the
authored module. On macOS hosts where Go 1.22 binaries fail with missing LC_UUID,
use `--ldflags=-linkmode=external` with CGO enabled.

| Case | Boundary exercised |
| --- | --- |
| `sqlite-basket` | Actual file-backed SQLite constraints, concurrent pools, transaction participation and persisted rows; wrapped driver boundaries for in-flight cancellation and lost commit acknowledgement. |
| `sql-lifecycle` | Actual database/sql with a stateful driver double: early rows close, Scan, terminal iteration, commit and cancellation. |
| `pgx-owned` | Native pgx interfaces with doubles: rows, owned completion, borrowed helper work and canceled cleanup. |
| `pgx-savepoint` | Native pgx interfaces with stateful doubles: child ownership, nested parents, staged state and rollback causes. |
| `pure-control` | Skill selection and a local positive sum. |

Review error identity, public contracts, persisted state and unnecessary
complexity, not just test counts. The SQLite cancellation probe waits until a
driver statement has entered, then cancels either before it changes state or
after its successful result is ready. The latter exercises the database/sql
automatic-rollback/Commit interleaving; repeat it to expose scheduling variation.
The acknowledgement probe commits real SQLite state before reporting an error.
Neither probe estimates production failure frequency. The cancellation wrapper
recognizes a direct `INSERT INTO reservation_lines`; adjudicate an equivalent
form that bypasses that gate before treating its timeout as a code defect.

Doubles cannot establish PostgreSQL SQL semantics, isolation, wire cancellation
or durability. The owned-pgx probe recognizes the fixture's query shapes,
including ordinal projections and scalar/array attachment predicates. Adjudicate
an unsupported equivalent query before calling a checker rejection a defect.
Use a real PostgreSQL service when those engine claims matter. Keep execution
outputs and review reports outside the checkout.
