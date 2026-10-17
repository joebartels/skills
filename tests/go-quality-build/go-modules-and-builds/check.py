"""Check module consumers, generation and build targets in disposable copies."""

import argparse
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import zipfile


ROOT = Path(__file__).resolve().parent
CASES = ("workspace-library", "replacement-transfer", "generated-targets", "pure-control")


def module_path(directory):
    match = re.search(r'^\s*module\s+("[^"\n]+"|[^\s/]+(?:/[^\s]+)*)',
                      (directory / "go.mod").read_text(), re.MULTILINE)
    if not match:
        raise ValueError(f"missing module directive: {directory / 'go.mod'}")
    return match.group(1).strip('"')


def escaped(value):
    """Apply the Go proxy protocol's uppercase escaping."""
    return "".join("!" + char.lower() if "A" <= char <= "Z" else char for char in value)


def build_proxy(registry, destination):
    """Publish registry/<name>/<version> modules as a deterministic file proxy."""
    destination.mkdir(parents=True, exist_ok=True)
    versions = {}
    for manifest in sorted(registry.glob("*/*/go.mod")):
        source = manifest.parent
        version = source.name
        if not re.fullmatch(r"v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?", version):
            raise ValueError(f"invalid registry version: {version}")
        module = module_path(source)
        target = destination / escaped(module) / "@v"
        target.mkdir(parents=True, exist_ok=True)
        stem = escaped(version)
        (target / (stem + ".mod")).write_bytes(manifest.read_bytes())
        (target / (stem + ".info")).write_text(json.dumps(
            {"Version": version, "Time": "2000-01-01T00:00:00Z"}, sort_keys=True) + "\n")
        with zipfile.ZipFile(target / (stem + ".zip"), "w", zipfile.ZIP_DEFLATED) as archive:
            for path in sorted(source.rglob("*")):
                if path.is_file():
                    name = f"{module}@{version}/{path.relative_to(source).as_posix()}"
                    info = zipfile.ZipInfo(name, date_time=(2000, 1, 1, 0, 0, 0))
                    info.external_attr = 0o100644 << 16
                    archive.writestr(info, path.read_bytes())
        versions.setdefault(target, set()).add(version)
    for target, entries in versions.items():
        (target / "list").write_text("\n".join(sorted(entries)) + "\n")
    return destination.resolve().as_uri()


def environment(go, temporary, proxy):
    """Use only local toolchain/configuration and disposable Go state."""
    env = {key: os.environ[key] for key in ("PATH", "SYSTEMROOT", "WINDIR", "TMPDIR")
           if key in os.environ}
    env.update(PATH=str(Path(go).parent) + os.pathsep + env.get("PATH", os.defpath),
               HOME=str(temporary / "home"), GOPATH=str(temporary / "gopath"),
               GOMODCACHE=str(temporary / "modules"), GOCACHE=str(temporary / "cache"),
               GOENV="off", GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="",
               GOPROXY=proxy, GOSUMDB="off", GONOPROXY="none", GONOSUMDB="none",
               GOPRIVATE="", GO111MODULE="on", GOTELEMETRY="off", CGO_ENABLED="0")
    Path(env["HOME"]).mkdir()
    return env


def install_probe(directory, probe):
    # NamedTemporaryFile avoids hiding any authored candidate test.
    with tempfile.NamedTemporaryFile(dir=directory, prefix="zz_module_probe_",
                                     suffix="_test.go", delete=False) as output:
        output.write(probe.read_bytes())


