#!/usr/bin/env python3
"""Read-only portable evidence gate. Requires Python3 and Git; no Go or model launch.

Original capture commands/absolute paths are historical provenance. This gate
reconstructs bytes in platform temporary storage and verifies externally anchored
seals, without rewriting captures or asserting their old checks passed anew.
"""
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

if sys.flags.optimize:
    raise SystemExit(
        "FAIL: Python optimization disables integrity checks; run without -O/-OO "
        "and unset PYTHONOPTIMIZE."
    )

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[3]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def tree(path, exclude=None):
    return {p.relative_to(path).as_posix(): sha(p) for p in path.rglob("*")
            if p.is_file() and p != exclude}


def safe_child(root, name):
    path = root / name
    assert path.resolve().is_relative_to(root.resolve()), name
    return path


def seal(root, expected, continuation, separately_sealed_subtree=None):
    index = root / "checksums.json"
    assert sha(index) == expected, (root.name, "external index anchor")
    assert expected in continuation, (root.name, "canonical anchor missing")
    actual = tree(root, index)
    if separately_sealed_subtree is not None:
        prefix = separately_sealed_subtree + "/"
        actual = {name: value for name, value in actual.items() if not name.startswith(prefix)}
    data = read(index)
    expected_files = data.get("sha256")
    if expected_files is None:
        expected_files = {row["path"]: row["sha256"] for row in data["artifacts"]}
    assert actual == expected_files, (root.name, "sealed bytes changed")


def verify():
    m = read(OUT / "manifest.json")
    continuation = (ROOT / "docs/go-quality-build/README.md").read_text()
    assert m["runtime_version"] == "0.2.0" and m["runtime_skill_count"] == 5
    assert m["author_attempts"] == 31 and m["selection_attempts"] == 26
    assert m["integrated_author_attempts"] == 0 and m["build_cases"] == 46
    assert m["promotion"] == "neither"
    package = ROOT / "plugins/go-quality-build"
    assert tree(package / "skills") == m["runtime_skill_sha256"]
    assert len(list((package / "skills").glob("*/SKILL.md"))) == 5
    for name in ("plugin.json", ".claude-plugin/plugin.json"):
        assert read(package / name)["version"] == "0.2.0"
    for name, expected in m["unchanged_package_review_sha256"].items():
        assert sha(safe_child(ROOT, name)) == expected, name
    for record in m["drafts"]:
        assert sha(ROOT / record["path"]) == record["sha256"]
        assert not (package / "skills" / record["name"]).exists()

    counts = {"attempts": 0, "completed": 0, "failed": 0, "source_files": 0}
    for record in m["archives"]:
        archive = safe_child(ROOT, record["path"])
        seal(archive, record["checksum_index_sha256"], continuation)
        archived = read(archive / "manifest.json")
        evals = ROOT / "tests/go-quality-build" / archived["candidate"] / "evals"
        for name, expected in archived["input_hashes"].items():
            assert sha(safe_child(evals, name)) == expected, name
        skills = archive / record["skill_catalog"]
        assert tree(skills) == archived["skill_hashes"], archive.name
        if "draft_revision_sha256" in archived:
            assert archived["disposition"] == "reviewed-draft-not-promoted"
            assert sha(skills / archived["candidate"] / "SKILL.md") == archived["draft_revision_sha256"]
        for attempt in archived["attempts"]:
            source = safe_child(archive, attempt["directory"])
            hashes = read(source / "source-hashes.json")
            assert tree(source / "source") == hashes, source
            with tempfile.TemporaryDirectory(prefix="go-delivery-") as temp:
                work = Path(temp) / "source"
                shutil.copytree(evals / "files" / attempt["case"], work)
                patch = source / "source.patch"
                if patch.read_bytes():
                    result = subprocess.run(["git", "apply", "--", str(patch)], cwd=work,
                                            capture_output=True, text=True)
                    assert result.returncode == 0, (source, result.stderr)
                assert tree(work) == hashes, (source, "reconstruction")
            completion = read(source / "completion.json")
            assert completion["exit_code"] == attempt["exit_code"]
            if attempt["status"] == "failed":
                assert not patch.read_bytes() and not (source / "events.jsonl").read_bytes()
            else:
                assert attempt["status"] == "completed" and attempt["exit_code"] == 0
                assert read(source / "verification.json")["reconstruction_hashes_match"]
            counts["attempts"] += 1
            counts[attempt["status"]] += 1
            counts["source_files"] += len(hashes)
        for freeze in (archive / "neutral-reviews").glob("*-freeze.json"):
            data = read(freeze)
            assert data["frozen_before_arm_disclosure"]
            group = freeze.name.removesuffix("-freeze.json")
            for name, expected in data["hashes"].items():
                assert sha(safe_child(archive / "neutral-reviews" / group, name)) == expected
        mapping = archived.get("neutral_review_mapping", {})
        for packet in mapping.values() if isinstance(mapping, dict) else ():
            if isinstance(packet, dict) and "frozen_card_hashes" in packet:
                reviews = archive / f"attempt-{packet['attempt']:02}" / "reviews"
                assert packet["cards_pending"] is False
                for name, expected in packet["frozen_card_hashes"].items():
                    assert sha(safe_child(reviews, name)) == expected
    assert counts == m["reconstruction_counts"], counts
    for record in m["additional_seals"]:
        seal(safe_child(ROOT, record["path"]), record["checksum_index_sha256"], continuation,
             record.get("separately_sealed_subtree"))
    exceptions = read(OUT / "raw-whitespace-exceptions.json")["entries"]
    assert len(exceptions) == m["whitespace_exception_count"]
    assert len({entry["path"] for entry in exceptions}) == len(exceptions)
    for entry in exceptions:
        assert sha(safe_child(ROOT, entry["path"])) == entry["sha256"]
    checks = read(OUT / "shared-verification.json")
    assert all(row["status"] == 0 for row in checks) and len(checks) == 14
    eligibility = ROOT / "tests/go-quality-build/results/2026-10-02-context-concurrency-combined"
    assert read(eligibility / "manifest.json")["status"] == "skipped-ineligible"
    assert not (ROOT / "tests/go-quality-build/context-concurrency-combined/evals").exists()
    expected = sha(OUT / "checksums.json")
    seal(OUT, expected, continuation)
    print("PASS: portable31 reconstructions (26 completed/5 failed), five runtime0.2.0 skills, anchored seals and preserved counterevidence")


if __name__ == "__main__":
    verify()
