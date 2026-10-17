"""Run authored database modules and controller probes in a disposable copy."""

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
    parser.add_argument("source", type=Path, help="authored Go module directory")
    parser.add_argument("--go", default="go", help="Go executable or path")
    parser.add_argument("--race", action="store_true")
    parser.add_argument("--count", type=int, default=1, help="test repetitions")
    parser.add_argument("--ldflags", default="", help="optional platform linker flags")
    args = parser.parse_args()
    go = shutil.which(args.go)
    if not go or not args.source.is_dir() or args.count < 1:
        parser.error("a Go executable, source directory and positive count are required")
    probes = sorted((ROOT / "probes" / args.case).glob("*.go.txt"))
    if not probes:
        parser.error("controller probes are missing")
    with tempfile.TemporaryDirectory(prefix="database-check-") as temporary:
        source = Path(temporary) / "source"
        shutil.copytree(args.source, source, ignore=shutil.ignore_patterns(".git"))
        for probe in probes:
            with tempfile.NamedTemporaryFile(dir=source, prefix="zz_database_",
                                             suffix="_test.go", delete=False) as target:
                target.write(probe.read_bytes())
        command = [str(Path(go).resolve()), "test", "-mod=readonly",
                   f"-count={args.count}", "-timeout=60s"]
        if args.race:
            command.append("-race")
        if args.ldflags:
            command.append(f"-ldflags={args.ldflags}")
        result = subprocess.run(command + ["./..."], cwd=source,
                                env={**os.environ, "GOWORK": "off", "GOTOOLCHAIN": "local"})
    return result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
