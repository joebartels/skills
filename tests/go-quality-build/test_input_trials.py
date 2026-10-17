"""Exercise trial isolation and checker exit status using disposable inputs."""

import json
from pathlib import Path
import subprocess
import tempfile
import unittest


TOOL = Path(__file__).with_name("go-input-and-serialization") / "trials.py"


class InputTrialsTests(unittest.TestCase):
    def invoke(self, *arguments):
        return subprocess.run(["python3", "-B", str(TOOL), *map(str, arguments)],
                              capture_output=True, text=True)

    def test_rejects_generated_output_inside_repository(self):
        result = self.invoke("prepare", "pure-control", "baseline", TOOL.parent / "generated")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("outside the repository", result.stderr)

    def test_prepares_independent_frozen_inputs_without_oracles(self):
        with tempfile.TemporaryDirectory() as temporary:
            trial = Path(temporary) / "t01"
            result = self.invoke("prepare", "pure-control", "baseline", trial)
            self.assertEqual(result.returncode, 0, result.stderr)
            manifest = json.loads((trial / "manifest.json").read_text())
            self.assertTrue(manifest["input_sha256"])
            self.assertTrue(manifest["catalog_sha256"])
            self.assertNotIn("go-input-and-serialization/SKILL.md", manifest["catalog_sha256"])
            self.assertNotIn("expected_output", (trial / "prompt.txt").read_text())
            original = (trial / "input" / "README.md").read_bytes()
            (trial / "source" / "README.md").write_text("candidate changed documentation\n")
            self.assertEqual((trial / "input" / "README.md").read_bytes(), original)
            again = self.invoke("prepare", "pure-control", "baseline", trial)
            self.assertNotEqual(again.returncode, 0)

    def test_guided_trial_uses_only_supplied_guidance(self):
        with tempfile.TemporaryDirectory() as temporary:
            temporary = Path(temporary)
            skill = temporary / "guidance"
            skill.mkdir()
            (skill / "SKILL.md").write_text(
                "---\nname: go-input-and-serialization\n"
                "description: Use when decoding external data.\n---\nGuidance.\n")
            trial = temporary / "t02"
            result = self.invoke("prepare", "pure-control", "skill-on", trial, "--skill", skill)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual((trial / "catalog/go-input-and-serialization/SKILL.md").read_bytes(),
                             (skill / "SKILL.md").read_bytes())
            manifest = json.loads((trial / "manifest.json").read_text())
            self.assertIn("go-input-and-serialization/SKILL.md", manifest["catalog_sha256"])

    def test_candidate_failure_produces_nonzero_checker_exit(self):
        with tempfile.TemporaryDirectory() as temporary:
            source = Path(temporary)
            (source / "go.mod").write_text("module example.com/failure\n\ngo 1.22\n")
            (source / "failure_test.go").write_text(
                'package failure\nimport "testing"\n'
                'func TestCandidateFailure(t *testing.T) { t.Fatal("candidate failure") }\n')
            result = self.invoke("check", "pure-control", source)
            self.assertNotEqual(result.returncode, 0)
            report = json.loads(result.stdout)
            self.assertFalse(report["pass"])
            self.assertNotEqual(report["candidate_tests"]["exit_code"], 0)
            self.assertIn("TestCandidateFailure", report["candidate_tests"]["output"])

    def test_zero_exit_without_probe_execution_is_not_a_pass(self):
        with tempfile.TemporaryDirectory() as temporary:
            temporary = Path(temporary)
            source = temporary / "source"
            source.mkdir()
            executable = temporary / "empty-go"
            executable.write_text("#!/bin/sh\nexit 0\n")
            executable.chmod(0o755)
            result = self.invoke("check", "pure-control", source, "--go", executable)
            self.assertNotEqual(result.returncode, 0)
            self.assertFalse(json.loads(result.stdout)["pass"])

    def test_private_probes_have_an_independent_consumer_namespace(self):
        with tempfile.TemporaryDirectory() as temporary:
            source = Path(temporary)
            (source / "go.mod").write_text("module example.com/oddtally\n\ngo 1.22\n")
            (source / "tally.go").write_text(
                "package tally\nfunc Total(values []int) int {\n"
                "total := 0; for _, v := range values { if v%2 != 0 { total += v } }; return total\n}\n")
            (source / "own_test.go").write_text(
                'package tally_test\nimport "testing"\n'
                'func TestNegativeOddValuesContribute(t *testing.T) {}\n')
            result = self.invoke("check", "pure-control", source)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertTrue(json.loads(result.stdout)["pass"])


if __name__ == "__main__":
    unittest.main()
