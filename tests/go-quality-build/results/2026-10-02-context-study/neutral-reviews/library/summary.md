# Frozen neutral library review

Candidate-1: overall C; Architecture A, Code Quality A, Correctness A, Testing C, Resilience A, Resources A, Reproducibility A; Security/Deployment not applicable within bounded source area. Testing has one major same-dynamic-type non-comparable cause panic regression gap and one moderate canceled-empty coverage gap.

Candidate-2: overall C; same topic grades/states. Testing has the same major panic regression gap and one moderate active-empty callback non-invocation assertion gap.

Both unmodified implementations independently pass Go 1.22.12/1.26.5 builds, author/provided contract suites, vet/format checks and current-toolchain exercised race/shuffled checks. Both author suites miss the frozen unsafe equality mutation and kill the frozen lost-custom-cause mutation. Supplementary empty-input mutations are independently distinguished and excluded from frozen efficacy.

Full cards, all-nine coverage and grade ledgers: [candidate-1/report.md](candidate-1/report.md), [candidate-2/report.md](candidate-2/report.md). Exact commands/output: [checks.json](checks.json). No candidate source was modified; no arm identity or external planning/writing evidence was inspected.
