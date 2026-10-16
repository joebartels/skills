"""Trial isolation and exact patch reconstruction, using real temporary repositories."""

import importlib.util
from contextlib import chdir, contextmanager
import hashlib
import json
import os
from pathlib import Path
import shutil
import sys
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

    def test_guided_catalog_survives_runtime_promotion(self):
        runtime = self.trials.REPO / "plugins/go-quality-build/skills/go-client-calls"
        runtime.mkdir()
        (runtime / "SKILL.md").write_text(
            "---\nname: go-client-calls\ndescription: Use when calling.\n---\n")
        references = runtime / "references"
        references.mkdir()
        (references / "http-clients.md").write_text("Promoted HTTP guidance\n")
        guided = self.prepare("skill-on")
        self.assertEqual(self.trials.hashes(guided / "catalog/go-client-calls"),
                         self.trials.hashes(runtime))
        draft = self.evals.parent / "draft"
        draft.mkdir()
        (draft / "SKILL.md").write_text(
            "---\nname: go-client-calls\ndescription: Use when revising calls.\n---\n")
        revised = self.prepare("skill-on", "t02")
        self.assertEqual(self.trials.hashes(revised / "catalog/go-client-calls"),
                         self.trials.hashes(draft))
        baseline = self.prepare("baseline", "t03")
        self.assertFalse((baseline / "catalog/go-client-calls").exists())

    def test_refuses_archive_overwrite(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        self.trials.archive_trial(trial, self.archive)
        with self.assertRaises(FileExistsError):
            self.trials.archive_trial(trial, self.archive)

    def test_prepare_and_archive_need_only_git_on_path(self):
        git = shutil.which("git")
        binary = self.root / "bin"
        binary.mkdir()
        (binary / "git").symlink_to(git)
        with patch.dict("os.environ", {"PATH": str(binary)}):
            trial = self.prepare()
            (trial / "trial-report.md").write_text("Result\n")
            (trial / "source/client.go").write_text("package client\nvar Value = 7\n")
            self.trials.archive_trial(trial, self.archive)
        self.assertEqual((self.archive / "t01/output/client.go").read_text(),
                         "package client\nvar Value = 7\n")

    def test_detects_catalog_mutation(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        next((trial / "catalog").glob("*/SKILL.md")).write_text("mutation")
        with self.assertRaisesRegex(ValueError, "catalog"):
            self.trials.archive_trial(trial, self.archive)
        self.assertFalse((self.archive / "t01").exists())

    def test_prompt_hash_uses_written_newlines(self):
        write_text = Path.write_text

        def translated_write(path, text, *args, **kwargs):
            if path.name == "prompt.txt":
                kwargs["newline"] = "\r\n"
            return write_text(path, text, *args, **kwargs)

        with patch.object(Path, "write_text", translated_write):
            trial = self.prepare()
        self.assertIn(b"\r\n", (trial / "prompt.txt").read_bytes())
        (trial / "trial-report.md").write_text("Result\n")
        self.trials.archive_trial(trial, self.archive)

    def test_patch_preserves_crlf_file_bytes(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        self.trials.run(["git", "config", "core.autocrlf", "false"], trial / "source")
        content = b"first\r\nsecond\r\n"
        (trial / "source/new.txt").write_bytes(content)
        result = self.trials.archive_trial(trial, self.archive)
        self.assertEqual((self.archive / "t01/output/new.txt").read_bytes(), content)
        patch_bytes = (self.archive / "t01/source.patch").read_bytes()
        self.assertEqual(result["patch_sha256"], hashlib.sha256(patch_bytes).hexdigest())

    def test_relative_archive_patch_path(self):
        self.archive = Path("archive")
        temporary_directory = tempfile.TemporaryDirectory

        @contextmanager
        def relative_temporary(*args, **kwargs):
            # Python 3.11 preserves a relative TemporaryDirectory dir argument.
            with temporary_directory(*args, **kwargs) as directory:
                yield os.path.relpath(directory) if kwargs.get("dir") else directory

        with chdir(self.root):
            trial = self.prepare()
            (trial / "trial-report.md").write_text("Result\n")
            (trial / "source/client.go").write_text("package client\nvar Value = 7\n")
            with patch.object(tempfile, "TemporaryDirectory", relative_temporary):
                self.trials.archive_trial(trial, self.archive)
        self.assertEqual((self.root / "archive/t01/output/client.go").read_text(),
                         "package client\nvar Value = 7\n")

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

    def test_reconstructs_when_temporary_directory_is_inside_git(self):
        self.trials.run(["git", "init", "-q"], self.trials.REPO)
        temporary = self.trials.REPO / "temporary"
        temporary.mkdir()
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        (trial / "source/client.go").unlink()
        (trial / "source/new.go").write_text("package client\nvar Value = 7\n")
        with patch.object(tempfile, "tempdir", str(temporary)):
            self.trials.archive_trial(trial, self.archive)
        output = self.archive / "t01/output"
        self.assertEqual({p.name for p in output.iterdir()}, {"README.md", "new.go"})
        self.assertEqual((output / "README.md").read_text(), "Task contract\n")
        self.assertEqual((output / "new.go").read_text(), "package client\nvar Value = 7\n")

    def test_cli_mismatch_cannot_seal_an_archive(self):
        trial = self.prepare()
        (trial / "trial-report.md").write_text("Result\n")
        result = subprocess.run(
            [sys.executable, "-B", str(Path(__file__).with_name("client_call_trials.py")),
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

    @unittest.skipUnless(shutil.which("go"), "candidate checker requires Go 1.22+")
    def test_checker_enforces_private_probes_without_mutating_source(self):
        suite = Path(__file__).parent / "go-client-calls"
        source = self.root / "candidate"
        shutil.copytree(suite / "evals/files/pure-control", source)
        command = [sys.executable, "-B", str(suite / "check.py"), "pure-control", str(source)]
        before = {p.name: p.read_bytes() for p in source.iterdir()}
        broken = subprocess.run(command, text=True, capture_output=True)
        self.assertNotEqual(broken.returncode, 0)
        self.assertIn("TestOnlyPositiveContributes", broken.stdout)
        self.assertEqual({p.name: p.read_bytes() for p in source.iterdir()}, before)
        (source / "client.go").write_text(
            "package client\nfunc SumPositive(values []int) int {\n"
            "sum := 0; for _, v := range values { if v > 0 { sum += v } }; return sum\n}\n")
        before["client.go"] = (source / "client.go").read_bytes()
        fixed = subprocess.run(command, text=True, capture_output=True)
        self.assertEqual(fixed.returncode, 0, fixed.stdout + fixed.stderr)
        self.assertEqual({p.name: p.read_bytes() for p in source.iterdir()}, before)
        authored_test = source / "zz_contract_0_test.go"
        authored_test.write_text(
            'package client\nimport "testing"\n'
            'func TestAuthorRegression(t *testing.T) { t.Fatal("author test must run") }\n')
        before[authored_test.name] = authored_test.read_bytes()
        failed = subprocess.run(command, text=True, capture_output=True)
        self.assertNotEqual(failed.returncode, 0)
        self.assertIn("TestAuthorRegression", failed.stdout)
        self.assertEqual({p.name: p.read_bytes() for p in source.iterdir()}, before)


if __name__ == "__main__":
    unittest.main()
