"""Contract checks for the three harness entry points."""

import importlib.util
import json
from pathlib import Path
import shutil
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[1] / "scripts/validate.py"
REPOSITORY = SCRIPT.parent.parent


def load_validator():
    spec = importlib.util.spec_from_file_location("repository_validator", SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value))


def package(root, name, skill="one", version="0.1.0"):
    folder = root / "plugins" / name
    write_json(
        folder / "plugin.json",
        {
            "$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
            "name": name,
            "version": version,
            "description": "Example package",
        },
    )
    write_json(folder / ".claude-plugin" / "plugin.json", {"name": name, "version": version})
    skill_file = folder / "skills" / skill / "SKILL.md"
    skill_file.parent.mkdir(parents=True, exist_ok=True)
    skill_file.write_text(f"---\nname: {skill}\ndescription: Use when reviewing code.\n---\n")


def catalogs(root, names):
    write_json(
        root / ".claude-plugin" / "marketplace.json",
        {
            "name": "personal-skills",
            "owner": {"name": "Owner"},
            "plugins": [{"name": name, "source": f"./plugins/{name}"} for name in names],
        },
    )
    write_json(
        root / ".agents" / "plugins" / "marketplace.json",
        {
            "name": "personal-skills",
            "plugins": [
                {
                    "name": name,
                    "source": {"source": "local", "path": f"./plugins/{name}"},
                    "policy": {"installation": "AVAILABLE", "authentication": "ON_INSTALL"},
                    "category": "Productivity",
                }
                for name in names
            ],
        },
    )
    write_json(
        root / "opencode.json",
        {
            "$schema": "https://opencode.ai/config.json",
            "skills": [f"./plugins/{name}/skills" for name in names],
        },
    )


class HarnessLayoutTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        package(self.root, "go-quality-review")
        catalogs(self.root, ["go-quality-review"])

    def errors(self):
        return load_validator().validate_harnesses(self.root)

    def test_valid_collection_is_accepted(self):
        self.assertEqual([], self.errors())

    def test_manifest_version_drift_is_rejected(self):
        write_json(
            self.root / "plugins/go-quality-review/.claude-plugin/plugin.json",
            {"name": "go-quality-review", "version": "0.2.0"},
        )
        self.assertTrue(any("version" in error for error in self.errors()))

    def test_catalog_and_opencode_paths_must_resolve(self):
        catalogs(self.root, ["missing-package"])
        errors = self.errors()
        self.assertTrue(any("Claude" in error for error in errors))
        self.assertTrue(any("Codex" in error for error in errors))
        self.assertTrue(any("OpenCode" in error for error in errors))

    def test_duplicate_skill_ids_across_packages_are_rejected(self):
        package(self.root, "another-package", skill="one")
        catalogs(self.root, ["go-quality-review", "another-package"])
        self.assertTrue(any("duplicate skill" in error for error in self.errors()))

    def test_broken_skill_reference_is_rejected(self):
        skill = self.root / "plugins/go-quality-review/skills/one/SKILL.md"
        skill.write_text(skill.read_text() + "\n[missing](references/missing.md)\n")
        self.assertTrue(any("missing skill reference" in error for error in self.errors()))

    def test_go_validator_requires_umbrella_skill(self):
        shutil.copytree(
            REPOSITORY / "plugins/go-quality-review/skills",
            self.root / "plugins/go-quality-review/skills",
            dirs_exist_ok=True,
        )
        shutil.copytree(
            REPOSITORY / "tests/go-quality-review",
            self.root / "tests/go-quality-review",
            ignore=shutil.ignore_patterns("__pycache__", "*.pyc"),
        )
        shutil.rmtree(self.root / "plugins/go-quality-review/skills/go-quality-report")
        errors, _, _ = load_validator().validate(self.root)
        self.assertTrue(any("Skill inventory" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