def check_candidate(case, source, go, *, probes_dir=None, timeout=120):
    """Return exact command evidence; never modify the candidate or global Go state."""
    if case not in CASES:
        raise ValueError(f"unknown case: {case}")
    source = Path(source).resolve()
    executable = shutil.which(str(go))
    if not executable or not source.is_dir():
        raise ValueError("a Go executable and candidate source directory are required")
    executable = str(Path(executable).resolve())
    probes_dir = Path(probes_dir) if probes_dir is not None else ROOT / "probes"
    report = {"case": case, "success": False, "checks": []}

    with tempfile.TemporaryDirectory(prefix="modules-builds-check-") as work:
        # Go workspace membership compares absolute paths; resolve macOS /var aliases.
        temporary = Path(work).resolve()
        candidate = temporary / "source"
        shutil.copytree(source, candidate, ignore=shutil.ignore_patterns(".git"))
        proxy = build_proxy(candidate / "registry", temporary / "proxy")
        env = environment(executable, temporary, proxy)

        def run(name, arguments, directory, scope=None):
            command = [executable, *arguments]
            scoped_env = {**env, **(scope or {})}
            check = {"name": name, "command": command, "cwd": str(directory),
                     "env": scoped_env}
            try:
                result = subprocess.run(command, cwd=directory, env=scoped_env,
                                        capture_output=True, text=True, timeout=timeout)
                check.update(stdout=result.stdout, stderr=result.stderr,
                             exit_code=result.returncode)
            except subprocess.TimeoutExpired as error:
                def decoded(value):
                    return value.decode(errors="replace") if isinstance(value, bytes) else value or ""
                check.update(stdout=decoded(error.stdout),
                             stderr=decoded(error.stderr) + f"\ncommand timed out after {timeout}s",
                             exit_code=124)
            report["checks"].append(check)
            return check["exit_code"] == 0

        def tests(name, directory, *, workspace="off", tags=None, pattern=None):
            arguments = ["test", "-count=1", "-timeout=30s"]
            if workspace == "off":
                arguments.append("-mod=readonly")
            if tags:
                arguments += ["-tags", tags]
            if pattern:
                arguments += ["-run", pattern]
            return run(name, arguments + ["./..."], directory, {"GOWORK": workspace})

        def condition(name, passed, message):
            report["checks"].append({"name": name, "command": [], "cwd": str(candidate),
                                     "env": {}, "stdout": message if passed else "",
                                     "stderr": "" if passed else message,
                                     "exit_code": 0 if passed else 1})

        def add_probe(directory):
            probe = probes_dir / f"{case}_test.go.txt"
            condition("withheld-probe-present", probe.is_file(), str(probe))
            if probe.is_file():
                install_probe(directory, probe)

        run("toolchain", ["version"], candidate)
        if case in {"workspace-library", "replacement-transfer"}:
            library = candidate / "lib"
            tests("workspace-tests", library, workspace=str(candidate / "go.work"))
            add_probe(library)
            tests("standalone-tests", library)
            consumer_probe = probes_dir / f"{case}_consumer.go.txt"
            condition("consumer-probe-present", consumer_probe.is_file(), str(consumer_probe))
            if consumer_probe.is_file():
                consumer = temporary / "consumer"
                consumer.mkdir()
                module = module_path(library)
                major = re.search(r"/v([2-9][0-9]*)$", module)
                version = f"v{major.group(1)}.0.0" if major else "v0.0.0"
                (consumer / "go.mod").write_text(
                    f"module checker.invalid/consumer\n\ngo 1.22.0\n\n"
                    f"require {module} {version}\nreplace {module} => {library.as_posix()}\n")
                (consumer / "main.go").write_bytes(consumer_probe.read_bytes())
                if run("consumer-tidy", ["mod", "tidy"], consumer):
                    run("consumer-run", ["run", "."], consumer)
        elif case == "generated-targets":
            add_probe(candidate)
            tests("initial-default-tests", candidate)
            tests("initial-compact-tests", candidate, tags="compact")
            for path in candidate.rglob("*_gen.go"):
                path.unlink()
            run("generate-clean", ["generate", "./..."], candidate)
            generated = {str(path.relative_to(candidate)): path.read_bytes()
                         for path in candidate.rglob("*_gen.go")}
            condition("generated-files-present", bool(generated), "generation must recreate *_gen.go files")
            tests("generated-default-tests", candidate)
            tests("generated-compact-tests", candidate, tags="compact")
            run("generate-repeat", ["generate", "./..."], candidate)
            repeated = {str(path.relative_to(candidate)): path.read_bytes()
                        for path in candidate.rglob("*_gen.go")}
            condition("generation-byte-identical", generated == repeated,
                      "generated output must be byte-identical on repeated generation")
            for architecture in ("amd64", "arm64"):
                for tags in (None, "compact"):
                    label = f"compile-only-linux-{architecture}-" + (tags or "default")
                    output = temporary / label
                    output.mkdir()
                    arguments = ["test", "-mod=readonly", "-c", "-o", str(output)]
                    if tags:
                        arguments += ["-tags", tags]
                    run(label, arguments + ["./..."], candidate,
                        {"GOOS": "linux", "GOARCH": architecture, "CGO_ENABLED": "0"})
            # Vary the source only after checks bound to the original schema finish.
            schema = candidate / "schema.txt"
            schema_probe = probes_dir / "generated-targets-schema_test.go.txt"
            condition("schema-source-present", schema.is_file(), str(schema))
            condition("schema-probe-present", schema_probe.is_file(), str(schema_probe))
            if schema.is_file() and schema_probe.is_file():
                original = schema.read_bytes()
                separator = b"" if not original or original.endswith(b"\n") else b"\n"
                schema.write_bytes(original + separator + b"50 fatal\n")
                condition("schema-source-variation", True, "appended '50 fatal' to disposable schema.txt")
                run("generate-schema", ["generate", "./..."], candidate)
                install_probe(candidate, schema_probe)
                # Authored tests may correctly assert 50 is unknown in the original schema.
                tests("schema-default-tests", candidate, pattern="^TestContractSchemaSource$")
                tests("schema-compact-tests", candidate, tags="compact",
                      pattern="^TestContractSchemaSource$")
        else:
            add_probe(candidate)
            tests("control-tests", candidate)

    report["success"] = all(check["exit_code"] == 0 for check in report["checks"])
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("case", choices=CASES)
    parser.add_argument("source", type=Path, help="candidate fixture directory")
    parser.add_argument("--go", default="go", help="Go executable or path")
    parser.add_argument("--json", type=Path, help="write full command evidence to this file")
    args = parser.parse_args()
    try:
        report = check_candidate(args.case, args.source, args.go)
    except (ValueError, OSError) as error:
        parser.error(str(error))
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, indent=2) + "\n")
    else:
        print(json.dumps(report, indent=2))
    return 0 if report["success"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
