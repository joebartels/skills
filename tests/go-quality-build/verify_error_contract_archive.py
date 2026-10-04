"""Verify the sealed error-contract archive without its original host or checkout."""

import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile



def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def inventory(folder):
    return {
        str(path.relative_to(folder)): digest(path)
        for path in sorted(folder.rglob("*")) if path.is_file()
    }


def verify(run: Path):
    if not __debug__:
        raise RuntimeError("archive verification requires Python assertions; disable -O/PYTHONOPTIMIZE")
    seal = json.loads((run / "seal.json").read_text())
    for relative, expected in seal["file_sha256"].items():
        assert digest(run / relative) == expected, relative
    manifest = json.loads((run / "manifest.json").read_text())
    assert inventory(run / "guidance") == manifest["guidance_sha256"]
    assert inventory(run / "prior") == manifest["prior_draft_sha256"]
    assert inventory(run / "probes") == manifest["probe_sha256"]

    completed = []
    skipped = []
    for record in manifest["cases"]:
        sample = run / "samples" / record["id"]
        assert inventory(sample / "original") == record["input_sha256"]
        assert digest(sample / "catalog.json") == record["catalog_sha256"]
        catalog = json.loads((sample / "catalog.json").read_text())
        assert "go-error-contracts" not in json.dumps(catalog)
        if record["status"] == "skipped-no-baseline-failure":
            assert record["arm"] == "guided"
            assert not (sample / "candidate").exists()
            skipped.append(record["id"])
            continue
        assert record["status"] == "archived", record
        assert record["arm"] == "baseline"
        assert inventory(sample / "candidate") == record["source_sha256"]
        assert digest(sample / "source.patch") == record["patch_sha256"]
        with tempfile.TemporaryDirectory(prefix="error-contract-integrity-") as directory:
            module = Path(directory) / "module"
            shutil.copytree(sample / "original", module)
            result = subprocess.run(["git", "apply", str(sample / "source.patch")],
                                    cwd=module, text=True, capture_output=True, timeout=60)
            assert result.returncode == 0, result.stderr
            assert inventory(module) == record["source_sha256"]
        checks = json.loads((sample / "verification.json").read_text())
        for check in checks:
            if check["scope"] == "actual-cli" and check["name"] == "negative":
                # Original contract permits invalid usage to return 2.
                assert check["exit_code"] == 2
                assert check["output_bytes"] is None
            else:
                assert check["exit_code"] == 0, check
                if check["command"][0] == "gofmt":
                    assert check["output"] == ""
                if check["scope"] == "actual-cli":
                    assert check["output_bytes"] == check["expected_output"]
        report = json.loads((sample / "author-report.json").read_text())
        assert "go-error-contracts" not in report["selected_skills"]
        for opened in report["opened_guidance"]:
            path = Path(opened)
            if "guidance" in path.parts:
                relative = Path(*path.parts[path.parts.index("guidance") + 1:])
                assert digest(run / "guidance" / relative) == manifest["guidance_sha256"][str(relative)]
            elif str(path) != record["catalog_path"]:
                assert str(path) in manifest["reported_external_guidance_sha256"], opened
        completed.append(record["id"])

    review_map = json.loads((run / "review-map.json").read_text())
    for review in review_map["cases"]:
        archived = run / "samples" / review["sample"]
        assert inventory(archived / "candidate") == review["source_sha256"]
    assert len(review_map["cases"]) == len(completed) == 8
    assert len(skipped) == 8
    review_ids = {review["review_id"]: review["sample"] for review in review_map["cases"]}
    sample_records = {record["id"]: record for record in manifest["cases"]}
    review_files = run / "blind-review"
    for prefix in ["", "supplement-"]:
        checks_file = review_files / (prefix + "checks.json")
        checks = json.loads(checks_file.read_text())
        summary = json.loads((review_files / (prefix + "summary.json")).read_text())
        assert summary["checks_sha256"] == digest(checks_file)
        assert checks["sources_unchanged"]
        assert checks["source_hashes_before"] == checks["source_hashes_after"]
        assert summary["findings"] == []
        for relative, expected in checks["source_hashes_before"].items():
            path = Path(relative)
            sample_id = review_ids[path.parts[0]]
            sample = run / "samples" / sample_id
            if path.parts[1] in {"base", "module"}:
                folder = "original" if path.parts[1] == "base" else "candidate"
                preserved = sample / folder / Path(*path.parts[2:])
            elif path.parts[1] == "task.txt":
                preserved = sample / "prompt.txt"
            else:
                assert path.parts[1] == "contract-probes.go.txt"
                preserved = run / "probes" / (sample_records[sample_id]["case"] + "_test.go.txt")
            assert digest(preserved) == expected, relative
    evidence = json.loads((review_files / "evidence/manifest.json").read_text())
    for record in evidence["files"]:
        assert digest(review_files / record["path"]) == record["sha256"]
    return {
        "status": "pass",
        "completed_baseline_samples": completed,
        "skipped_guided_samples": skipped,
        "checks": ["original archive seal", "archived input and prior draft hashes",
                   "archived guidance and probe hashes", "catalogs without error guidance",
                   "candidate and patch hashes", "exact patch reconstructions",
                   "recorded held tests and candidate checks", "adjudicated CLI statuses and output",
                   "review inputs mapped to archived copies", "review inputs/outputs and preserved probes"],
        "limits": ["Machine-local external guidance and temporary paths are historical attestations; unarchived contents are not verified",
                   "Current runtime guidance and live fixtures are outside this archived check",
                   "Authors report opened guidance; full tool traces are not archived",
                   "This integrity check does not rerun already recorded Go tests"]
    }


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("archive", nargs="?", type=Path,
                        default=Path(__file__).parent / "results/2026-10-03-error-contracts-r2")
    args = parser.parse_args()
    print(json.dumps(verify(args.archive.resolve())))
