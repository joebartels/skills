"""Run a candidate's tests and withheld probes in a disposable copy."""

import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


ROOT = Path(__file__).resolve().parent


def main():
    cases = json.loads((ROOT / "evals/evals.json").read_text())["evals"]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("case", choices=[case["id"] for case in cases])
    parser.add_argument("source", type=Path, help="candidate Go module directory")
    parser.add_argument("--go", default="go", help="Go executable or path")
    parser.add_argument("--race", action="store_true")
    args = parser.parse_args()
    go = shutil.which(args.go)
    if not go or not args.source.is_dir():
        parser.error("a Go executable and candidate source directory are required")
    with tempfile.TemporaryDirectory(prefix="client-call-check-") as temporary:
        source = Path(temporary) / "source"
        shutil.copytree(args.source, source, ignore=shutil.ignore_patterns(".git"))
        probes = [ROOT / f"probes/{args.case}_test.go.txt"]
        probes += sorted((ROOT / "probes" / args.case).glob("*.go.txt"))
        for index, probe in enumerate(probes):
            shutil.copyfile(probe, source / f"zz_contract_{index}_test.go")
        command = [str(Path(go).resolve()), "test", "-mod=readonly", "-count=1", "-timeout=30s"]
        if args.race:
            command.append("-race")
        result = subprocess.run(command + ["./..."], cwd=source,
                                env={**os.environ, "GOWORK": "off", "GOTOOLCHAIN": "local"})
    return result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
