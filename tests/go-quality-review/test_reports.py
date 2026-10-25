"""Exercise the report validator on complete reports and inconsistent findings."""

import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("review_validator", ROOT / "scripts/validate.py")
VALIDATOR = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VALIDATOR)
CONTRACT = json.loads((ROOT / "tests/go-quality-review/shared/review-contract.json").read_text())
TOPIC = "Code Quality & Go Idioms"
SKILL = "go-code-quality-and-idioms"


OLD_REPORT = """## Code Quality & Go Idioms — A-
Scope: auth package at abc123, code-area review.
Coverage: all production source in this package.
Rationale: one local logging issue.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [G1] Guard clauses keep the normal path readable.

Bad

- [F1][minor][existing-in-scope] lookup.go:20 logs a failure its caller logs again.

Suggested changes

- [F1] Remove the inner log; check one record is emitted.

Limits: source inspection only; no executed checks.
"""


REPORT = """# Code Quality & Go Idioms — A-
Scope: auth package at abc123, code-area review.
Coverage: all production source in this package.
Rationale: one local logging issue.
Finding counts: critical=0, major=0, moderate=0, minor=1

## Findings summary

| ID | Severity | Topic | Finding / consequence | Evidence | Correction |
| --- | --- | --- | --- | --- | --- |
| F1 | minor | Code Quality & Go Idioms | Duplicate log records obscure one failure | lookup.go:20 and its caller | Remove the inner log |

## Finding details

### F1 — Duplicate logging

Severity: minor
Topic: Code Quality & Go Idioms
Status: existing-in-scope
Trigger: lookup fails and its caller logs the returned error.
Evidence: lookup.go:20 and handler.go:42 both log the same returned error.
Consequence: two records describe one failure.
Correction: remove the inner log and preserve the wrapped return.
Verification: capture logs for one failure and assert one record.

## Strengths

- [G1] lookup.go:10 uses guard clauses, keeping the normal path readable.

## Checks and limits

Source inspection only; no executed checks. The log assertion is proposed.
"""


def without_findings(grade="A"):
    text = REPORT.replace("— A-\n", f"— {grade}\n", 1)
    text = text.replace("minor=1", "minor=0")
    text = text.replace("Rationale: one local logging issue.", "Rationale: guard clauses preserve the contract with no actionable issue.")
    start = text.index("| F1 |")
    end = text.index("## Finding details")
    text = text[:start] + "None found.\n\n" + text[end:]
    start = text.index("### F1")
    end = text.index("## Strengths")
    return text[:start] + "None found.\n\n" + text[end:]


def umbrella(text=REPORT):
    text = text.replace(f"# {TOPIC}", "# Go Quality Report", 1)
    text += "\n## Report cards\n\n| Topic | Grade/state | Coverage | Cards |\n| --- | --- | --- | --- |\n"
    for topic in CONTRACT["topics"].values():
        grade = "A-" if topic == TOPIC else "Not applicable"
        text += f"| {topic} | {grade} | Assessed or justified irrelevant in manifest | card.md |\n"
    return text + "\n## Maintainability profile\n\nSee the seven-row profile in maintainability.md.\n"


