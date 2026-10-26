"""Exercise the installed review record → grade → Markdown boundary."""

import importlib.util
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPTS = ROOT / "plugins/go-quality-review/skills/go-quality-report/scripts"
CQ = "go-code-quality-and-idioms"
ARCH = "go-architecture-and-design"
CONTRACT = json.loads((ROOT / "tests/go-quality-review/shared/review-contract.json").read_text())
spec = importlib.util.spec_from_file_location("validator", ROOT / "scripts/validate.py")
VALIDATOR = importlib.util.module_from_spec(spec)
spec.loader.exec_module(VALIDATOR)


def evidence(text="The same installation mapping appears in three handlers."):
    return [{"kind": "inspected", "reference": "handler.go:42", "text": text}]


def finding(fid="F1", severity="minor", topic=CQ):
    return {
        "id": fid, "title": "Repeated cookie installation", "cause": "Three implementations of one mapping",
        "status": "existing-in-scope", "trigger": "Audit how login and refresh install the cookie suite.",
        "evidence": evidence(), "analysis": "All three callers install the same result; names and attributes already have one owner.",
        "assessments": [{"topic": topic, "severity": severity, "consequence": "The same role-to-value mapping needs three inspections."}],
        "owner": "transport", "correction": "Extract the suite setter, preserving response handling.",
        "verification": "Exercise login and both refresh routes; assert cookie values and attributes.",
        "sources": ["transport/" + fid],
    }


def record():
    return {
        "version": 1, "report": CQ, "mode": "code-area", "scope": "service at abc123; production transport and tests",
        "rationale": "One localized repeated implementation.",
        "topics": [{"id": CQ, "rationale": "One minor mapping cost.", "coverage": [
            {"id": "transport", "status": "complete", "reason": "All three handlers and consumers inspected."}], "cards": []}],
        "findings": [finding()], "decisions": [],
        "strengths": [{"id": "G1", "topics": [CQ], "description": "Cookie attributes have one owner.", "evidence": evidence()}],
        "safeguards": [], "checks": [], "limits": ["No service tests executed."],
    }


def overall():
    data = record()
    data["report"] = "overall"
    for topic in CONTRACT["topics"]:
        if topic != CQ:
            data["topics"].append({"id": topic, "rationale": "Outside this review's applicable decisions.", "cards": [],
                "coverage": [{"id": "scope", "status": "not_applicable", "reason": "No applicable decision in this fixture."}]})
    data["maintainability"] = {
        "rows": [{"dimension": name, "rating": "Not demonstrated", "confidence": "Low", "evidence": "Not assessed in this fixture.",
                  "limits": "Requires representative source traces."} for name in (
            "Clarity", "Change locality", "Cohesion and boundaries", "Abstraction cost", "Rule ownership", "Testability", "Complexity hotspots")],
        "cleanliness": "Only transport mapping was assessed.", "structure": "No broader claim.", "extensibility": "No extension experiment.",
    }
    return data


