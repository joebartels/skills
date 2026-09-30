# Go quality review

This plugin contains one overall Go review skill and nine topic skills. Use [go-quality-report](skills/go-quality-report/SKILL.md) for a full changeset or code-area review. It delegates the relevant topic skills, reconciles shared findings, and produces one evidence-backed grade. Use a topic skill directly for a narrower review.

| Topic | Skill |
| --- | --- |
| Architecture and Design | [go-architecture-and-design](skills/go-architecture-and-design/SKILL.md) |
| Code Quality and Go Idioms | [go-code-quality-and-idioms](skills/go-code-quality-and-idioms/SKILL.md) |
| Correctness and Compatibility | [go-correctness-and-compatibility](skills/go-correctness-and-compatibility/SKILL.md) |
| Dependencies and Reproducibility | [go-dependencies-and-reproducibility](skills/go-dependencies-and-reproducibility/SKILL.md) |
| Deployment and Operations | [go-deployment-and-operations](skills/go-deployment-and-operations/SKILL.md) |
| Observability and Resilience | [go-observability-and-resilience](skills/go-observability-and-resilience/SKILL.md) |
| Performance and Resource Management | [go-performance-and-resource-management](skills/go-performance-and-resource-management/SKILL.md) |
| Security | [go-security](skills/go-security/SKILL.md) |
| Testing | [go-testing](skills/go-testing/SKILL.md) |

Each skill keeps its references in its own directory, so individual topic directories remain portable. The umbrella skill needs the applicable topic skills available. Evaluation cases and historical results live in the source repository's `tests/go-quality-review/` tree and are not bundled with this plugin.
