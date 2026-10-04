"""Check archived inputs, guidance, patches, reviews and recorded checks."""

import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile


RUN = Path(__file__).resolve().parent
REPO = RUN.parents[3]


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def inventory(folder):
    return {
        str(path.relative_to(folder)): digest(path)
        for path in sorted(folder.rglob("*")) if path.is_file()
    }


def verify():
    manifest = json.loads((RUN / "manifest.json").read_text())
    assert inventory(RUN / "guidance") == manifest["guidance_sha256"]
    assert inventory(REPO / "plugins/go-quality-build/skills") == manifest["guidance_sha256"]
    assert inventory(RUN / "prior") == manifest["prior_draft_sha256"]
    assert digest(REPO / "tests/go-quality-build/go-error-contracts/draft/SKILL.md") == manifest["prior_draft_sha256"]["SKILL.md"]
    assert inventory(RUN / "probes") == manifest["probe_sha256"]
    for path, expected in manifest["reported_external_guidance_sha256"].items():
        assert digest(Path(path)) == expected, path

    completed = []
    skipped = []
    for record in manifest["cases"]:
        sample = RUN / "samples" / record["id"]
        assert inventory(sample / "original") == record["input_sha256"]
        assert inventory(REPO / "tests/go-quality-build/go-error-contracts/evals/files" / record["case"]) == record["input_sha256"]
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
        with tempfile.TemporaryDirectory(prefix="error-contract-integrity-", dir="/private/tmp") as directory:
            module = Path(directory) / "module"
            shutil.copytree(sample / "original", module)
            result = subprocess.run(["rtk", "proxy", "git", "apply", str(sample / "source.patch")],
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
                assert digest(path) == manifest["guidance_sha256"][str(relative)]
            elif str(path) != record["catalog_path"]:
                assert str(path) in manifest["reported_external_guidance_sha256"], opened
        completed.append(record["id"])

    review_map = json.loads((RUN / "review-map.json").read_text())
    for review in review_map["cases"]:
        neutral = Path(review_map["directory"]) / review["review_id"]
        archived = RUN / "samples" / review["sample"]
        assert inventory(neutral / "module") == review["source_sha256"]
        assert inventory(neutral / "base") == inventory(archived / "original")
        assert (neutral / "task.txt").read_bytes() == (archived / "prompt.txt").read_bytes()
    assert len(review_map["cases"]) == len(completed) == 8
    assert len(skipped) == 8
    review_ids = {review["review_id"]: review["sample"] for review in review_map["cases"]}
    sample_records = {record["id"]: record for record in manifest["cases"]}
    review_files = RUN / "blind-review"
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
            sample = RUN / "samples" / sample_id
            if path.parts[1] in {"base", "module"}:
                folder = "original" if path.parts[1] == "base" else "candidate"
                preserved = sample / folder / Path(*path.parts[2:])
            elif path.parts[1] == "task.txt":
                preserved = sample / "prompt.txt"
            else:
                assert path.parts[1] == "contract-probes.go.txt"
                preserved = RUN / "probes" / (sample_records[sample_id]["case"] + "_test.go.txt")
            assert digest(preserved) == expected, relative
    evidence = json.loads((review_files / "evidence/manifest.json").read_text())
    for record in evidence["files"]:
        assert digest(review_files / record["path"]) == record["sha256"]
    return {
        "status": "pass",
        "completed_baseline_samples": completed,
        "skipped_guided_samples": skipped,
        "checks": ["input and live fixture hashes", "unchanged runtime and prior draft",
                   "guidance, reported external files and probe hashes", "catalogs without error guidance",
                   "candidate and patch hashes", "exact patch reconstructions",
                   "recorded held tests and candidate checks", "adjudicated CLI statuses and output",
                   "neutral review inputs unchanged", "review inputs/outputs and preserved probes"],
        "limits": ["Authors report opened guidance; full tool traces are not archived",
                   "This integrity check does not rerun already recorded Go tests"]
    }


if __name__ == "__main__":
    result = verify()
    (RUN / "integrity.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result))
