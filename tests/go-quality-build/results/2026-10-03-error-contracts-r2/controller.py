"""Replay frozen error-contract samples in disposable modules."""

import argparse
import difflib
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


RUN = Path(__file__).resolve().parent


def inventory(folder):
    return {
        str(path.relative_to(folder)): hashlib.sha256(path.read_bytes()).hexdigest()
        for path in sorted(folder.rglob("*")) if path.is_file()
    }


def command(args, cwd, env):
    try:
        result = subprocess.run(
            ["rtk", "proxy", *args], cwd=cwd, env=env,
            text=True, capture_output=True, timeout=60,
        )
        return {"command": args, "exit_code": result.returncode,
                "output": result.stdout + result.stderr}
    except subprocess.TimeoutExpired:
        return {"command": args, "exit_code": None, "output": "60-second timeout"}


def archive_sample(sample):
    manifest = json.loads((RUN / "manifest.json").read_text())
    record = next(case for case in manifest["cases"] if case["id"] == sample)
    archive = RUN / "samples" / sample
    source = Path(record["workdir"])
    assert Path(record["report_path"]).is_file(), "author report is missing"
    candidate = archive / "candidate"
    assert not candidate.exists(), "sample already archived"
    shutil.copytree(source, candidate)
    shutil.copy(record["report_path"], archive / "author-report.json")
    original = archive / "original"
    assert inventory(original) == record["input_sha256"], "frozen input changed"
    paths = sorted(set(inventory(original)) | set(inventory(candidate)))
    patch = []
    for relative in paths:
        old, new = original / relative, candidate / relative
        old_lines = old.read_text().splitlines(keepends=True) if old.exists() else []
        new_lines = new.read_text().splitlines(keepends=True) if new.exists() else []
        if old_lines != new_lines:
            patch.extend(difflib.unified_diff(
                old_lines, new_lines,
                fromfile=f"a/{relative}" if old.exists() else "/dev/null",
                tofile=f"b/{relative}" if new.exists() else "/dev/null",
            ))
    (archive / "source.patch").write_text("".join(patch))
    environment = dict(os.environ, GOWORK="off", GOTOOLCHAIN="local",
                       GOCACHE=manifest["cache"])
    checks = []
    with tempfile.TemporaryDirectory(prefix="error-contract-replay-", dir="/private/tmp") as directory:
        copy = Path(directory) / "module"
        shutil.copytree(original, copy)
        if patch:
            applied = command(["git", "apply", str(archive / "source.patch")], copy, environment)
            assert applied["exit_code"] == 0, applied
        assert inventory(copy) == inventory(candidate), "patch reconstruction differs"
        for args in [["go", "test", "-count=1", "./..."], ["go", "vet", "./..."],
                     ["gofmt", "-l", "."]]:
            check = command(args, copy, environment)
            check["scope"] = "candidate"
            checks.append(check)
        shutil.copy(RUN / "probes" / (record["case"] + "_test.go.txt"),
                    copy / "zz_controller_contract_test.go")
        check = command(["go", "test", "-count=1", "-run", "TestContract", "./..."], copy, environment)
        check["scope"] = "held-contract-probes"
        checks.append(check)
        if record["case"] == "cli-completion":
            binary = Path(directory) / "lineexport"
            built = command(["go", "build", "-o", str(binary), "./cmd/lineexport"], copy, environment)
            built["scope"] = "actual-cli-build"
            checks.append(built)
            if built["exit_code"] == 0:
                process_root = Path(directory) / "process"
                process_root.mkdir()
                for name, contents, limit, status, output in [
                    ("success", "a\nb\n", None, 0, "a\nb\n"),
                    ("limited", "a\nb\nc\n", "2", 0, "a\nb\n"),
                    ("negative", "a\n", "-1", 1, ""),
                ]:
                    (process_root / "-input").write_text(contents)
                    target = process_root / "-output"
                    target.unlink(missing_ok=True)
                    args = [str(binary), "-input", "-output"]
                    if limit is not None:
                        args.append(limit)
                    observed = command(args, process_root, environment)
                    observed.update(scope="actual-cli", name=name, expected_status=status,
                                    output_bytes=target.read_text() if target.exists() else None,
                                    expected_output=output)
                    checks.append(observed)
    (archive / "verification.json").write_text(json.dumps(checks, indent=2) + "\n")
    record.update(status="archived", source_sha256=inventory(candidate),
                  patch_sha256=hashlib.sha256((archive / "source.patch").read_bytes()).hexdigest())
    (RUN / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps({"id": sample, "case": record["case"], "checks": [
        {"scope": check["scope"], "exit_code": check["exit_code"]} for check in checks]}))


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("sample")
    archive_sample(parser.parse_args().sample)
