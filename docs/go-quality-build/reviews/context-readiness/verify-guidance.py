#!/usr/bin/env python3
"""Reviewer-only identity/example checks; no authors, probes or evidence edits."""

import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import time


ROOT = Path(__file__).resolve().parents[4]
OUT = Path(__file__).resolve().parent
STUDY = ROOT / "tests/go-quality-build/results/2026-10-02-context-study"
BASE = "0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96"
EXPECTED = "6368240e5a000cd8f056afa5980099317882d8ebcefa0e1f29921975ccc4627d"


def digest(data):
    return hashlib.sha256(data).hexdigest()


def run(argv, cwd, overrides=None):
    started = time.monotonic()
    proc = subprocess.run(
        ["rtk", "proxy", *argv], cwd=cwd,
        env={**os.environ, **(overrides or {})}, capture_output=True,
        text=True, timeout=45,
    )
    return {
        "argv": ["rtk", "proxy", *argv], "cwd": str(cwd),
        "overrides": overrides or {}, "exit_code": proc.returncode,
        "stdout": proc.stdout, "stderr": proc.stderr,
        "elapsed_seconds": round(time.monotonic() - started, 3),
    }


result = {"review_kind": "exact-guidance, independent ungraded check", "checks": []}
manifest = json.loads((STUDY / "manifest.json").read_text())
draft = ROOT / "tests/go-quality-build/go-context-and-deadlines/draft/SKILL.md"
snapshot = STUDY / "skills/go-context-and-deadlines/SKILL.md"
result["guidance_identity"] = {
    "expected_sha256": EXPECTED,
    "current_draft_sha256": digest(draft.read_bytes()),
    "frozen_snapshot_sha256": digest(snapshot.read_bytes()),
    "current_equals_snapshot": draft.read_bytes() == snapshot.read_bytes(),
    "draft_inventory": [str(p.relative_to(draft.parent)) for p in draft.parent.rglob("*") if p.is_file()],
}
assert result["guidance_identity"]["current_draft_sha256"] == EXPECTED
assert result["guidance_identity"]["frozen_snapshot_sha256"] == EXPECTED

result["catalog_identity"] = []
for relative, expected in sorted(manifest["skill_hashes"].items()):
    p = STUDY / "skills" / relative
    row = {"path": relative, "sha256": digest(p.read_bytes()), "expected_sha256": expected}
    assert row["sha256"] == expected
    if not relative.startswith("go-context-and-deadlines/"):
        runtime = ROOT / "plugins/go-quality-build/skills" / relative
        original = subprocess.run(
            ["rtk", "proxy", "git", "show", f"{BASE}:plugins/go-quality-build/skills/{relative}"],
            cwd=ROOT, capture_output=True, check=True,
        ).stdout
        row["runtime_sha256"] = digest(runtime.read_bytes())
        row["original_revision_sha256"] = digest(original)
        assert row["sha256"] == row["runtime_sha256"] == row["original_revision_sha256"]
    result["catalog_identity"].append(row)

result["review_guidance_identity"] = []
for p in sorted((ROOT / "plugins/go-quality-review/skills").rglob("*")):
    if not p.is_file() or "__pycache__" in p.parts:
        continue
    relative = str(p.relative_to(ROOT))
    original = subprocess.run(
        ["rtk", "proxy", "git", "show", f"{BASE}:{relative}"],
        cwd=ROOT, capture_output=True, check=True,
    ).stdout
    row = {"path": relative, "sha256": digest(p.read_bytes()), "original_revision_sha256": digest(original)}
    assert row["sha256"] == row["original_revision_sha256"]
    result["review_guidance_identity"].append(row)

capture_root = Path("/private/tmp/go-skills-inspiration-2026-10-02")
inventory = json.loads((ROOT / "docs/go-quality-build/context-concurrency-source-inventory.json").read_text())
result["source_capture_identity"] = []
for item in inventory:
    p = capture_root / item["repository"].replace("/", "--") / item["path"]
    row = {**item, "observed_sha256": digest(p.read_bytes()), "observed_bytes": p.stat().st_size}
    assert row["sha256"] == row["observed_sha256"] and row["bytes"] == row["observed_bytes"]
    result["source_capture_identity"].append(row)

api_root = Path("/opt/homebrew/Cellar/go/1.26.5/libexec/api")
result["api_history"] = []
for version, name in [("1.20", "Cause"), ("1.20", "WithCancelCause"), ("1.21", "WithoutCancel"), ("1.21", "AfterFunc")]:
    p = api_root / f"go{version}.txt"
    lines = [line for line in p.read_text().splitlines() if line.startswith(f"pkg context, func {name}(")]
    assert len(lines) == 1
    result["api_history"].append({"version": version, "helper": name, "record": lines[0], "record_file_sha256": digest(p.read_bytes())})

result["primary_semantic_source_identity"] = []
for relative in ["src/context/context.go", "src/net/http/request.go"]:
    p = api_root.parent / relative
    result["primary_semantic_source_identity"].append({"distribution": "official Go1.26.5", "path": relative, "sha256": digest(p.read_bytes())})

example = re.search(r"```go\n(.*?)\n```", draft.read_text(), re.S).group(1)
check_source = STUDY / "guidance-preflight-checks/main.go"
source = check_source.read_text()
assert example in source
result["semantic_source"] = {
    "origin": str(check_source.relative_to(ROOT)), "sha256": digest(check_source.read_bytes()),
    "exact_finalization_example_present": True,
    "scope": "Reuse archived reviewer check in a fresh disposable module; exact draft function verified as substring. Counterexamples are reviewer demonstrations, never authors or selection probes.",
}
with tempfile.TemporaryDirectory(prefix="context-readiness-guidance-", dir="/private/tmp") as temp:
    work = Path(temp)
    (work / "go.mod").write_text("module example.com/contextreadiness\n\ngo 1.22\n")
    (work / "main.go").write_text(source)
    overrides = {"GOWORK": "off", "GOTOOLCHAIN": "local", "GOCACHE": "/private/tmp/context-readiness-guidance-cache"}
    for go in ["/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go", "/opt/homebrew/Cellar/go/1.26.5/libexec/bin/go"]:
        for args in [["version"], ["run", "."]]:
            check = run([go, *args], work, overrides)
            result["checks"].append(check)
            assert check["exit_code"] == 0, check

check = run(["python3", "/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py", str(draft.parent)], ROOT)
result["checks"].append(check)
assert check["exit_code"] == 0
result["limits"] = [
    "No author launch, native selection probe, guidance edit, rubric edit or raw-outcome edit.",
    "Fresh example/semantic checks cover only their named boundaries on darwin/arm64 Go1.22.12 and Go1.26.5.",
    "Guidance validation cannot establish measured benefit, automatic package installation, other-harness implementation quality or broad reuse readiness.",
]
(OUT / "guidance-verification.json").write_text(json.dumps(result, indent=2) + "\n")
print(json.dumps({"draft_sha256": EXPECTED, "catalog_files": len(result["catalog_identity"]), "unchanged_review_files": len(result["review_guidance_identity"]), "verified_source_captures": len(result["source_capture_identity"]), "semantic_and_validator_checks": len(result["checks"]), "status": "pass"}, indent=2))
