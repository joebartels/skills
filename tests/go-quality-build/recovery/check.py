#!/usr/bin/env python3
"""Replay only the compact recovery trials, in disposable directories."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(command, cwd, env):
    try:
        result = subprocess.run(command, cwd=cwd, env=env,
                                capture_output=True, text=True, timeout=90)
        return {"command": command, "exit_code": result.returncode,
                "output": result.stdout + result.stderr}
    except subprocess.TimeoutExpired:
        return {"command": command, "exit_code": None, "output": "controller deadline exceeded"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--trial", action="append", help="Select names; default is every recorded trial")
    parser.add_argument("--probe", action="append", default=[], help="Additional labelled post-inspection probes")
    parser.add_argument("--output", type=Path, required=True, help="Write a fresh result outside sealed historical evidence")
    args = parser.parse_args()
    here = Path(__file__).resolve().parent
    root = here.parents[2]
    manifest = json.loads((here / "trials.json").read_text())
    for relative, expected in manifest["guidance_sha256"].items():
        assert digest(root / "plugins/go-quality-build/skills" / relative) == expected, relative
    for relative, expected in manifest.get("probe_sha256", {}).items():
        assert digest(root / relative) == expected, relative
    for relative, expected in manifest.get("concurrency_shared_guidance_sha256", {}).items():
        assert digest(root / "plugins/go-quality-build/skills" / relative) == expected, relative
    historical_guidance = set()
    for skill, revision in manifest.get("guidance_revisions", {}).items():
        relative = skill + "/SKILL.md"
        current = root / "plugins/go-quality-build/skills" / relative
        latest = "revision_2" if "delta" in revision else "revision_1"
        assert digest(current) == revision[latest], relative + ": " + latest
        if "delta" in revision:
            delta = here / revision["delta"]
            assert digest(delta) == revision["delta_sha256"], revision["delta"]
            with tempfile.TemporaryDirectory(prefix="go-guidance-revision-") as tmp:
                previous = Path(tmp) / "SKILL.md"
                shutil.copyfile(current, previous)
                applied = run(["git", "apply", "--reverse", "--unidiff-zero", str(delta)],
                              tmp, os.environ)
                assert applied["exit_code"] == 0, applied["output"]
                assert digest(previous) == revision["revision_1"], relative + ": revision_1"
                historical_guidance.add((relative, revision["revision_1"]))
    for trial in manifest["trials"]:
        for field in ("extra_guidance_sha256", "shared_guidance_sha256"):
            for relative, expected in trial.get(field, {}).items():
                current = root / "plugins/go-quality-build/skills" / relative
                assert (digest(current) == expected or
                        (relative, expected) in historical_guidance), trial["name"] + ": " + relative
    results = []
    for trial in manifest["trials"]:
        if args.trial and trial["name"] not in args.trial:
            continue
        with tempfile.TemporaryDirectory(prefix="go-recovery-", dir="/private/tmp") as tmp:
            module = Path(tmp) / "module"
            shutil.copytree(root / trial["input"], module)
            if "historical_readme" in trial:
                shutil.copyfile(here / trial["historical_readme"], module / "README.md")
            actual_input = {str(p.relative_to(module)): digest(p)
                            for p in module.rglob("*") if p.is_file()}
            assert actual_input == trial["input_sha256"], trial["name"]
            env = dict(os.environ, GOWORK="off", GOCACHE="/private/tmp/go-skill-recovery-cache", GOTOOLCHAIN="local")
            patch = here / trial["patch"]
            assert digest(patch) == trial["patch_sha256"], trial["name"]
            checks = []
            if patch.stat().st_size:
                applied = run(["git", "apply", "--unidiff-zero", str(patch)], module, env)
                checks.append(applied)
                if applied["exit_code"] != 0:
                    results.append({"name": trial["name"], "checks": checks})
                    continue
            actual = {str(p.relative_to(module)): digest(p) for p in module.rglob("*") if p.is_file()}
            assert actual == trial["output_sha256"], trial["name"]
            checks.append(run(["go", "test", "-count=1", "-timeout=15s", "./..."], module, env))
            checks.append(run(["go", "vet", "./..."], module, env))
            for probe in trial.get("probes", []) + args.probe:
                shutil.copyfile(root / probe, module / ("controller_" + Path(probe).name))
            checks.append(run(["go", "test", "-count=1", "-timeout=15s", "./..."], module, env))
            floor = dict(env, GOPATH="/private/tmp/go-skill-recovery-toolchains", GOTOOLCHAIN="go1.22.12", CGO_ENABLED="0")
            checks.append(run(["go", "test", "-count=1", "-timeout=15s", "./..."], module, floor))
            if trial.get("race"):
                checks.append(run(["go", "test", "-race", "-shuffle=on", "-count=3", "-timeout=30s", "./..."], module, env))
            results.append({"name": trial["name"], "checks": checks})
    assert results, "No trials selected"
    args.output.write_text(json.dumps({"current_go": run(["go", "version"], root, os.environ),
                                     "additional_probes": {probe: digest(root / probe) for probe in args.probe},
                                     "results": results}, indent=2) + "\n")
    for result in results:
        failures = sum(check["exit_code"] != 0 for check in result["checks"])
        print(f'{result["name"]}: {len(result["checks"])} checks, {failures} failures')
    return int(any(check["exit_code"] != 0 for result in results for check in result["checks"]))


if __name__ == "__main__":
    raise SystemExit(main())
