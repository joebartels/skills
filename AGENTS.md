# Repository agent instructions

For Go quality build work, read the [authoring guide](docs/go-quality-build/README.md) and the relevant runtime skills before editing.

## Repository content

Every committed file must have a concrete current use for a skill included in this repository: using, maintaining or distributing it, documenting it, or meaningfully evaluating it. Shared tooling and repository infrastructure must serve those same needs. A new skill and its necessary support files may land together once the skill is accepted.

Keep candidate experiments and generated outputs outside the checkout until retention is justified. If a candidate is abandoned or fails acceptance, exclude its candidate-only guidance, fixtures, probes, tests, helpers and documentation from commits and PRs. Do not deliver an evaluation harness for a skill that did not make the cut.

Possible future reuse is not sufficient justification. Retain an experimental artifact only if it is actually used by a retained skill or its evaluation; identify that consumer and integrate it into the relevant workflow. Before committing, review the final diff and remove orphaned scaffolding, archives and files that cannot meet this bar.

## Documentation

Keep maintained documentation concise and evergreen. Include only guidance needed to use or develop the skills. Keep task history, status logs, review outcomes, and future proposals out of documentation. Use Git history for historical context when needed.
