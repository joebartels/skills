"""Exercise the disposable module/build checker with real local Go modules."""

import importlib.util
import os
from pathlib import Path
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
            for path in root.rglob("*") if path.is_file()}


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
