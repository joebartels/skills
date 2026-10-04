"""Trial isolation and exact patch reconstruction, using real temporary repositories."""

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
import subprocess
from unittest.mock import patch


class ClientCallTrialsTests(unittest.TestCase):
    def setUp(self):
        spec = importlib.util.spec_from_file_location(
            "trials", Path(__file__).with_name("client_call_trials.py"))
        self.trials = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(self.trials)
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        repo = self.root / "repo"
        self.patcher = patch.object(self.trials, "REPO", repo)
        self.patcher.start()
        self.addCleanup(self.patcher.stop)
        for name in self.trials.CATALOG:
            skill = repo / "plugins/go-quality-build/skills" / name / "SKILL.md"
            skill.parent.mkdir(parents=True)
            skill.write_text(f"---\nname: {name}\ndescription: Use when relevant.\n---\n")
        self.evals = repo / "tests/go-quality-build/go-client-calls/evals"
        source = self.evals / "files/one"
        source.mkdir(parents=True)
        (source / "client.go").write_text("package client\n")
        (source / "README.md").write_text("Task contract\n")
        suite = {"evals": [{"id": "one", "prompt": "Implement the contract.",
                            "files": ["files/one/client.go", "files/one/README.md"]}]}
        (self.evals / "evals.json").write_text(json.dumps(suite))
        self.scratch = self.root / "scratch"
        self.archive = self.root / "archive"

    def prepare(self, arm="baseline", trial_id="t01"):
        return self.trials.prepare_trial("one", arm, trial_id, self.scratch, self.archive)

    def test_prepare_freezes_exact_catalog(self):
        trial = self.prepare()
        self.assertEqual({p.name for p in (trial / "catalog").iterdir()},
                         set(self.trials.CATALOG))
        original = self.trials.hashes(trial / "catalog")
        skill = self.trials.REPO / "plugins/go-quality-build/skills" / self.trials.CATALOG[0] / "SKILL.md"
        skill.write_text("description: Use when changed after preparation.\n")
        self.assertEqual(self.trials.hashes(trial / "catalog"), original)
        draft = self.evals.parent / "draft/SKILL.md"
        draft.parent.mkdir()
        draft.write_text("---\nname: go-client-calls\ndescription: Use when calling.\n---\n")
        guided = self.prepare("skill-on", "t02")
        self.assertEqual({p.name for p in (guided / "catalog").iterdir()},
                         set(self.trials.CATALOG) | {"go-client-calls"})

    def test_rejects_path_escape(self):
        for case, arm, trial_id in [("../one", "baseline", "t01"),
                                    ("one", "unexpected", "t01"),
                                    ("one", "baseline", "../out")]:
            with self.subTest(case=case, arm=arm, trial_id=trial_id):
                with self.assertRaises(ValueError):
                    self.trials.prepare_trial(case, arm, trial_id, self.scratch, self.archive)
        suite = json.loads((self.evals / "evals.json").read_text())
        suite["evals"][0]["files"] = ["../outside.go"]
        (self.evals / "evals.json").write_text(json.dumps(suite))
        with self.assertRaises(ValueError):
            self.prepare()

    def test_refuses_archive_overwrite(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        self.trials.archive_trial(trial, self.archive)
        with self.assertRaises(FileExistsError):
            self.trials.archive_trial(trial, self.archive)

    def test_detects_catalog_mutation(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        next((trial / "catalog").glob("*/SKILL.md")).write_text("mutation")
        with self.assertRaisesRegex(ValueError, "catalog"):
            self.trials.archive_trial(trial, self.archive)
        self.assertFalse((self.archive / "t01").exists())

    def test_patch_reconstructs_added_and_deleted_files(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        (trial / "source/client.go").unlink()
        (trial / "source/new.go").write_text("package client\nvar Value = 7\n")
        result = self.trials.archive_trial(trial, self.archive)
        self.assertEqual(result["output_hashes"],
                         self.trials.hashes(self.archive / "t01/output"))
        self.assertFalse((self.archive / "t01/output/client.go").exists())
        self.assertEqual((self.archive / "t01/output/new.go").read_text(),
                         "package client\nvar Value = 7\n")

    def test_cli_mismatch_cannot_seal_an_archive(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        result = subprocess.run(
            ["rtk", "proxy", "python3", "-B", str(Path(__file__).with_name("client_call_trials.py")),
             "archive", "--case", "wrong", "--arm", "baseline", "--trial-id", "t01",
             "--scratch", str(self.scratch), "--archive", str(self.archive)],
            capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse((self.archive / "t01").exists(),
                         "a mismatched CLI invocation sealed an output before rejecting it")

    def test_archive_reconstructs_inside_a_parent_repository(self):
        self.trials.run(["git", "init", "-q"], self.trials.REPO)
        self.archive = self.trials.REPO / "results"
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        (trial / "source/client.go").write_text("package client\nvar Changed = true\n")
        self.trials.archive_trial(trial, self.archive)
        self.assertEqual((self.archive / "t01/output/client.go").read_text(),
                         "package client\nvar Changed = true\n")


if __name__ == "__main__":
    unittest.main()
