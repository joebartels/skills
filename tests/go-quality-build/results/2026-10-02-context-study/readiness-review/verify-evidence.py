#!/usr/bin/env python3
"""Independent read-only reconstruction and raw-evidence identity checks."""

import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile


ROOT = Path(__file__).resolve().parents[4]
OUT = Path(__file__).resolve().parent
STUDY = ROOT / "tests/go-quality-build/results/2026-10-02-context-study"
SUITE = ROOT / "tests/go-quality-build/go-context-and-deadlines/evals"


def digest(data):
    return hashlib.sha256(data).hexdigest()


manifest = json.loads((STUDY / "manifest.json").read_text())
result = {"review_kind": "independent evidence identity checks, no author or probe launch", "input_hashes": {}, "attempts": [], "selection": []}
for relative, expected in manifest["input_hashes"].items():
    observed = digest((SUITE / relative).read_bytes())
    assert expected == observed, relative
    result["input_hashes"][relative] = observed

for row in manifest["attempts"]:
    attempt = STUDY / row["directory"]
    hashes = json.loads((attempt / "source-hashes.json").read_text())
    with tempfile.TemporaryDirectory(prefix="context-readiness-reconstruct-", dir="/private/tmp") as temp:
        work = Path(temp)
        for original in (SUITE / "files" / row["case"]).rglob("*"):
            if original.is_file():
                target = work / original.relative_to(SUITE / "files" / row["case"])
                target.parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(original, target)
        patch = attempt / "source.patch"
        patch_result = {"argv": [], "exit_code": 0, "stdout": "", "stderr": "", "meaning": "empty patch; original retained"}
        if patch.stat().st_size:
            argv = ["rtk", "proxy", "patch", "-p1", "--batch", "-i", str(patch)]
            proc = subprocess.run(argv, cwd=work, capture_output=True, text=True, timeout=15)
            patch_result = {"argv": argv, "cwd": str(work), "exit_code": proc.returncode, "stdout": proc.stdout, "stderr": proc.stderr}
            assert proc.returncode == 0, patch_result
        files = {}
        for relative, expected in hashes.items():
            reconstructed = digest((work / relative).read_bytes())
            archived = digest((attempt / "source" / relative).read_bytes())
            assert reconstructed == archived == expected, (row["number"], relative)
            files[relative] = reconstructed
        archived_inventory = {str(p.relative_to(attempt / "source")) for p in (attempt / "source").rglob("*") if p.is_file()}
        assert archived_inventory == set(hashes), (row["number"], archived_inventory, set(hashes))
        assert (work / "go.mod").read_bytes() == (SUITE / "files" / row["case"] / "go.mod").read_bytes()
        result["attempts"].append({"attempt": row["number"], "case": row["case"], "arm": row["arm"], "profile": row["profile"], "status": row["status"], "counted": row["counted"], "files": files, "patch_sha256": digest(patch.read_bytes()), "patch_application": patch_result})

selection = json.loads((STUDY / "selection-summary.json").read_text())
assert len(selection["rows"]) == 12
for row in selection["rows"]:
    folder = STUDY / Path(row["completion"]).parent
    completion = json.loads((folder / "completion.json").read_text())
    commands = completion["observed_commands"]
    body_read = any("go-context-and-deadlines/SKILL.md" in command for command in commands)
    expected_read = row["request"].startswith("applicable")
    assert completion["status"] == row["status"] == 0
    assert body_read == row["context_body_read_observed"] == expected_read
    assert row["catalog_bytes_unchanged"]
    dispatch = json.loads((folder / "dispatch.json").read_text())
    prompt = (folder / "prompt.txt").read_text()
    assert "go-context-and-deadlines" not in prompt
    result["selection"].append({"profile": row["profile"], "request": row["request"], "body_read_observed": body_read, "status": completion["status"], "observed_commands": commands, "completion_sha256": digest((folder / "completion.json").read_bytes()), "dispatch_sha256": digest((folder / "dispatch.json").read_bytes()), "events_sha256": digest((folder / "events.jsonl").read_bytes()), "prompt_sha256": digest(prompt.encode()), "declaration_sha256": digest((folder / "report.md").read_bytes()), "native_settings": [arg for arg in dispatch["argv"] if arg in ("skip_host_skill_discovery", "--ignore-user-config", "--ephemeral", "--approve-for-me")]})

