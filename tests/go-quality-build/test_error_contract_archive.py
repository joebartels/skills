"""Verify the historical archive without the original checkout or host files."""

import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch


class ErrorContractArchiveTests(unittest.TestCase):
    def setUp(self):
        self.script = Path(__file__).with_name("verify_error_contract_archive.py")
        spec = importlib.util.spec_from_file_location("archive_verifier", self.script)
        self.verifier = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(self.verifier)
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        self.archive = self.root / "archive"
        shutil.copytree(Path(__file__).parent / "results/2026-10-03-error-contracts-r2",
                        self.archive)

    def test_archive_is_sufficient_and_cli_preserves_recorded_results(self):
        before = self.verifier.inventory(self.archive)
        read_bytes = Path.read_bytes

        def archived_bytes(path):
            self.assertTrue(path.resolve().is_relative_to(self.root.resolve()),
                            f"read unarchived input: {path}")
            return read_bytes(path)

        with patch.object(tempfile, "tempdir", str(self.root)), \
                patch.object(Path, "read_bytes", archived_bytes):
            result = self.verifier.verify(self.archive)
        self.assertEqual(result["status"], "pass")
        self.assertEqual(len(result["completed_baseline_samples"]), 8)
        self.assertEqual(len(result["skipped_guided_samples"]), 8)
        command = subprocess.run(
            [sys.executable, "-B", str(self.script), str(self.archive)],
            cwd=self.root, capture_output=True, text=True, timeout=30)
        self.assertEqual(command.returncode, 0, command.stdout + command.stderr)
        self.assertEqual(json.loads(command.stdout)["status"], "pass")
        self.assertEqual(self.verifier.inventory(self.archive), before)

    def test_modified_archived_source_is_rejected(self):
        source = next(self.archive.glob("samples/*/candidate/*.go"))
        source.write_text(source.read_text() + "\n// changed after sealing\n")
        with self.assertRaises(AssertionError):
            self.verifier.verify(self.archive)

    def test_optimized_cli_cannot_report_false_success(self):
        source = next(self.archive.glob("samples/*/candidate/*.go"))
        source.write_text(source.read_text() + "\n// changed after sealing\n")
        command = subprocess.run(
            [sys.executable, "-O", "-B", str(self.script), str(self.archive)],
            cwd=self.root, capture_output=True, text=True, timeout=30)
        self.assertNotEqual(command.returncode, 0, command.stdout)

    def test_cli_requires_only_git_on_path(self):
        binaries = self.root / "bin"
        binaries.mkdir()
        (binaries / "git").symlink_to(shutil.which("git"))
        command = subprocess.run(
            [sys.executable, "-B", str(self.script), str(self.archive)],
            cwd=self.root, env={**os.environ, "PATH": str(binaries)},
            capture_output=True, text=True, timeout=30)
        self.assertEqual(command.returncode, 0, command.stdout + command.stderr)
        self.assertEqual(json.loads(command.stdout)["status"], "pass")


if __name__ == "__main__":
    unittest.main()
