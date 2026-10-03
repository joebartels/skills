#!/usr/bin/env python3
"""Read-only eligibility check; does not launch authors or rewrite evidence."""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
ARCHIVE = Path(__file__).resolve().parent
m = json.loads((ARCHIVE / "manifest.json").read_text())
assert m["status"] == "skipped-ineligible"
assert m["integrated_author_attempts"] == 0
assert m["group_author_attempts"] == 31
assert m["expected_build_cases"] == 46
assert m["interaction_fixture_created"] is False
assert not (ROOT / "tests/go-quality-build/context-concurrency-combined/evals").exists()
for record in m["individual_dispositions"]:
    path = ROOT / record["manifest"]
    assert hashlib.sha256(path.read_bytes()).hexdigest() == record["sha256"]
    individual = json.loads(path.read_text())
    assert individual["disposition"] == record["disposition"] == "reviewed-draft-not-promoted"
    assert individual["draft_revision_sha256"] == record["draft_revision_sha256"]
print("PASS: interaction ineligible; zero integrated launches; both individual decisions exact")