mapping = json.loads((STUDY / "neutral-packet-map.json").read_text())
assert mapping["cards_frozen_before_mapping_disclosure"]
result["outcome_freezes"] = {}
result["neutral_source_identity"] = {}
case_names = {"library": "library-call-cause", "service": "service-total-budget", "control": "not-context-work", "transfer": "cli-finalization-budget"}
for packet, index_relative in mapping["freeze_indices"].items():
    index = STUDY / index_relative
    freeze = json.loads(index.read_text())
    assert freeze["frozen_before_arm_disclosure"]
    folder = STUDY / "neutral-reviews" / packet
    for relative, expected in freeze["hashes"].items():
        observed = digest((folder / relative).read_bytes())
        assert observed == expected, (packet, relative)
    result["outcome_freezes"][packet] = {"index_sha256": digest(index.read_bytes()), "artifacts_checked": len(freeze["hashes"]), "frozen_before_arm_disclosure": freeze["frozen_before_arm_disclosure"]}
    if packet == "library":
        snapshots = json.loads((folder / "source-hashes.json").read_text())
    elif packet == "service":
        snapshots = {}
        for candidate in ["candidate-1", "candidate-2"]:
            snapshots.update(json.loads((folder / candidate / "snapshot.json").read_text()))
    elif packet == "control":
        saved = json.loads((folder / "snapshot.json").read_text())
        snapshots = {f"original/{r}": h for r, h in saved["original_hashes"].items()}
        for candidate, files in saved["candidate_hashes"].items():
            snapshots.update({f"{candidate}/{r}": h for r, h in files.items()})
    else:
        saved = json.loads((folder / "source-snapshots.json").read_text())
        snapshots = {f"{candidate}/{r}": h for candidate, files in saved.items() for r, h in files.items()}
    for relative, expected in snapshots.items():
        source_id, source_relative = relative.split("/", 1)
        if source_id == "original":
            p = SUITE / "files" / case_names[packet] / source_relative
        elif source_id == "provided-contract-probes":
            p = SUITE / "controller-probes" / case_names[packet] / source_relative
        else:
            number = mapping["mapping"][packet][source_id]
            p = STUDY / f"attempt-{number:02d}" / "source" / source_relative
        observed = digest(p.read_bytes())
        assert observed == expected, (packet, relative)
    result["neutral_source_identity"][packet] = {"source_hashes_checked": len(snapshots), "candidate_attempt_mapping": mapping["mapping"][packet]}
    for candidate in ["candidate-1", "candidate-2"]:
        files = ["architecture", "code-quality", "correctness", "testing", "security", "resilience", "resources", "reproducibility", "deployment"]
        if packet == "transfer":
            files[7] = "dependencies"
        if packet == "control":
            files = ["architecture-and-design", "code-quality-and-idioms", "correctness-and-compatibility", "testing", "security", "observability-and-resilience", "performance-and-resource-management", "dependencies-and-reproducibility", "deployment-and-operations"]
        for topic in files:
            relative = f"{candidate}/{'cards/' if packet == 'control' else ''}{topic}.md"
            assert relative in freeze["hashes"] and (folder / relative).is_file(), (packet, relative)

result["final_summary_sha256"] = {name: digest((STUDY / name).read_bytes()) for name in ["comparison.md", "manifest.json", "neutral-packet-map.json", "selection-summary.json", "attempt-integrity.json"]}

result["limits"] = [
    "Completed sources reconstructed independently; failed attempts15/16 reconstruct unchanged inputs, never outcomes.",
    "Selection access cross-checked against captured observed commands; hidden full runtime context and discarded temporary native catalog cannot be independently recaptured.",
    "All72 topic cards exist in exact frozen packet inventories; card coverage, interpretation and severity remain independent reviewer judgments, not proof from a hash.",
    "No tests, authors, probes, input edits, draft edits, runtime edits or raw-archive edits performed.",
]
(OUT / "evidence-verification.json").write_text(json.dumps(result, indent=2) + "\n")
print(json.dumps({"status": "pass", "inputs": len(result["input_hashes"]), "reconstructed_attempts": len(result["attempts"]), "reconstructed_source_files": sum(len(r["files"]) for r in result["attempts"]), "checked_completed_sources": sum(r["status"] == "completed" for r in result["attempts"]), "selection_checks": len(result["selection"]), "frozen_review_artifacts": sum(r["artifacts_checked"] for r in result["outcome_freezes"].values()), "neutral_source_hashes": sum(r["source_hashes_checked"] for r in result["neutral_source_identity"].values()), "topic_cards_present": 72}))
