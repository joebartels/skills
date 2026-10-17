"""Exercise the disposable module/build checker with real local Go modules."""

import importlib.util
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import unittest
from unittest.mock import patch


CHECK_PATH = Path(__file__).parent / "go-modules-and-builds/check.py"
GO = shutil.which("go")


def load_checker():
    if not CHECK_PATH.exists():
        return None
    spec = importlib.util.spec_from_file_location("modules_builds_checker", CHECK_PATH)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def write(root, name, content):
    target = root / name
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(content)


def snapshot(root):
    return {str(path.relative_to(root)): path.read_bytes()
            for path in root.rglob("*") if path.is_file() and ".git" not in path.relative_to(root).parts}


@unittest.skipUnless(GO, "Go executable required")
class ModulesBuildsCheckerTests(unittest.TestCase):
    def setUp(self):
        self.checker = load_checker()
        self.assertIsNotNone(self.checker, "disposable checker has not been implemented")
        self.temporary = tempfile.TemporaryDirectory(prefix="module-check-test-")
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)

    def workspace(self, require):
        source = self.root / "candidate"
        write(source, "go.work", "go 1.22.0\nuse (\n ./lib\n ./local\n)\n")
        write(source, "lib/go.mod", "module example.org/library\ngo 1.22.0\n" +
              ("require example.org/Token v1.0.0\n" if require else ""))
        write(source, "lib/value.go", 'package library\nimport "example.org/Token"\nfunc Value() string { return token.Value() }\n')
        write(source, "lib/zz_contract_test.go", 'package library\nimport "testing"\nfunc TestAuthored(t *testing.T) { if Value() != "stable" { t.Fatal(Value()) } }\n')
        for directory in ("local", "registry/token/v1.0.0"):
            write(source, directory + "/go.mod", "module example.org/Token\ngo 1.22.0\n")
            write(source, directory + "/value.go", 'package token\nfunc Value() string { return "stable" }\n')
        probes = self.root / "probes"
        write(probes, "workspace-library_test.go.txt", 'package library\nimport "testing"\nfunc TestProbe(t *testing.T) { if Value() != "stable" { t.Fatal(Value()) } }\n')
        write(probes, "workspace-library_consumer.go.txt", 'package main\nimport "example.org/library"\nfunc main() { if library.Value() != "stable" { panic(library.Value()) } }\n')
        return source, probes

    def test_proxy_resolves_real_dependency_and_preserves_candidate(self):
        # A malformed proxy or leaked workspace must prevent this consumer passing.
        source, probes = self.workspace(require=True)
        proxy = self.checker.build_proxy(source / "registry", self.root / "prep-proxy")
        self.checker.build_proxy(source / "registry", self.root / "second-proxy")
        self.assertEqual(snapshot(self.root / "prep-proxy"), snapshot(self.root / "second-proxy"))
        env = {key: value for key, value in os.environ.items()
               if not key.startswith("GO") and key not in {"CGO_ENABLED"}}
        env.update(GOWORK="off", GOENV="off", GOTOOLCHAIN="local", GOPROXY=proxy,
                   GOSUMDB="off", GOPATH=str(self.root / "prep-gopath"),
                   GOCACHE=str(self.root / "prep-cache"))
        result = subprocess.run([GO, "mod", "tidy"], cwd=source / "lib", env=env,
                                capture_output=True, text=True, timeout=120)
        self.assertEqual(result.returncode, 0, result.stderr)
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-library", source, GO,
                                              probes_dir=probes)
        self.assertTrue(report["success"], report)
        self.assertEqual(snapshot(source), before)
        self.assertTrue(any(check["name"] == "consumer-run" and check["exit_code"] == 0
                            for check in report["checks"]))
        for check in report["checks"]:
            if check.get("command"):
                self.assertEqual(check["env"]["GOTOOLCHAIN"], "local")
                self.assertEqual(check["env"]["GOENV"], "off")
                self.assertEqual(check["command"][0], str(Path(GO).resolve()))

    def test_workspace_success_does_not_credit_missing_library_requirement(self):
        # Omitting the standalone check would incorrectly accept this module.
        source, probes = self.workspace(require=False)
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-library", source, GO,
                                              probes_dir=probes)
        self.assertFalse(report["success"], report)
        checks = {check["name"]: check for check in report["checks"]}
        self.assertEqual(checks["workspace-tests"]["exit_code"], 0, checks)
        self.assertNotEqual(checks["standalone-tests"]["exit_code"], 0, checks)
        self.assertEqual(snapshot(source), before)

    def tidy(self, source):
        prep = self.root / "prepare"
        prep.mkdir(exist_ok=True)
        proxy = self.checker.build_proxy(source / "registry", prep / "proxy")
        with tempfile.TemporaryDirectory(dir=prep) as work:
            env = self.checker.environment(str(Path(GO).resolve()), Path(work), proxy)
            result = subprocess.run([GO, "mod", "tidy"], cwd=source / "lib", env=env,
                                    capture_output=True, text=True, timeout=120)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_workspace_mvs_does_not_credit_missing_release_requirement(self):
        # A companion module selects the API that the library forgot to require.
        source, probes = self.workspace(require=True)
        shutil.rmtree(source / "local")
        write(source, "go.work", "go 1.22.0\nuse (\n ./lib\n ./tools\n)\n")
        write(source, "tools/go.mod", "module example.org/tool\ngo 1.22.0\nrequire example.org/Token v1.1.0\n")
        write(source, "tools/tool.go", "package tool\n")
        write(source, "Makefile", "test:\n\tgo test ./lib/... ./tools/...\nbuild:\n\tgo build ./tools/...\n")
        write(source, "registry/token/v1.1.0/go.mod", "module example.org/Token\ngo 1.22.0\n")
        write(source, "registry/token/v1.1.0/value.go", 'package token\nfunc Cursor() string { return "stable" }\n')
        write(source, "lib/value.go", 'package library\nimport "example.org/Token"\nfunc Value() string { return token.Cursor() }\n')
        for suffix in ("_test.go.txt", "_consumer.go.txt"):
            shutil.copyfile(probes / ("workspace-library" + suffix),
                            probes / ("workspace-selection" + suffix))
        self.tidy(source)
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-selection", source, GO, probes_dir=probes)
        self.assertFalse(report["success"])
        checks = {check["name"]: check for check in report["checks"]}
        self.assertEqual(checks["workspace-tests"]["exit_code"], 0, checks)
        self.assertIn("undefined: token.Cursor", checks["standalone-tests"]["stderr"])
        self.assertEqual(snapshot(source), before)
        write(source, "lib/go.mod", "module example.org/library\ngo 1.22.0\nrequire example.org/Token v1.1.0\n")
        self.tidy(source)
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-selection", source, GO, probes_dir=probes)
        self.assertTrue(report["success"], report)
        self.assertEqual(snapshot(source), before)

    def test_workspace_selection_requires_documented_companion_membership(self):
        # The SDK alone still passes when the command is omitted from go.work.
        source, probes = self.workspace(require=True)
        write(source, "tools/go.mod", "module example.org/toolcli\ngo 1.22.0\n")
        write(source, "tools/main.go", 'package main\nfunc name() string { return "ready" }\nfunc main() { if name() != "ready" { panic("tool broken") } }\n')
        write(source, "tools/main_test.go", 'package main\nimport "testing"\nfunc TestCompanion(t *testing.T) { if name() != "ready" { t.Fatal(name()) } }\n')
        write(source, "Makefile", "test:\n\tgo test ./lib/... ./tools/...\nbuild:\n\tgo build ./tools/...\n")
        for suffix in ("_test.go.txt", "_consumer.go.txt"):
            shutil.copyfile(probes / ("workspace-library" + suffix),
                            probes / ("workspace-selection" + suffix))
        self.tidy(source)
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-selection", source, GO, probes_dir=probes)
        self.assertFalse(report["success"], "documented companion was omitted but checker passed")
        checks = {check["name"]: check for check in report["checks"]}
        self.assertEqual(checks["workspace-tests"]["exit_code"], 0)
        self.assertEqual(checks["standalone-tests"]["exit_code"], 0)
        failures = {check["name"] for check in report["checks"] if check["exit_code"] != 0}
        self.assertEqual(failures, {"workspace-root-tests", "workspace-root-build"})
        self.assertEqual(snapshot(source), before)

        write(source, "go.work", "go 1.22.0\nuse (\n ./lib\n ./local\n ./tools\n)\n")
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-selection", source, GO, probes_dir=probes)
        self.assertTrue(report["success"], report)
        self.assertEqual(snapshot(source), before)
        for check in report["checks"]:
            if check["name"] in {"workspace-root-tests", "workspace-root-build"}:
                self.assertEqual(check["env"]["PATH"].split(os.pathsep)[0], str(Path(GO).resolve().parent))
                self.assertEqual(check["env"]["GOWORK"], str(Path(check["cwd"]) / "go.work"))

    def test_selected_toolchain_tests_its_version_variant(self):
        # Editing only the modern variant must fail when --go selects Go 1.22.
        source = self.root / "candidate"
        write(source, "go.mod", "module example.org/variant\ngo 1.22.0\n")
        write(source, "modern.go", "//go:build go1.23\n\npackage variant\nfunc Value() int { return 3 }\n")
        write(source, "compat.go", "//go:build !go1.23\n\npackage variant\nfunc Value() int { return 2 }\n")
        probes = self.root / "probes"
        write(probes, "toolchain-variants_test.go.txt", 'package variant\nimport "testing"\nfunc TestProbe(t *testing.T) { if Value() != 3 { t.Fatal("variant missing update") } }\n')
        before = snapshot(source)
        report = self.checker.check_candidate("toolchain-variants", source, GO, probes_dir=probes)
        version = subprocess.run([GO, "version"], capture_output=True, text=True, check=True).stdout
        minor = int(re.search(r"go1\.(\d+)", version).group(1))
        self.assertEqual(report["success"], minor >= 23, report)
        self.assertEqual(snapshot(source), before)
        if minor < 23:
            write(source, "compat.go", "//go:build !go1.23\n\npackage variant\nfunc Value() int { return 3 }\n")
            before = snapshot(source)
            report = self.checker.check_candidate("toolchain-variants", source, GO, probes_dir=probes)
            self.assertTrue(report["success"], report)
            self.assertEqual(snapshot(source), before)

    def release_fixture(self):
        source = self.root / "candidate"
        write(source, "go.mod", "module example.org/release\ngo 1.22.0\n")
        write(source, ".gitignore", "/preview/\n/dist/\n")
        write(source, "schema.txt", "object\n")
        write(source, "release.go", 'package release\nimport _ "embed"\n//go:embed preview/schema.json\nvar data string\nfunc Schema() string { return data }\n')
        write(source, "release_test.go", 'package release\nimport "testing"\nfunc TestAuthored(t *testing.T) { if Schema() == "" { t.Fatal("empty") } }\n')
        write(source, "cmd/codes/main.go", 'package main\nimport "example.org/release"\nfunc main() { if release.Schema() == "" { panic("empty") } }\n')
        write(source, "Makefile", 'GO ?= go\npreview:\n\tmkdir -p preview\n\tcp schema.txt preview/schema.json\nsource:\n\tmkdir -p dist\n\tgit ls-files --cached --others --exclude-standard -z | tar --null -T - -czf dist/codes-src.tar.gz\n')
        probes = self.root / "probes"
        write(probes, "release-inputs_test.go.txt", 'package release\nimport "testing"\nfunc TestProbe(t *testing.T) { if Schema() != "object\\n" { t.Fatal(Schema()) } }\n')
        return source, probes

    def git(self, source, *arguments):
        result = subprocess.run(["git", "-C", str(source), *arguments], capture_output=True,
                                text=True, timeout=30,
                                env={"PATH": os.environ["PATH"], "HOME": str(self.root),
                                     "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.devnull})
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout

    def test_ignored_warm_asset_is_not_a_source_release_input(self):
        # Passing warm embed tests must not conceal an absent source archive input.
        source, probes = self.release_fixture()
        write(source, ".gitignore", "/dist/\n")
        write(source, "preview/schema.json", "object\n")
        self.git(source, "init", "-q")
        write(source, ".git/info/exclude", "/preview/\n")
        before = snapshot(source)
        report = self.checker.check_candidate("release-inputs", source, GO, probes_dir=probes)
        checks = {check["name"]: check for check in report["checks"]}
        self.assertEqual(checks["warm-tests"]["exit_code"], 0, checks)
        self.assertNotEqual(checks["export-tests"]["exit_code"], 0, checks)
        self.assertFalse(report["success"])
        self.assertEqual(snapshot(source), before)

    def test_force_added_ignored_asset_is_a_source_release_input(self):
        # Explicitly staging an ignored input must survive the copied export index.
        source, probes = self.release_fixture()
        write(source, "preview/schema.json", "object\n")
        self.git(source, "init", "-q")
        self.git(source, "add", ".")
        self.git(source, "add", "-f", "preview/schema.json")
        before = snapshot(source)
        index = (source / ".git/index").read_bytes()
        report = self.checker.check_candidate("release-inputs", source, GO, probes_dir=probes)
        self.assertTrue(report["success"], report)
        self.assertEqual(snapshot(source), before)
        self.assertEqual((source / ".git/index").read_bytes(), index)

    def test_parent_repository_does_not_supply_the_source_index(self):
        # An ancestor repository's force-added assets do not define this export's inventory.
        source, probes = self.release_fixture()
        write(source, "preview/schema.json", "object\n")
        self.git(self.root, "init", "-q")
        self.git(self.root, "add", "candidate")
        self.git(self.root, "add", "-f", "candidate/preview/schema.json")
        before = snapshot(source)
        report = self.checker.check_candidate("release-inputs", source, GO, probes_dir=probes)
        checks = {check["name"]: check for check in report["checks"]}
        self.assertEqual(checks["warm-tests"]["exit_code"], 0, checks)
        self.assertNotEqual(checks["export-tests"]["exit_code"], 0, checks)
        self.assertFalse(report["success"])
        self.assertEqual(snapshot(source), before)

    def test_existing_authored_test_remains_effective_with_probe(self):
        # Overwriting zz_contract_test.go would hide the intentional failure.
        source = self.root / "candidate"
        write(source, "go.mod", "module example.org/control\ngo 1.22.0\n")
        write(source, "value.go", "package control\nfunc Value() int { return 3 }\n")
        write(source, "zz_contract_test.go", 'package control\nimport "testing"\nfunc TestAuthored(t *testing.T) { t.Fatal("authored failure") }\n')
        probes = self.root / "probes"
        write(probes, "pure-control_test.go.txt", 'package control\nimport "testing"\nfunc TestProbe(t *testing.T) { if Value() != 3 { t.Fatal(Value()) } }\n')
        before = snapshot(source)
        with patch.dict(os.environ, {"GOFLAGS": "-invalid-user-flag", "GOROOT": "/missing-go",
                                     "GOWORK": "/missing-workspace", "GOTOOLCHAIN": "go9.0.0"}):
            report = self.checker.check_candidate("pure-control", source, GO, probes_dir=probes)
        self.assertFalse(report["success"])
        self.assertIn("authored failure", "".join(check["stdout"] for check in report["checks"]))
        self.assertEqual(snapshot(source), before)

    def test_supplied_workspace_probe_requires_wrapped_error_cause(self):
        # Returning ErrName directly must fail even though errors.Is matches its cause.
        source = self.root / "candidate"
        fixture = CHECK_PATH.parent / "evals/files/workspace-library"
        shutil.copytree(fixture, source)
        write(source, "lib/go.mod", "module example.org/catalog\ngo 1.22.0\nrequire example.org/keytext v1.1.0\n")
        write(source, "lib/canonical.go", 'package catalog\nimport "example.org/keytext"\nfunc CanonicalName(name string) (string, error) { return keytext.Canonical(name) }\n')
        prep = self.root / "prepare"
        prep.mkdir()
        proxy = self.checker.build_proxy(source / "registry", prep / "proxy")
        env = self.checker.environment(str(Path(GO).resolve()), prep, proxy)
        result = subprocess.run([GO, "mod", "tidy"], cwd=source / "lib", env=env,
                                capture_output=True, text=True, timeout=120)
        self.assertEqual(result.returncode, 0, result.stderr)
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-library", source, GO)
        self.assertFalse(report["success"], "supplied workspace probe accepted a bare sentinel")
        failures = {check["name"] for check in report["checks"] if check["exit_code"] != 0}
        self.assertEqual(failures, {"standalone-tests"})
        self.assertIn("TestContractCanonicalName", next(check["stdout"] for check in report["checks"]
                                                       if check["name"] == "standalone-tests"))
        self.assertEqual(snapshot(source), before)

        write(source, "lib/canonical.go", 'package catalog\nimport ("example.org/keytext"; "fmt")\nfunc CanonicalName(name string) (string, error) { value, err := keytext.Canonical(name); if err != nil { return "", fmt.Errorf("canonical name: %w", err) }; return value, nil }\n')
        before = snapshot(source)
        report = self.checker.check_candidate("workspace-library", source, GO)
        self.assertTrue(report["success"], report)
        self.assertEqual(snapshot(source), before)

    def generated_fixture(self, *, stale=False, hardcoded=False):
        source = self.root / "candidate"
        write(source, "go.mod", "module example.org/generated\ngo 1.22.0\n")
        write(source, "generate.go", "package generated\n//go:generate go run ./cmd/gen\n")
        write(source, "schema.txt", "10 debug\n20 info\n30 warn\n40 error\n")
        cases = 'case 10: return "debug"; case 20: return "info";'
        if not stale:
            cases += ' case 30: return "warn"; case 40: return "error";'
        for filename, tag in (("normal_gen.go", "!compact"), ("compact_gen.go", "compact")):
            write(source, filename, f'//go:build {tag}\n\npackage generated\nfunc Name(n int) string {{ switch n {{ {cases} }}; return "unknown" }}\n')
        # Authored tests legitimately bind 50 to the original schema's unknown result.
        write(source, "zz_contract_test.go", 'package generated\nimport "testing"\nfunc TestAuthored(t *testing.T) { if Name(10) != "debug" || Name(20) != "info" || Name(50) != "unknown" { t.Fatal("original schema changed") } }\n')
        generator = '''package main
import ("fmt"; "os"; "strings")
func main() {
    schema, err := os.ReadFile("schema.txt")
    if err != nil { panic(err) }
    body := "package generated\\nfunc Name(n int) string { switch n {\\n"
    for _, line := range strings.Split(strings.TrimSuffix(string(schema), "\\n"), "\\n") {
        var code int
        var name string
        if _, err := fmt.Sscan(line, &code, &name); err != nil { panic(err) }
        body += fmt.Sprintf("case %d: return %q;\\n", code, name)
    }
    body += "}; return \\"unknown\\" }\\n"
    for filename, tag := range map[string]string{"normal_gen.go": "!compact", "compact_gen.go": "compact"} {
        data := "//go:build " + tag + "\\n\\n" + body
        if err := os.WriteFile(filename, []byte(data), 0644); err != nil { panic(err) }
    }
}
'''
        if hardcoded:
            generator = generator.replace('schema, err := os.ReadFile("schema.txt")\n    if err != nil { panic(err) }',
                                          'schema := []byte("10 debug\\n20 info\\n30 warn\\n40 error\\n")')
        write(source, "cmd/gen/main.go", generator)
        probes = self.root / "probes"
        write(probes, "generated-targets_test.go.txt", 'package generated\nimport "testing"\nfunc TestProbe(t *testing.T) { if Name(40) != "error" { t.Fatalf("Name(40) = %q, want error", Name(40)) } }\n')
        write(probes, "generated-targets-schema_test.go.txt", 'package generated\nimport "testing"\nfunc TestContractSchemaSource(t *testing.T) { if Name(50) != "fatal" { t.Fatalf("Name(50) = %q, want fatal", Name(50)) } }\n')
        return source, probes, generator

    def test_stale_committed_tables_fail_before_regeneration(self):
        # Regeneration must not hide stale tables that compile and pass authored tests.
        source, probes, _ = self.generated_fixture(stale=True)
        before = snapshot(source)
        report = self.checker.check_candidate("generated-targets", source, GO, probes_dir=probes)
        self.assertFalse(report["success"], "stale committed tables were accepted")
        failures = {check["name"] for check in report["checks"] if check["exit_code"] != 0}
        self.assertEqual(failures, {"initial-default-tests", "initial-compact-tests"})
        self.assertTrue(all('Name(40) = "unknown", want error' in check["stdout"]
                            for check in report["checks"] if check["name"] in failures))
        self.assertEqual(snapshot(source), before)

    def test_schema_ignoring_generator_fails_changed_schema_in_both_tags(self):
        # Hardcoded regeneration must not earn credit merely for reproducing the initial table.
        source, probes, _ = self.generated_fixture(hardcoded=True)
        before = snapshot(source)
        report = self.checker.check_candidate("generated-targets", source, GO, probes_dir=probes)
        self.assertFalse(report["success"], "schema-ignoring generator was accepted")
        failures = {check["name"] for check in report["checks"] if check["exit_code"] != 0}
        self.assertEqual(failures, {"schema-default-tests", "schema-compact-tests"})
        self.assertTrue(all('Name(50) = "unknown", want fatal' in check["stdout"]
                            for check in report["checks"] if check["name"] in failures))
        self.assertEqual(snapshot(source), before)

    def test_generation_rebuilds_both_tags_and_rejects_changing_output(self):
        # Skipping clean generation, schema consumption, compact, or equality would accept a broken generator.
        source, probes, generator = self.generated_fixture()
        before = snapshot(source)
        report = self.checker.check_candidate("generated-targets", source, GO, probes_dir=probes)
        self.assertTrue(report["success"], [(check["name"], check["stdout"], check["stderr"])
                                          for check in report["checks"] if check["exit_code"] != 0])
        self.assertEqual(snapshot(source), before)
        cross_builds = [check for check in report["checks"]
                        if check["name"].startswith("compile-only-")]
        self.assertEqual(len(cross_builds), 4)
        self.assertEqual({check["env"]["GOARCH"] for check in cross_builds}, {"amd64", "arm64"})
        self.assertTrue(all(check["env"]["GOOS"] == "linux" and
                            "-c" in check["command"] for check in cross_builds))

        # A changing comment still compiles, so failure specifically requires byte comparison.
        write(source, "cmd/gen/main.go", generator.replace('"strings")', '"strings"; "time")').replace(
            '[]byte(data)', '[]byte(data + "// " + time.Now().String() + "\\n")'))
        before = snapshot(source)
        report = self.checker.check_candidate("generated-targets", source, GO, probes_dir=probes)
        self.assertFalse(report["success"])
        failures = {check["name"] for check in report["checks"] if check["exit_code"] != 0}
        self.assertEqual(failures, {"generation-byte-identical"})
        self.assertEqual(snapshot(source), before)


if __name__ == "__main__":
    unittest.main()
