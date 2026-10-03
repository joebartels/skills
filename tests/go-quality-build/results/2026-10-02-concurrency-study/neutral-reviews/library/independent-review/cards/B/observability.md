## Observability & Resilience — Not applicable

Scope: Candidate B, complete bounded code area `candidates/B/{README.md,go.mod,totals.go,totals_test.go}`; original README/source are the behavioral baseline. Go 1.22 minimum, standard library only. Exact file SHA-256 values are in [manifest](../../manifest.json).

Skill used: [go-observability-and-resilience](/private/tmp/go-neutral-library-study/review-guidance/go-observability-and-resilience/SKILL.md), including [observability-resilience-decisions.md](/private/tmp/go-neutral-library-study/review-guidance/go-observability-and-resilience/references/observability-resilience-decisions.md).

Coverage: The complete supplied library, input/ownership model and module metadata were inspected to decide applicability.

Rationale: Not applicable in this bounded code area: synchronous local arithmetic has no fallible external dependency, retry, queue, telemetry export or service lifecycle contract. Mutex ownership/liveness is assessed under Correctness and Performance. No context, instrumentation or operational machinery is required by the supplied contract.

Limits: This is justified irrelevance within the bounded local code area, not unavailable material evidence or an approval of an external application/release. Surrounding application, deployment and policy evidence is outside the requested scope.
