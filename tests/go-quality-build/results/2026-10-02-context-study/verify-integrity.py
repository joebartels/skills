"""Read-only closeout gate: source reconstruction, frozen identities and declared contrary outcomes."""
from pathlib import Path
import hashlib, json, shutil, subprocess, tempfile

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[3]
EVALS = ROOT / "tests/go-quality-build/go-context-and-deadlines/evals"
manifest = json.loads((OUT / "manifest.json").read_text())

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

assert manifest.get("disposition") == "reviewed-draft-not-promoted", "Disposition pending"
assert manifest["authors_spent_total"] == 16
assert len(manifest["attempts"]) == 10 and all(a["counted"] for a in manifest["attempts"])
assert manifest["selection_probes_spent_total"] == 14
assert sha(ROOT / "tests/go-quality-build/go-context-and-deadlines/draft/SKILL.md") == manifest["draft_revision_sha256"]
for name, expected in manifest["skill_hashes"].items():
    assert sha(OUT / "skills" / name) == expected, name
for name, expected in manifest["input_hashes"].items():
    assert sha(EVALS / name) == expected, name
for row in manifest["attempts"]:
    archive = OUT / row["directory"]
    work = Path(tempfile.mkdtemp(prefix="go-integrity-", dir="/private/tmp"))
    shutil.copytree(EVALS / "files" / row["case"], work, dirs_exist_ok=True)
    patch = archive / "source.patch"
    if patch.read_text():
        result = subprocess.run(["rtk", "proxy", "git", "apply", str(patch)], cwd=work, capture_output=True, text=True)
        assert result.returncode == 0, result.stderr
    actual = {str(p.relative_to(work)): sha(p) for p in work.rglob("*") if p.is_file()}
    assert actual == json.loads((archive / "source-hashes.json").read_text()), row["number"]
    if row["status"] == "failed":
        assert not patch.read_text() and not (archive / "events.jsonl").read_text()
    else:
        evidence = json.loads((archive / "verification.json").read_text())
        assert evidence["reconstruction_hashes_match"]
        expected = [0] * 8
        held = [0, 0] if row["case"] == "not-context-work" else [0, 0, 0]
        if row["number"] == 14:
            expected[-1] = 1
            held[1] = 1
        assert [c["status"] for c in evidence["ordinary_checks"]] == expected
        assert not evidence["ordinary_checks"][6]["stdout"].strip(), "Unformatted source"
        assert [c["status"] for c in evidence["held_contract_checks"]] == held
        assert len(evidence["frozen_mutation_sensitivity"]) == (2 if row["case"] == "library-call-cause" else 1)
        assert all(m["qualified"] for m in evidence["frozen_mutation_sensitivity"])
    shutil.rmtree(work)
for group in ("library", "service", "control", "transfer"):
    freeze = json.loads((OUT / "neutral-reviews" / (group + "-freeze.json")).read_text())
    assert freeze["frozen_before_arm_disclosure"]
    for name, expected in freeze["hashes"].items():
        assert sha(OUT / "neutral-reviews" / group / name) == expected, (group, name)
selection = json.loads((OUT / "selection-summary.json").read_text())
assert len(selection["rows"]) == 12
assert all(r["status"] == 0 and r["context_body_read_observed"] == r["request"].startswith("applicable") for r in selection["rows"])
assert sha(OUT / manifest["readiness_review"]["artifact"]) == manifest["readiness_review"]["sha256"]
seal = json.loads((OUT / "checksums.json").read_text())
actual = {str(p.relative_to(OUT)): sha(p) for p in OUT.rglob("*") if p.is_file() and p != OUT / "checksums.json"}
assert actual == seal["sha256"], "Archive seal mismatch"
print("PASS: 16 counted attempts, eight reconstructed checked outcomes, declared minimum failure, frozen cards, 14 selection attempts and archive seal")
