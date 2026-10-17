"""Prepare matched author inputs or check a candidate in a disposable copy."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile


ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]
SKILL = "go-input-and-serialization"


def hashes(root):
    result = {}
    for path in sorted(root.rglob("*")):
        if ".git" in path.relative_to(root).parts:
            continue
        if path.is_symlink():
            raise ValueError(f"symlink is not a trial input: {path}")
        if path.is_file():
            result[path.relative_to(root).as_posix()] = hashlib.sha256(path.read_bytes()).hexdigest()
    return result


def prepare(case, arm, destination, skill=None):
    destination = destination.resolve()
    if destination.is_relative_to(REPO.resolve()):
        raise ValueError("generated trials must stay outside the repository")
    if destination.exists():
        raise FileExistsError(destination)
    suite = json.loads((ROOT / "evals/evals.json").read_text())
    selected = next(item for item in suite["evals"] if item["id"] == case)
    fixture = ROOT / "evals/files" / case
    inputs = hashes(fixture)
    expected = {f"files/{case}/{name}" for name in inputs}
    if set(selected["files"]) != expected:
        raise ValueError("fixture inventory differs from evaluation inputs")
    skills = sorted((REPO / "plugins/go-quality-build/skills").iterdir())
    skills = [path for path in skills if path.is_dir() and path.name != SKILL]
    if arm == "skill-on":
        skill = skill or REPO / "plugins/go-quality-build/skills" / SKILL
        if not (skill / "SKILL.md").is_file():
            raise ValueError("skill-on requires --skill or the promoted runtime skill")
        skills.append(skill)
    destination.mkdir(parents=True)
    shutil.copytree(fixture, destination / "input")
    shutil.copytree(fixture, destination / "source")
    descriptions = []
    for path in skills:
        hashes(path)
        name = SKILL if path == skill else path.name
        copy = destination / "catalog" / name
        shutil.copytree(path, copy)
        description = next(line.partition(":")[2].strip()
                           for line in (copy / "SKILL.md").read_text().splitlines()
                           if line.startswith("description:"))
        descriptions.append(f"{name}: {description}\nPath: {copy / 'SKILL.md'}")
    prompt = (
        f"Complete this Go change in {destination / 'source'}.\n\n{selected['prompt']}\n\n"
        "Read README.md and existing code/tests. Select and open relevant skills from the catalog "
        "below before editing. The only allowed task inputs are source/, selected catalog files, "
        "and primary Go dependency documentation/source. Other repository files, skill catalogs, "
        "probes, input/, manifest.json, other trials and results are outside allowed inputs. "
        "Do not use agents. Preserve the module's minimum Go version. Use rtk for shell commands.\n\n"
        + "\n\n".join(descriptions)
        + f"\n\nImplement and verify in source/ only. Keep binaries outside source/. Do not stage "
        f"or commit. Write {destination / 'report.md'} with selected/opened skills, decisions, "
        "actual checks and limits. Do not read any other trial files.\n"
    )
    (destination / "prompt.txt").write_text(prompt)
    manifest = {
        "case": case, "arm": arm, "input_sha256": inputs,
        "catalog_sha256": hashes(destination / "catalog"),
        "prompt_sha256": hashlib.sha256((destination / "prompt.txt").read_bytes()).hexdigest(),
        "exposure": "description catalog with author-selected relevant skills",
    }
    (destination / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    return destination


def command(arguments, source):
    try:
        result = subprocess.run(arguments, cwd=source, capture_output=True, text=True,
                                timeout=90, env={**os.environ, "GOWORK": "off",
                                                "GOTOOLCHAIN": "local", "GOPROXY": "off",
                                                "GOCACHE": os.environ.get("GOCACHE") or str(
                                                    Path(tempfile.gettempdir()) / "go-input-evaluation-cache")})
        return {"command": arguments, "exit_code": result.returncode,
                "output": result.stdout + result.stderr}
    except subprocess.TimeoutExpired as error:
        return {"command": arguments, "exit_code": None, "output": str(error), "timeout": True}


def check(case, candidate, go="go", race=False):
    executable = shutil.which(go)
    if executable is None or not candidate.is_dir():
        raise ValueError("a Go executable and candidate module are required")
    probe = ROOT / "probes" / f"{case}_test.go.txt"
    if not probe.is_file():
        raise ValueError("missing private contract probes")
    report = {"case": case, "source_sha256": hashes(candidate)}
    with tempfile.TemporaryDirectory(prefix="input-contract-check-") as temporary:
        source = Path(temporary) / "source"
        shutil.copytree(candidate, source, ignore=shutil.ignore_patterns(".git"))
        test = [executable, "test", "-mod=readonly", "-count=1", "-timeout=30s"]
        if race:
            test.append("-race")
        report["toolchain"] = command([executable, "version"], source)
        report["candidate_tests"] = command(test + ["./..."], source)
        report["vet"] = command([executable, "vet", "-mod=readonly", "./..."], source)
        consumer = Path(temporary) / "consumer"
        consumer.mkdir()
        fixture_module = (ROOT / "evals/files" / case / "go.mod").read_text()
        module = re.search(r"^module\s+(\S+)", fixture_module, re.M).group(1)
        (consumer / "go.mod").write_text(
            f"module example.com/inputprobe\n\ngo 1.22.0\n\nrequire {module} v0.0.0\n"
            f"replace {module} => {json.dumps(str(source))}\n")
        (consumer / "contract_test.go").write_bytes(probe.read_bytes())
        report["contract_tests"] = command(test + ["-json", "./..."], consumer)
        events = []
        for line in report["contract_tests"]["output"].splitlines():
            try:
                event = json.loads(line)
            except ValueError:
                continue
            if event.get("Action") in ("pass", "fail", "skip") and event.get("Test"):
                events.append({"test": event["Test"], "result": event["Action"]})
        report["tests"] = events
        expected = set(re.findall(r"^func (Test\w+)\(", probe.read_text(), re.M))
        completed = {event["test"] for event in events if event["result"] in ("pass", "fail")}
        report["missing_probe_tests"] = sorted(expected - completed)
    report["pass"] = all(report[name]["exit_code"] == 0
                         for name in ("toolchain", "candidate_tests", "vet", "contract_tests"))
    report["pass"] = report["pass"] and bool(expected) and not report["missing_probe_tests"]
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    actions = parser.add_subparsers(dest="action", required=True)
    prepare_parser = actions.add_parser("prepare")
    prepare_parser.add_argument("case")
    prepare_parser.add_argument("arm", choices=("baseline", "skill-on"))
    prepare_parser.add_argument("destination", type=Path)
    prepare_parser.add_argument("--skill", type=Path)
    check_parser = actions.add_parser("check")
    check_parser.add_argument("case")
    check_parser.add_argument("candidate", type=Path)
    check_parser.add_argument("--go", default="go")
    check_parser.add_argument("--race", action="store_true")
    args = parser.parse_args()
    try:
        if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", args.case):
            raise ValueError("invalid case ID")
        if args.action == "prepare":
            print(prepare(args.case, args.arm, args.destination, args.skill))
            return 0
        report = check(args.case, args.candidate, args.go, args.race)
        print(json.dumps(report, indent=2))
        return 0 if report["pass"] else 1
    except (ValueError, OSError, StopIteration) as error:
        parser.error(str(error))


if __name__ == "__main__":
    raise SystemExit(main())
