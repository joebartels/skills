"""CLI regression checks for the current Go study evidence gate."""
import os
from pathlib import Path
import subprocess
import sys
import unittest


ROOT = Path(__file__).resolve().parents[1]
GATE = ROOT / "tests/go-quality-build/results/2026-10-02-context-concurrency-delivery/verify-delivery.py"


class GoStudyIntegrityCLI(unittest.TestCase):
    def run_gate(self, flags=(), optimize=None):
        env = os.environ.copy()
        env.pop("PYTHONOPTIMIZE", None)
        if optimize is not None:
            env["PYTHONOPTIMIZE"] = optimize
        return subprocess.run(
            [sys.executable, "-B", *flags, str(GATE)],
            cwd=ROOT, env=env, capture_output=True, text=True, timeout=30,
        )

    def test_normal_execution_verifies_current_evidence(self):
        result = self.run_gate()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("PASS:", result.stdout)

    def test_optimized_modes_fail_closed(self):
        for flags, optimize in [(('-O',), None), (('-OO',), None), ((), '1')]:
            with self.subTest(flags=flags, optimize=optimize):
                result = self.run_gate(flags, optimize)
                self.assertNotEqual(result.returncode, 0, result.stdout)
                self.assertNotIn("PASS:", result.stdout)


if __name__ == "__main__":
    unittest.main()
