"""Read-only closeout gate; checks identities, reconstructions and declared counterevidence."""
from pathlib import Path
import hashlib, json, shutil, subprocess, tempfile

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[3]
EVALS = ROOT / "tests/go-quality-build/go-concurrency-and-ownership/evals"
m = json.loads((OUT / "manifest.json").read_text())

def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()

assert m.get("disposition") == "reviewed-draft-not-promoted", "Independent disposition pending"
assert m["authors_spent_total"] == 15 and m["group_authors_spent_total"] == 31
assert len(m["attempts"]) == 12 and all(r["counted"] and r["status"] == "completed" for r in m["attempts"])
assert m["selection_probes_spent_total"] == 12 and m["group_selection_probes_spent_total"] == 26
assert sha(ROOT / "tests/go-quality-build/go-concurrency-and-ownership/draft/SKILL.md") == m["skill_hashes"]["go-concurrency-and-ownership/SKILL.md"]
for relative, expected in m["input_hashes"].items():
    assert sha(EVALS / relative) == expected, relative
for relative, expected in m["skill_hashes"].items():
    assert sha(OUT / "skills" / relative) == expected, relative
for row in m["attempts"]:
    archive = OUT / row["directory"]
    with tempfile.TemporaryDirectory(prefix="go-integrity-") as temporary:
        work = Path(temporary)
        shutil.copytree(EVALS / "files" / row["case"], work, dirs_exist_ok=True)
        p = subprocess.run(["rtk", "proxy", "git", "apply", str(archive / "source.patch")], cwd=work, capture_output=True, text=True)
        assert p.returncode == 0, p.stderr
        hashes = {str(p.relative_to(work)): sha(p) for p in work.rglob("*") if p.is_file()}
        assert hashes == json.loads((archive / "source-hashes.json").read_text()), row["number"]
        assert hashes == {str(p.relative_to(archive / "source")): sha(p) for p in (archive / "source").rglob("*") if p.is_file()}
        assert (work / "go.mod").read_bytes() == (EVALS / "files" / row["case"] / "go.mod").read_bytes()
    dispatch = json.loads((archive / "dispatch.txt").read_text())
    assert dispatch["environment_overrides"]["GOTOOLCHAIN"] == "local"
    assert dispatch["host_disable_sha256"] == m["host_disable_sha256"]
    assert "--approve-for-me" in dispatch["argv"] and "--sandbox" not in dispatch["argv"]
    assert ("go-concurrency-and-ownership" in (archive / "prompt.txt").read_text()) == (row["arm"] == "exposure")
    evidence = json.loads((archive / "verification.json").read_text())
    assert evidence["reconstruction_hashes_match"]
    for name, expected in m["declared_check_statuses"][str(row["number"])].items():
        assert [c["status"] for c in evidence[name]] == expected, (row["number"], name)
    assert not evidence["ordinary_checks"][6]["stdout"].strip()
    assert (archive / "mutations.json").is_file()
mapping = json.loads((OUT / "neutral-packet-map.json").read_text())
assert mapping["cards_frozen_before_mapping_disclosure"]
cards = 0
for group, labels in mapping["mapping"].items():
    freeze = json.loads((OUT / "neutral-reviews" / (group + "-freeze.json")).read_text())
    assert freeze["frozen_before_arm_disclosure"]
    folder = OUT / "neutral-reviews" / group
    for name, expected in freeze["hashes"].items():
        assert sha(folder / name) == expected, (group, name)
    for label, n in labels.items():
        observed = {str(p.relative_to(folder / "candidates" / label)): sha(p) for p in (folder / "candidates" / label).rglob("*") if p.is_file()}
        assert observed == json.loads((OUT / f"attempt-{n:02d}/source-hashes.json").read_text())
    cards += len([p for p in folder.rglob("*.md") if p.parent.name in labels and p.parent.parent.name == "cards"])
assert cards == 108, cards
selection = json.loads((OUT / "selection-summary.json").read_text())
assert len(selection["rows"]) == 12
for r in selection["rows"]:
    c = json.loads((OUT / r["completion"]).read_text())
    read = any("go-concurrency-and-ownership/SKILL.md" in command for command in c["observed_commands"])
    assert c["status"] == r["status"] == 0
    assert read == r["candidate_body_read_observed"]
    assert r["matches_intended_selection"] == (read == r["request"].startswith("applicable"))
    assert "go-concurrency-and-ownership" not in (OUT / Path(r["completion"]).parent / "prompt.txt").read_text()
assert sha(OUT / m["readiness_review"]["artifact"]) == m["readiness_review"]["sha256"]
assert sum(r["matches_intended_selection"] for r in selection["rows"]) == 10
assert all(c["status"] == 0 for c in json.loads((OUT / "shared-verification.json").read_text()))
seal = json.loads((OUT / "checksums.json").read_text())
actual = {str(p.relative_to(OUT)): sha(p) for p in OUT.rglob("*") if p.is_file() and p != OUT / "checksums.json"}
assert actual == seal["sha256"], "Archive seal mismatch"
print("PASS: 15 counted authors, twelve exact reconstructed outcomes, declared failures, 108 frozen cards, twelve native selection probes and archive seal")
