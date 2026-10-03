"""Recovery replay must reject guidance drift before running any Go trial."""

import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[1]


class RecoveryIdentityTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        self.here = self.root / "tests/go-quality-build/recovery"
        shutil.copytree(ROOT / "plugins/go-quality-build/skills",
                        self.root / "plugins/go-quality-build/skills")
        self.here.mkdir(parents=True)
        for name in ("check.py", "trials.json"):
            shutil.copyfile(ROOT / "tests/go-quality-build/recovery" / name,
                            self.here / name)
        shutil.copytree(ROOT / "tests/go-quality-build/recovery/patches",
                        self.here / "patches")
        for probe in self.manifest()["probe_sha256"]:
            target = self.root / probe
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(ROOT / probe, target)

    def manifest(self):
        return json.loads((self.here / "trials.json").read_text())

    def save_manifest(self, manifest):
        (self.here / "trials.json").write_text(json.dumps(manifest))

    def preflight(self, optimized=False):
        # An absent selection ends after identity checks, without executing Go.
        command = [sys.executable]
        if optimized:
            command.append("-O")
        return subprocess.run(
            command + ["-B", str(self.here / "check.py"),
             "--trial", "not-a-recorded-trial", "--output", str(self.root / "result.json")],
            capture_output=True, text=True, timeout=15,
        )

    def assert_rejected(self, identity):
        for optimized in (False, True):
            with self.subTest(optimized=optimized):
                result = self.preflight(optimized)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(identity, result.stderr)
                self.assertNotIn("No trials selected", result.stderr)

    def test_valid_guidance_reaches_trial_selection(self):
        for optimized in (False, True):
            with self.subTest(optimized=optimized):
                result = self.preflight(optimized)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("No trials selected", result.stderr)

    def test_rejects_changed_concurrency_guidance(self):
        path = self.root / "plugins/go-quality-build/skills/go-concurrency-and-ownership/SKILL.md"
        path.write_text(path.read_text() + "\nChanged instruction.\n")
        self.assert_rejected("go-concurrency-and-ownership/SKILL.md")

    def test_rejects_changed_revision_delta(self):
        path = self.here / "patches/context-guidance-r2.patch"
        path.write_text(path.read_text() + "\n")
        self.assert_rejected("context-guidance-r2.patch")

    def test_rejects_wrong_historical_trial_guidance(self):
        manifest = self.manifest()
        trial = next(t for t in manifest["trials"] if t["name"] == "context-transfer-guided")
        trial["extra_guidance_sha256"]["go-context-and-deadlines/SKILL.md"] = "0" * 64
        self.save_manifest(manifest)
        self.assert_rejected("context-transfer-guided")

    def test_rejects_wrong_shared_trial_guidance(self):
        manifest = self.manifest()
        trial = next(t for t in manifest["trials"] if t["name"] == "concurrency-transfer-baseline")
        trial["shared_guidance_sha256"]["go-context-and-deadlines/SKILL.md"] = "0" * 64
        self.save_manifest(manifest)
        self.assert_rejected("concurrency-transfer-baseline")

    def test_reconstructs_and_checks_historical_revision(self):
        manifest = self.manifest()
        revision = manifest["guidance_revisions"]["go-context-and-deadlines"]
        path = self.here / revision["delta"]
        original = path.read_text()
        changed = original.replace("-The scope owner", "-A different scope owner")
        self.assertNotEqual(original, changed)
        path.write_text(changed)
        revision["delta_sha256"] = hashlib.sha256(path.read_bytes()).hexdigest()
        self.save_manifest(manifest)
        self.assert_rejected("revision_1")


if __name__ == "__main__":
    unittest.main()
