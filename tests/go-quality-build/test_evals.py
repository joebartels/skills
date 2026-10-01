"""Build evaluation validation rejects unusable or escaping inputs."""

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

SPEC = importlib.util.spec_from_file_location(
    "validator", Path(__file__).resolve().parents[2] / "scripts/validate.py"
)
validator = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(validator)


class BuildEvalsTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        skill = self.root / "plugins/go-quality-build/skills/go-example/SKILL.md"
        skill.parent.mkdir(parents=True)
        skill.write_text("example")
        self.evals = self.root / "tests/go-quality-build/go-example/evals"
        self.evals.mkdir(parents=True)
        (self.evals / "input.go").write_text("package example\n")
        self.suite = {"skill_name": "go-example", "evals": [{
            "id": "case-one", "prompt": "Move a responsibility.",
            "expected_output": "A useful boundary.",
            "assertions": ["Consumers retain their behavior."],
            "files": ["input.go"],
        }]}

    def validate(self):
        (self.evals / "evals.json").write_text(json.dumps(self.suite))
        self.assertTrue(callable(getattr(validator, "validate_build_evals", None)),
                        "validate_build_evals is missing")
        return validator.validate_build_evals(self.root)

    def test_valid_suite(self):
        self.assertEqual(self.validate(), ([], 1))

    def test_candidate_suite_without_runtime_skill(self):
        self.validate()
        (self.root / "plugins/go-quality-build/skills/go-example/SKILL.md").unlink()
        self.assertEqual(validator.validate_build_evals(self.root), ([], 1))
        self.suite["evals"][0]["assertions"] = []
        self.assertTrue(self.validate()[0])

    def test_combined_suite_is_required_and_uses_case_rules(self):
        self.validate()
        combined = self.root / "tests/go-quality-build/combined/evals"
        combined.mkdir(parents=True)
        errors, count = validator.validate_build_evals(self.root)
        self.assertTrue(any("combined: invalid evaluation file" in error for error in errors))
        self.assertEqual(count, 1)

        (combined / "input.go").write_text("package combined\n")
        suite = dict(self.suite, skill_name="combined")
        (combined / "evals.json").write_text(json.dumps(suite))
        self.assertEqual(validator.validate_build_evals(self.root), ([], 2))

        suite["evals"][0]["files"] = ["../outside.go"]
        (combined / "evals.json").write_text(json.dumps(suite))
        errors, count = validator.validate_build_evals(self.root)
        self.assertTrue(any("combined/case-one: missing or non-portable fixture" in error
                            for error in errors))
        self.assertEqual(count, 2)

    def test_missing_suite(self):
        self.validate()
        (self.evals / "evals.json").unlink()
        errors, count = validator.validate_build_evals(self.root)
        self.assertTrue(errors)
        self.assertEqual(count, 0)

    def test_duplicate_ids(self):
        self.suite["evals"].append(dict(self.suite["evals"][0]))
        self.assertTrue(self.validate()[0])

    def test_empty_assertions(self):
        self.suite["evals"][0]["assertions"] = []
        self.assertTrue(self.validate()[0])

    def test_fixture_escape(self):
        (self.evals.parent / "outside.go").write_text("package outside\n")
        self.suite["evals"][0]["files"] = ["../outside.go"]
        self.assertTrue(self.validate()[0])

    def test_empty_cases(self):
        self.suite["evals"] = []
        self.assertTrue(self.validate()[0])

    def test_empty_prompt_or_expectation(self):
        for field in ("prompt", "expected_output"):
            with self.subTest(field=field):
                self.suite["evals"][0][field] = " "
                self.assertTrue(self.validate()[0])
                self.suite["evals"][0][field] = "valid"

    def test_wrong_case_shapes(self):
        for cases in ([None], [{"id": []}], []):
            with self.subTest(cases=cases):
                self.suite["evals"] = cases
                self.assertTrue(self.validate()[0])


if __name__ == "__main__":
    unittest.main()
