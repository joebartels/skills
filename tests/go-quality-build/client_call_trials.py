"""Prepare sealed inputs and reconstruct complete outputs for client-call trials.

No model runner. Authors may read only prompt.txt, catalog/ and source/, and
write source/ and trial-report.md. Reviewers receive separately neutralized code.
"""

import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import tempfile

REPO = Path(__file__).resolve().parents[2]
CATALOG = ("go-package-boundaries", "go-api-contracts", "go-interfaces-and-composition",
           "go-names-and-comments", "go-behavior-tests", "go-test-isolation",
           "go-context-and-deadlines", "go-concurrency-and-ownership")


def run(args, cwd):
    result = subprocess.run(["rtk", "proxy", *args], cwd=cwd,
                            text=True, capture_output=True)
    if result.returncode:
        raise RuntimeError(f"{args}: {result.stdout}\n{result.stderr}")
    return result.stdout


def hashes(root):
    result = {}
    for path in sorted(root.rglob("*")):
        if ".git" in path.relative_to(root).parts:
            continue
        if path.is_symlink():
            raise ValueError(f"symlinks are not trial inputs: {path}")
        if path.is_file():
            result[path.relative_to(root).as_posix()] = hashlib.sha256(path.read_bytes()).hexdigest()
    return result


def component(value):
    if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", value):
        raise ValueError(f"invalid trial component: {value!r}")
    return value


def prepare_trial(case: str, arm: str, trial_id: str, scratch: Path, archive: Path) -> Path:
    component(case)
    component(trial_id)
    if arm not in ("baseline", "skill-on"):
        raise ValueError("arm must be baseline or skill-on")
    trial = scratch.resolve() / trial_id
    if trial.exists() or (archive.resolve() / trial_id).exists():
        raise FileExistsError(trial_id)
    evals = REPO / "tests/go-quality-build/go-client-calls/evals"
    suite = json.loads((evals / "evals.json").read_text())
    selected = next(item for item in suite["evals"] if item["id"] == case)
    fixture = evals / "files" / case
    source_files = hashes(fixture)
    expected = []
    for item in selected["files"]:
        path = Path(item)
        if path.is_absolute() or ".." in path.parts or not path.parts[:2] == ("files", case):
            raise ValueError("input path escapes fixture")
        expected.append(Path(*path.parts[2:]).as_posix())
    if set(expected) != set(source_files):
        raise ValueError("fixture inventory differs from frozen case")
    trial.mkdir(parents=True)
    shutil.copytree(fixture, trial / "input")
    shutil.copytree(fixture, trial / "source")
    catalog = trial / "catalog"
    for name in CATALOG:
        original = REPO / "plugins/go-quality-build/skills" / name
        hashes(original)
        shutil.copytree(original, catalog / name)
    if arm == "skill-on":
        original = evals.parent / "draft"
        if not original.exists():
            original = REPO / "plugins/go-quality-build/skills/go-client-calls"
        hashes(original)
        shutil.copytree(original, catalog / "go-client-calls")
    descriptions = []
    for skill in sorted(catalog.glob("*/SKILL.md")):
        description = next(line.removeprefix("description:").strip()
                           for line in skill.read_text().splitlines() if line.startswith("description:"))
        descriptions.append(f"{skill.parent.name}: {description}\nPath: {skill}")
    prompt = (f"Complete this Go change in {trial / 'source'}.\n\n{selected['prompt']}\n\n"
              "Read README and existing code/tests. Select relevant skills from this catalog and read "
              "their SKILL.md before editing. Other skills, repository docs, private probes, other trials, "
              "input/, metadata.json and results are outside allowed inputs. Do not use agents. "
              "Use only the task files, selected catalog files and primary dependency documentation/source.\n\n"
              + "\n\n".join(descriptions)
              + f"\n\nImplement and verify in source only. Do not stage or commit. Keep binaries outside source. "
              f"Write {trial / 'trial-report.md'} with selected/opened skill paths, decisions, checks "
              "actually run and limits. Prefix shell commands with rtk. No external actions.\n")
    (trial / "prompt.txt").write_text(prompt)
    source = trial / "source"
    run(["git", "init", "-q"], source)
    run(["git", "add", "."], source)
    run(["git", "-c", "commit.gpgsign=false", "-c", "user.name=Evaluation",
         "-c", "user.email=evaluation@example.invalid", "commit", "-qm", "trial input"], source)
    metadata = {"case": case, "arm": arm, "trial_id": trial_id,
                "archive_root": str(archive.resolve()), "input_hashes": source_files,
                "catalog_hashes": hashes(catalog),
                "prompt_sha256": hashlib.sha256(prompt.encode()).hexdigest(),
                "base_commit": run(["git", "rev-parse", "HEAD"], source).strip(),
                "model": "inherited; exact runtime ID unavailable", "effort": "inherited; exact value unavailable",
                "exposure": "description catalog; manual selection reported by author"}
    (trial / "metadata.json").write_text(json.dumps(metadata, indent=2) + "\n")
    return trial


