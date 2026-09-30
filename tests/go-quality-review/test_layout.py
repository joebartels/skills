"""Check that the Go review collection has one runtime skill tree."""

from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
SKILLS = ROOT / "plugins/go-quality-review/skills"
EVALUATIONS = ROOT / "tests/go-quality-review"


class GoQualityLayoutTests(unittest.TestCase):
    def test_runtime_and_evaluations_are_separated(self):
        names = {
            "go-architecture-and-design",
            "go-code-quality-and-idioms",
            "go-correctness-and-compatibility",
            "go-dependencies-and-reproducibility",
            "go-deployment-and-operations",
            "go-observability-and-resilience",
            "go-performance-and-resource-management",
            "go-quality-report",
            "go-security",
            "go-testing",
        }
        self.assertEqual({path.parent.name for path in SKILLS.glob("go-*/SKILL.md")}, names)
        for name in names:
            with self.subTest(skill=name):
                self.assertTrue((EVALUATIONS / name / "evals/evals.json").is_file())
                self.assertFalse((SKILLS / name / "evals").exists())


if __name__ == "__main__":
    unittest.main()
