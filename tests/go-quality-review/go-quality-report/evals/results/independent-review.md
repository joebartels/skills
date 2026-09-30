# Independent umbrella implementation review

No substantive correctness issue found in the requested implementation.

Reviewed `go-quality-report/SKILL.md`, all three references, `scripts/grade.py`, `evals/test_grade.py`, umbrella behavioral cases, `golang/scripts/validate.py`, the shared review contract, and relevant architecture, correctness, resilience and operations topic boundaries. Executed the calculator suite: 12 tests passed.

The grading implementation matches the shared table, preserves findings under incomplete coverage, rejects all-not-applicable coverage with defects, and requires positive evidence for clean grades. The orchestration instructions explicitly account for all topics, material cross-boundary obligations, snapshot changes, applicable missing skills, unavailable delegation and host model controls. Reconciliation preserves distinct fixes and topic consequences while preventing duplicate counts and clean-chunk dilution. Source-card mappings, original cards and correction addenda provide traceability.

## Optional follow-ups

- **Clarify report-validator scope.** `golang/scripts/validate.py:135` iterates only the nine entries in `contract["topics"]`; `--reports` never checks umbrella outputs or requires them to exist. Existing README wording describes `<skill-name>/<case-id>.md` generically. Document that this command validates topic report outputs only, and that umbrella behavioral responses require the separate semantic evaluation workflow. An umbrella-specific structural validator is optional because many cases intentionally request plans rather than full reports.
- **Normalize invalid-input errors.** In `grade.py`, JSON values such as an array for `coverage[].status` or `findings[].severity` raise uncaught `TypeError` during set/dictionary membership, whereas other invalid ledger values produce the intended concise `Invalid ledger:` error. This remains a nonzero failure and cannot award an incorrect grade, so it is a usability hardening idea rather than a material grading defect. Explicit string validation would make diagnostics consistent.