class RenderTests(unittest.TestCase):
    def run_cli(self, data, *args, script=None):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "record.json"
            path.write_text(json.dumps(data))
            return subprocess.run([sys.executable, str(script or SCRIPTS / "render.py"), str(path), *args],
                                  cwd=tmp, text=True, capture_output=True, env={**os.environ, "PYTHONDONTWRITEBYTECODE": "1"})

    def render(self, data, *args):
        result = self.run_cli(data, *args)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(VALIDATOR.check_report(result.stdout, CONTRACT), [])
        return result.stdout

    def test_topic_grade_and_one_record_feed_both_table_and_details(self):
        text = self.render(record())
        self.assertIn("# Code Quality & Go Idioms — A-", text)
        self.assertIn("critical=0, major=0, moderate=0, minor=1", text)
        self.assertEqual(text.count("| F1 |"), 1)
        self.assertEqual(text.count("### F1 —"), 1)
        self.assertIn("Verification: Proposed:", text)

    def test_shared_cause_keeps_topic_severity_without_double_counting(self):
        data = overall()
        arch = next(t for t in data["topics"] if t["id"] == ARCH)
        arch["coverage"][0]["status"] = "complete"
        data["findings"][0]["assessments"].append({"topic": ARCH, "severity": "moderate", "consequence": "Substantial boundary coordination."})
        text = self.render(data)
        self.assertIn("# Go Quality Report — B", text)
        self.assertIn("moderate=1, minor=0", text)
        self.assertIn("| Architecture & Design | B |", text)
        self.assertIn("| Code Quality & Go Idioms | A- |", text)
        self.assertIn("Architecture & Design (moderate)", text)
        self.assertIn("# Architecture & Design — B", self.render(data, "--topic", ARCH))

    def test_unresolved_decision_prevents_complete_grade(self):
        data = record()
        data["decisions"] = [{"id": "D1", "title": "Unknown caller policy", "disposition": "unresolved", "topics": [CQ],
                              "reason": "Missing caller could change admission.", "evidence": evidence()}]
        text = self.render(data)
        self.assertIn("— Insufficient evidence", text)
        self.assertIn("Assessed-scope defect grade: A-", text)
        self.assertIn("Unknown caller policy", text)
        data["findings"] = []
        text = self.render(data)
        self.assertNotIn("Finding counts:", text)
        self.assertNotIn("Assessed-scope defect grade:", text)

    def test_rejected_and_merged_decisions_do_not_change_counts(self):
        data = record()
        data["decisions"] = [
            {"id": "D1", "title": "Rename cookie", "disposition": "rejected", "topics": [CQ], "reason": "Name already centralized.", "evidence": evidence()},
            {"id": "D2", "title": "Third setter", "disposition": "merged", "topics": [CQ], "reason": "Same correction.", "evidence": evidence(), "finding_id": "F1"}]
        text = self.render(data)
        self.assertIn("minor=1", text)
        self.assertIn("Ungraded decisions", text)
        self.assertIn("F1", text.split("Ungraded decisions")[1])

    def test_related_finding_outside_solo_topic_is_visible_but_ungraded(self):
        data = record()
        data["decisions"] = [{"id": "D1", "title": "Unused API input", "disposition": "related", "topics": [ARCH],
                              "reason": "Confirmed boundary cost belongs to Architecture, outside this topic review.", "evidence": evidence()}]
        text = self.render(data)
        self.assertIn("— A-\n", text)
        self.assertIn("Unused API input", text)
        self.assertIn("Architecture & Design", text.split("Ungraded decisions")[1])

    def test_critical_and_systemic_major_reach_survive_projection(self):
        for level, systemic, expected in [("major", True, "C-"), ("critical", False, "F")]:
            data = record()
            data["findings"][0]["assessments"][0].update(severity=level, systemic=systemic, reach="Distinct affected workflows traced.")
            text = self.render(data)
            self.assertIn("— " + expected + "\n", text)
            self.assertIn("Distinct affected workflows traced.", text)

    def test_documented_record_example_runs(self):
        guide = (SCRIPTS.parent / "references/review-record.md").read_text()
        data = json.loads(re.search(r"```json\n(.*?)\n```", guide, re.S)[1])
        self.assertIn("— A-\n", self.render(data))

    def test_successful_file_output_matches_stdout(self):
        with tempfile.TemporaryDirectory() as tmp:
            output = Path(tmp) / "report.md"
            result = self.run_cli(record(), "--output", str(output))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout, "")
            self.assertEqual(output.read_text(), self.render(record()))

    def test_zero_findings_requires_strength_and_safeguards_for_a_plus(self):
        data = record()
        data["findings"] = []
        self.assertIn("— A\n", self.render(data))
        data["safeguards"] = [{"id": "S1", "topics": [CQ], "risk": "Cancellation loses security response", "evidence": evidence(), "nonroutine": True},
                              {"id": "S2", "topics": [CQ], "risk": "Key outage disables independent flow", "evidence": evidence(), "nonroutine": True}]
        self.assertIn("— A+\n", self.render(data))
        data["strengths"] = []
        self.assertIn("— Insufficient evidence", self.render(data))

    def test_not_applicable_and_partial_states_preserve_structure(self):
        for state, title in [("not_applicable", "Not applicable"), ("partial", "Insufficient evidence")]:
            data = record()
            data["findings"] = []
            data["strengths"] = []
            data["topics"][0]["coverage"][0]["status"] = state
            text = self.render(data)
            self.assertIn("— " + title, text)
            self.assertNotIn("Finding counts:", text)

    def test_natural_id_order_and_markdown_escaping(self):
        data = record()
        data["findings"] = [finding("F10"), finding("F2"), finding("F3", "moderate")]
        data["findings"][0]["title"] = "A | B\nnot a heading"
        data["findings"][0]["evidence"][0]["text"] = "`a|b`\nsecond line"
        text = self.render(data)
        self.assertLess(text.index("| F3 |"), text.index("| F2 |"))
        self.assertLess(text.index("| F2 |"), text.index("| F10 |"))
        self.assertIn(r"A \| B", text)

    def test_render_is_deterministic(self):
        data = overall()
        self.assertEqual(self.render(data), self.render(data))

    def test_invalid_records_fail_without_tracebacks_or_partial_report(self):
        variants = []
        for key, value in [("version", 99), ("version", True), ("report", "unknown"), ("grade", "A+")]:
            data = record(); data[key] = value; variants.append(data)
        data = record(); data["findings"] *= 2; variants.append(data)
        data = record(); data["findings"][0]["evidence"] = []; variants.append(data)
        data = record(); data["findings"][0]["assessments"][0]["severity"] = "low"; variants.append(data)
        data = record(); data["findings"][0]["assessments"][0]["topic"] = ARCH; variants.append(data)
        data = record(); data["findings"][0]["assessments"][0]["systemic"] = True; variants.append(data)
        data = record(); data["findings"][0]["assessments"][0]["severity"] = "critical"; variants.append(data)
        data = record(); data["mode"] = "changeset"; variants.append(data)
        data = record(); data["topics"][0]["coverage"][0]["status"] = "not_applicable"; variants.append(data)
        data = overall(); data["topics"].pop(); variants.append(data)
        data = overall(); data["maintainability"]["rows"].pop(); variants.append(data)
        data = record(); data["decisions"] = [{"id": "D1", "title": "Merge", "disposition": "merged", "reason": "Same cause", "topics": [CQ], "evidence": evidence(), "finding_id": "F99"}]; variants.append(data)
        for index, data in enumerate(variants):
            with self.subTest(index=index):
                result = self.run_cli(data)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(result.stdout, "")
                self.assertIn("error:", result.stderr)
                self.assertNotIn("Traceback", result.stderr)

    def test_invalid_input_does_not_clobber_output_or_source(self):
        with tempfile.TemporaryDirectory() as tmp:
            source, output = Path(tmp) / "input.json", Path(tmp) / "report.md"
            output.write_text("keep")
            for raw in ['{"version":1,"version":2}', '{"bad":']:
                source.write_text(raw)
                result = subprocess.run([sys.executable, str(SCRIPTS / "render.py"), str(source), "--output", str(output)], capture_output=True, text=True)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(output.read_text(), "keep")
                self.assertNotIn("Traceback", result.stderr)
            source.write_text(json.dumps(record()))
            result = subprocess.run([sys.executable, str(SCRIPTS / "render.py"), str(source), "--output", str(source)], capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(json.loads(source.read_text()), record())

    def test_runtime_works_without_repository_or_extra_packages(self):
        with tempfile.TemporaryDirectory() as tmp:
            installed = Path(tmp) / "installed"
            shutil.copytree(SCRIPTS, installed)
            result = self.run_cli(overall(), script=installed / "render.py")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(VALIDATOR.check_report(result.stdout, CONTRACT), [])


if __name__ == "__main__":
    unittest.main()