class ReportTests(unittest.TestCase):
    def assertValid(self, text):
        self.assertEqual(VALIDATOR.check_report(text, CONTRACT), [])

    def assertInvalid(self, text, message):
        errors = VALIDATOR.check_report(text, CONTRACT)
        self.assertTrue(any(message in error for error in errors), errors)

    def test_complete_topic(self):
        self.assertValid(REPORT)

    def test_complete_umbrella(self):
        self.assertValid(umbrella())

    def test_zero_findings_keeps_table_and_sections(self):
        self.assertValid(without_findings())

    def test_not_applicable_has_no_counts(self):
        text = without_findings("Not applicable").replace("Finding counts: critical=0, major=0, moderate=0, minor=0\n", "")
        self.assertValid(text)

    def test_incomplete_preserves_confirmed_findings(self):
        self.assertValid(REPORT.replace("— A-\n", "— Insufficient evidence\n", 1))

    def test_incomplete_allows_assessed_scope_a_minus(self):
        text = REPORT.replace("— A-\n", "— Insufficient evidence\n", 1)
        self.assertValid(text.replace("Scope:", "Assessed-scope defect grade: A- (reviewed areas only)\nScope:", 1))

    def test_assessed_scope_grade_uses_observed_findings_without_tally(self):
        text = REPORT.replace("— A-\n", "— Insufficient evidence\n", 1)
        text = text.replace("Finding counts: critical=0, major=0, moderate=0, minor=1\n", "")
        self.assertValid(text.replace("Scope:", "Assessed-scope defect grade: A- (reviewed areas only)\nScope:", 1))
        self.assertInvalid(text.replace("Scope:", "Assessed-scope defect grade: F (reviewed areas only)\nScope:", 1), "assessed-scope grade conflicts")

    def test_rejects_invalid_assessed_scope_grade(self):
        text = REPORT.replace("— A-\n", "— Insufficient evidence\n", 1)
        for grade in ("XYZ", "A", "A+"):
            with self.subTest(grade=grade):
                self.assertInvalid(text.replace("Scope:", f"Assessed-scope defect grade: {grade}\nScope:", 1), "assessed-scope")

    def test_rejects_assessed_scope_grade_conflicting_with_findings(self):
        text = REPORT.replace("— A-\n", "— Insufficient evidence\n", 1)
        self.assertInvalid(text.replace("Scope:", "Assessed-scope defect grade: F (reviewed areas only)\nScope:", 1), "assessed-scope grade conflicts")

    def test_rejects_assessed_scope_grade_outside_incomplete_report(self):
        self.assertInvalid(REPORT.replace("Scope:", "Assessed-scope defect grade: A-\nScope:", 1), "assessed-scope grade requires")

    def test_rejects_repeated_or_empty_assessed_scope_grade(self):
        text = REPORT.replace("— A-\n", "— Insufficient evidence\n", 1)
        for field in ("Assessed-scope defect grade:\n", "Assessed-scope defect grade: A-\nAssessed-scope defect grade: A-\n"):
            with self.subTest(field=field):
                self.assertInvalid(text.replace("Scope:", field + "Scope:", 1), "assessed-scope")

    def test_incomplete_without_confirmed_findings_omits_counts(self):
        text = without_findings("Insufficient evidence").replace("Finding counts: critical=0, major=0, moderate=0, minor=0\n", "")
        self.assertValid(text)

    def test_escaped_pipes_in_table_evidence(self):
        self.assertValid(REPORT.replace("lookup.go:20 and its caller |", "`lookup.go:20 \\| caller.go:12` |"))

    def test_large_report_retains_every_finding(self):
        text = REPORT.replace("— A-\n", "— B+\n", 1).replace("minor=1", "minor=12")
        row = next(line for line in text.splitlines() if line.startswith("| F1 |"))
        detail = text[text.index("### F1"):text.index("## Strengths")]
        text = text.replace(row, "\n".join(row.replace("F1", f"F{i}") for i in range(1, 13)))
        text = text.replace(detail, "".join(detail.replace("F1", f"F{i}") for i in range(1, 13)))
        self.assertValid(text)

    def test_rejects_numeric_id_order_drift(self):
        text = REPORT.replace("— A-\n", "— B+\n", 1).replace("minor=1", "minor=2")
        row = next(line for line in text.splitlines() if line.startswith("| F1 |"))
        detail = text[text.index("### F1"):text.index("## Strengths")]
        text = text.replace(row, row.replace("F1", "F10") + "\n" + row.replace("F1", "F2"))
        text = text.replace(detail, detail.replace("F1", "F10") + detail.replace("F1", "F2"))
        self.assertInvalid(text, "finding order")

    def test_rejects_severity_order_drift(self):
        text = REPORT.replace("— A-\n", "— B\n", 1).replace("moderate=0", "moderate=1")
        row = next(line for line in text.splitlines() if line.startswith("| F1 |"))
        detail = text[text.index("### F1"):text.index("## Strengths")]
        text = text.replace(row, row + "\n" + row.replace("F1", "F2").replace("minor", "moderate"))
        text = text.replace(detail, detail + detail.replace("F1", "F2").replace("minor", "moderate"))
        self.assertInvalid(text, "finding order")

    def test_single_report_cli_accepts_current_and_rejects_old_format(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "report.md"
            for report, expected_status in ((REPORT, 0), (OLD_REPORT, 1)):
                with self.subTest(expected_status=expected_status):
                    path.write_text(report)
                    result = subprocess.run(
                        [sys.executable, str(ROOT / "scripts/validate.py"), "--report-file", str(path)],
                        capture_output=True, text=True, check=False,
                    )
                    self.assertEqual(result.returncode, expected_status, result.stdout + result.stderr)

    def test_rejects_missing_table(self):
        self.assertInvalid(REPORT.replace("## Findings summary", "## Findings"), "Findings summary")

    def test_rejects_empty_table_without_none_found(self):
        text = without_findings().replace("None found.\n\n## Finding details", "\n## Finding details")
        self.assertInvalid(text, "None found")

    def test_rejects_missing_summary_row(self):
        text = "\n".join(line for line in REPORT.splitlines() if not line.startswith("| F1 |"))
        self.assertInvalid(text, "summary/detail IDs")

    def test_rejects_missing_detail(self):
        start, end = REPORT.index("### F1"), REPORT.index("## Strengths")
        self.assertInvalid(REPORT[:start] + "None found.\n\n" + REPORT[end:], "summary/detail IDs")

    def test_rejects_duplicate_summary_id(self):
        row = next(line for line in REPORT.splitlines() if line.startswith("| F1 |"))
        self.assertInvalid(REPORT.replace(row, row + "\n" + row), "duplicate summary ID")

    def test_rejects_duplicate_detail_id(self):
        detail = REPORT[REPORT.index("### F1"):REPORT.index("## Strengths")]
        self.assertInvalid(REPORT.replace(detail, detail + detail), "duplicate detail ID")

    def test_rejects_severity_mismatch(self):
        self.assertInvalid(REPORT.replace("Severity: minor", "Severity: moderate"), "severity differs")

    def test_rejects_topic_mismatch(self):
        self.assertInvalid(REPORT.replace("Topic: Code Quality & Go Idioms", "Topic: Security"), "topic differs")

    def test_rejects_count_mismatch(self):
        self.assertInvalid(REPORT.replace("minor=1", "minor=2"), "do not match counts")

    def test_rejects_grade_mismatch(self):
        self.assertInvalid(REPORT.replace("— A-\n", "— C+\n", 1), "conflicts with counts")

    def test_rejects_missing_detail_fields(self):
        for field in ("Trigger", "Evidence", "Consequence", "Correction", "Verification"):
            with self.subTest(field=field):
                text = "\n".join(line for line in REPORT.splitlines() if not line.startswith(field + ":"))
                self.assertInvalid(text, field)

    def test_rejects_empty_verification(self):
        self.assertInvalid(REPORT.replace("Verification: capture logs for one failure and assert one record.", "Verification:"), "Verification")

    def test_rejects_wrong_section_order(self):
        text = REPORT.replace("## Strengths", "## temporary").replace("## Checks and limits", "## Strengths").replace("## temporary", "## Checks and limits")
        self.assertInvalid(text, "section order")

    def test_rejects_not_applicable_with_findings(self):
        self.assertInvalid(REPORT.replace("— A-\n", "— Not applicable\n", 1), "Not applicable")

    def test_rejects_incomplete_with_zero_counts(self):
        self.assertInvalid(without_findings("Insufficient evidence"), "zero counts")

    def test_rejects_incomplete_with_assessed_scope_a(self):
        text = REPORT.replace("— A-\n", "— Insufficient evidence\n", 1)
        self.assertInvalid(text.replace("Scope:", "Assessed-scope defect grade: A\nScope:", 1), "assessed-scope")

    def test_rejects_umbrella_missing_topic(self):
        text = "\n".join(line for line in umbrella().splitlines() if not line.startswith("| Security |"))
        self.assertInvalid(text, "Report cards")

    def test_rejects_umbrella_missing_profile(self):
        self.assertInvalid(umbrella().split("## Maintainability profile")[0], "Maintainability profile")

    def test_batch_rejects_old_prose_without_findings_table(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            contract_path = root / "tests/go-quality-review/shared/review-contract.json"
            contract_path.parent.mkdir(parents=True)
            contract = dict(CONTRACT, topics={SKILL: TOPIC})
            contract_path.write_text(json.dumps(contract))
            suite = root / f"tests/go-quality-review/{SKILL}/evals/evals.json"
            suite.parent.mkdir(parents=True)
            suite.write_text(json.dumps({"evals": [{"id": "logging"}]}))
            reports = root / "reports"
            output = reports / SKILL / "logging.md"
            output.parent.mkdir(parents=True)
            output.write_text(OLD_REPORT)
            errors, count = VALIDATOR.check_reports(root, reports)
            self.assertEqual(count, 1)
            self.assertTrue(errors, "a prose-only report must not pass the shared structure contract")


if __name__ == "__main__":
    unittest.main()