def archive_trial(trial: Path, archive: Path) -> dict:
    metadata = json.loads((trial / "metadata.json").read_text())
    trial_id = component(metadata["trial_id"])
    if archive.resolve() != Path(metadata["archive_root"]):
        raise ValueError("archive root changed")
    destination = archive.resolve() / trial_id
    if destination.exists():
        raise FileExistsError(destination)
    for name, key in (("catalog", "catalog_hashes"), ("input", "input_hashes")):
        if hashes(trial / name) != metadata[key]:
            raise ValueError(f"{name} mutated after preparation")
    if hashlib.sha256((trial / "prompt.txt").read_bytes()).hexdigest() != metadata["prompt_sha256"]:
        raise ValueError("prompt mutated after preparation")
    source = trial / "source"
    output_hashes = hashes(source)
    run(["git", "add", "--all"], source)
    patch_text = run(["git", "diff", "--binary", "--full-index", metadata["base_commit"]], source)
    archive.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="seal-", dir=archive) as temporary:
        staging = Path(temporary)
        for name in ("input", "catalog"):
            shutil.copytree(trial / name, staging / name)
        for name in ("prompt.txt", "trial-report.md"):
            shutil.copyfile(trial / name, staging / name)
        patch_file = staging / "source.patch"
        patch_file.write_text(patch_text)
        # Git apply from a subdirectory of a parent repository can skip paths.
        # Reconstruct outside that repository before sealing a plain file tree.
        with tempfile.TemporaryDirectory(prefix="client-reconstruct-") as reconstructed:
            check = Path(reconstructed) / "output"
            shutil.copytree(trial / "input", check)
            if patch_text:
                run(["git", "apply", str(patch_file)], check)
            if hashes(check) != output_hashes:
                raise ValueError("patch reconstruction mismatch")
            shutil.copytree(check, staging / "output")
        metadata.update(output_hashes=output_hashes,
                        patch_sha256=hashlib.sha256(patch_text.encode()).hexdigest(), reconstruction="PASS")
        (staging / "manifest.json").write_text(json.dumps(metadata, indent=2) + "\n")
        staging.rename(destination)
    return metadata


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("prepare", "archive"))
    parser.add_argument("--case", required=True)
    parser.add_argument("--arm", choices=("baseline", "skill-on"), required=True)
    parser.add_argument("--trial-id", required=True)
    parser.add_argument("--scratch", type=Path, required=True)
    parser.add_argument("--archive", type=Path, required=True)
    args = parser.parse_args()
    if args.action == "prepare":
        print(prepare_trial(args.case, args.arm, args.trial_id, args.scratch, args.archive))
    else:
        trial = args.scratch.resolve() / component(args.trial_id)
        metadata = json.loads((trial / "metadata.json").read_text())
        if (metadata["case"], metadata["arm"]) != (args.case, args.arm):
            raise ValueError("case/arm mismatch")
        result = archive_trial(trial, args.archive)
        print(json.dumps(result))
