"""Independent read-only readiness evidence checks; no author/check outcome repair."""
from pathlib import Path
import argparse
import hashlib
import json
import re
import shutil
import subprocess
import tempfile

ROOT = next(p for p in Path(__file__).resolve().parents if (p / "plugins/go-quality-build").is_dir())
STUDY = ROOT / "tests/go-quality-build/results/2026-10-02-concurrency-study"
EVALS = ROOT / "tests/go-quality-build/go-concurrency-and-ownership/evals"
PREFLIGHT = ROOT / "docs/go-quality-build/reviews/concurrency-readiness/preflight"
DRAFT_SHA = "0ab21aee46b9446356b455d59776263a8160fcf2a0c3b5050c2b1792ecf13002"
INDEX_SHA = "d4639503c4138b8c9d33cd865a2619ddfdab87b6cfc03f12a04a8cf63e3ccd74"
COMPARATOR = "0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96"


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def tree(path):
    return {str(p.relative_to(path)): digest(p) for p in path.rglob("*") if p.is_file()}


def run(argv, cwd=ROOT):
    p = subprocess.run(argv, cwd=cwd, capture_output=True, text=True)
    assert p.returncode == 0, (argv, p.returncode, p.stderr)
    return p.stdout


def check():
    parser = argparse.ArgumentParser()
    parser.add_argument("--final", action="store_true")
    args = parser.parse_args()
    m = read(STUDY / "manifest.json")
    results = {"mode": "final" if args.final else "identity-stage", "study_manifest_sha256": digest(STUDY / "manifest.json")}

    # Reuse, rather than recompute, the accepting preflight's expensive checks.
    assert digest(PREFLIGHT / "checksums.json") == INDEX_SHA == m["preflight"]["index_sha256"]
    index = read(PREFLIGHT / "checksums.json")["artifacts"]
    assert len(index) == 64
    assert {r["path"]: r["sha256"] for r in index} == {k: v for k, v in tree(PREFLIGHT).items() if k != "checksums.json"}
    for r in index:
        assert (PREFLIGHT / r["path"]).stat().st_size == r["bytes"]
    assert tree(PREFLIGHT) == tree(STUDY / "preflight-review")
    assert read(PREFLIGHT / "findings.json")["disposition"] == "Accept"
    assert digest(PREFLIGHT / "preflight-review.md") == m["preflight"]["artifact_sha256"]
    draft = ROOT / "tests/go-quality-build/go-concurrency-and-ownership/draft/SKILL.md"
    assert digest(draft) == DRAFT_SHA == m["draft_revision_sha256"]
    assert draft.stat().st_size == 5274
    fence = re.search(r"```go\n(.*?)\n```", draft.read_text(), re.S).group(1)
    assert (PREFLIGHT / "example/totals.go").read_text() == "package example\n\nimport \"sync\"\n\n" + fence + "\n"
    results["preflight"] = {"index_sha256": INDEX_SHA, "artifacts": 64, "copied_archive_identical": True, "draft_sha256": DRAFT_SHA, "example_bytes_identical": True, "expensive_checks_reused": True}

    # Current comparator and review guidance must remain equal to the frozen base.
    identity_rows = read(PREFLIGHT / "raw/input-identities.json")
    comparator_rows = [r for r in identity_rows if r["path"].startswith("plugins/go-quality-build/skills/")]
    reviewer_rows = [r for r in identity_rows if r["path"].startswith("plugins/go-quality-review/skills/")]
    assert len(comparator_rows) == 8 and len(reviewer_rows) == 23
    for row in comparator_rows + reviewer_rows:
        path = ROOT / row["path"]
        assert digest(path) == row["sha256"]
        committed = run(["rtk", "proxy", "git", "show", COMPARATOR + ":" + row["path"]])
        assert hashlib.sha256(committed.encode()).hexdigest() == row["sha256"]
    assert len(m["skill_hashes"]) == 9
    assert tree(STUDY / "skills") == m["skill_hashes"]
    assert len(m["input_hashes"]) == 23
    for relative, expected in m["input_hashes"].items():
        assert digest(EVALS / relative) == expected
    assert m["comparator_revision"] == COMPARATOR
    results["frozen_inputs"] = {"comparator_files": 8, "review_guidance_files": 23, "catalog_files": 9, "eval_probe_files": 23, "all_equal_to_frozen_identity": True}
    runtime_skill_names = sorted(p.name for p in (ROOT / "plugins/go-quality-build/skills").iterdir() if p.is_dir())
    assert len(runtime_skill_names) == 5 and "go-concurrency-and-ownership" not in runtime_skill_names and "go-context-and-deadlines" not in runtime_skill_names
    runtime_metadata = [ROOT / "plugins/go-quality-build/plugin.json", ROOT / "plugins/go-quality-build/.claude-plugin/plugin.json"]
    assert all(read(p)["version"] == "0.2.0" for p in runtime_metadata)
    results["runtime"] = {"skill_names": runtime_skill_names, "versions": {str(p.relative_to(ROOT)): read(p)["version"] for p in runtime_metadata}, "candidate_uninstalled": True}

    # Check source-copy aids and local primary sources without assuming the heuristic proves authorship.
    sources = read(ROOT / "docs/go-quality-build/context-concurrency-source-inventory.json")
    assert len(sources) == 20
    captures = read(PREFLIGHT / "raw/source-inventory-verification.json")
    assert len(captures) == 20
    assert {(r["repository"], r["revision"], r["path"], r["sha256"]) for r in sources} == {(r["repository"], r["revision"], r["path"], r["sha256"]) for r in captures}
    normalized = lambda s: re.findall(r"[a-z0-9]+", s.lower())
    draft_words = normalized(draft.read_text())
    draft_grams = {tuple(draft_words[i:i + 12]) for i in range(len(draft_words) - 11)}
    overlaps = []
    for row in captures:
        path = Path(row["capture_path"])
        assert digest(path) == row["sha256"] and path.stat().st_size == row["bytes"]
        if row["path"] != "LICENSE":
            words = normalized(path.read_text())
            overlaps += [{"path": row["path"], "words": words[i:i + 12]} for i in range(len(words) - 11) if tuple(words[i:i + 12]) in draft_grams]
    assert not overlaps
    primary_rows = read(PREFLIGHT / "raw/local-primary-doc-verification.json")
    primary_present = 0
    for row in primary_rows:
        path = Path(row["path"])
        assert path.exists() == row["exists"]
        if path.exists():
            assert digest(path) == row["sha256"]
            primary_present += 1
    results["source_semantics"] = {"pinned_captures_verified": 20, "guidance_captures": 16, "normalized_12_word_overlaps": overlaps, "primary_local_records": len(primary_rows), "primary_local_files_present": primary_present, "overlap_limit": "Supporting inspection only; not a legal threshold or proof of authorship."}

    # Preserve and verify all three excluded discovery authors in the spent budget.
    discovery = ROOT / m["discovery_archive"]
    discovery_manifest = read(discovery / "manifest.json")
    assert discovery_manifest["authors_spent_total"] == 3
    assert discovery_manifest["group_authors_spent_total"] == 19
    assert m["group_authors_spent_before_candidate"] == 16
    assert len(discovery_manifest["attempts"]) == 3
    discovery_reconstructions = []
    for row in discovery_manifest["attempts"]:
        archive = discovery / row["directory"]
        assert row["counted"] and row["status"] == "completed"
        with tempfile.TemporaryDirectory(prefix="independent-discovery-reconstruct-", dir="/private/tmp") as temporary:
            work = Path(temporary)
            shutil.copytree(EVALS / "files" / row["case"], work, dirs_exist_ok=True)
            run(["rtk", "proxy", "git", "apply", str(archive / "source.patch")], work)
            assert tree(work) == read(archive / "source-hashes.json") == tree(archive / "source")
        dispatch = read(archive / "dispatch.txt")
        assert dispatch["environment_overrides"] == {"GOCACHE": "/private/tmp/go-quality-author-cache", "GOMODCACHE": "/private/tmp/go-quality-author-modcache"}
        assert "go-concurrency-and-ownership" not in (archive / "prompt.txt").read_text()
        events = [json.loads(line) for line in (archive / "events.jsonl").read_text().splitlines()]
        cost = read(archive / "cost.json")
        assert [e["usage"] for e in events if e["type"] == "turn.completed"] == cost["usage"]
        assert cost["elapsed_seconds"] == row["elapsed_seconds"]
        discovery_reconstructions.append({"attempt": row["number"], "source_identity": True, "raw_cost_identity": True, "matched_benefit_eligible": False, "environment_policy": "GOTOOLCHAIN inherited; study forces local"})
    results["excluded_discovery"] = discovery_reconstructions

    # Reconstruct all twelve completed study outcomes from the original input, independently.
    assert len(m["attempts"]) == 12 and len({r["number"] for r in m["attempts"]}) == 12
    assert m["authors_spent_total"] == 15 <= m["author_ceiling"] == 16
    assert m["group_authors_spent_total"] == 31 <= m["group_ceiling"] == 36
    assert m["authors_spent_before_study"] == 3
    assert m["reference_environment_correction"]["discovery_matched_benefit_eligible"] is False
    assert m["reference_environment_correction"]["pairs"] == [[13, 4], [14, 5], [15, 6]]
    cost_rows = read(STUDY / "cost-summary.json")["rows"]
    assert len(cost_rows) == 12
    reconstructions = []
    for row in m["attempts"]:
        archive = STUDY / row["directory"]
        assert row["counted"] and row["status"] == "completed"
        with tempfile.TemporaryDirectory(prefix="independent-concurrency-reconstruct-", dir="/private/tmp") as temporary:
            work = Path(temporary)
            shutil.copytree(EVALS / "files" / row["case"], work, dirs_exist_ok=True)
            run(["rtk", "proxy", "git", "apply", str(archive / "source.patch")], work)
            hashes = tree(work)
            assert hashes == read(archive / "source-hashes.json") == tree(archive / "source")
            assert (work / "go.mod").read_bytes() == (EVALS / "files" / row["case"] / "go.mod").read_bytes()
        dispatch = read(archive / "dispatch.txt")
        assert dispatch["case"] == row["case"]
        assert dispatch["environment_overrides"] == {"GOCACHE": "/private/tmp/go-quality-author-cache", "GOMODCACHE": "/private/tmp/go-quality-author-modcache", "GOTOOLCHAIN": "local"}
        assert dispatch["host_disable_sha256"] == m["host_disable_sha256"]
        argv = dispatch["argv"]
        assert argv[argv.index("--model") + 1] == row["model"]
        assert 'model_reasoning_effort="' + row["reasoning"] + '"' in argv
        assert all(flag in argv for flag in ["--ignore-user-config", "--ephemeral", "--approve-for-me"])
        assert "--sandbox" not in argv
        physical_catalog = Path(dispatch["cwd"]) / ".agents/skills"
        expected_catalog = {k: v for k, v in m["skill_hashes"].items() if row["arm"] == "exposure" or not k.startswith("go-concurrency-and-ownership/")}
        assert tree(physical_catalog) == expected_catalog
        prompt = (archive / "prompt.txt").read_text()
        assert ("go-concurrency-and-ownership" in prompt) == (row["arm"] == "exposure")
        assert "go-context-and-deadlines" not in prompt
        assert "controller-probes" not in prompt and "expected_output" not in prompt
        completion = read(archive / "completion.json")
        assert completion["status"] == "completed" and completion["exit_code"] == row["exit_code"]
        assert completion["elapsed_seconds"] == row["elapsed_seconds"]
        evidence = read(archive / "verification.json")
        assert evidence["reconstruction_hashes_match"]
        for name, expected in m["declared_check_statuses"][str(row["number"])].items():
            assert [c["status"] for c in evidence[name]] == expected
        assert len(evidence["ordinary_checks"]) == 8 and all(c["status"] == 0 for c in evidence["ordinary_checks"])
        assert not evidence["ordinary_checks"][6]["stdout"].strip()
        cost = read(archive / "cost.json")
        events = [json.loads(line) for line in (archive / "events.jsonl").read_text().splitlines()]
        usage = [e["usage"] for e in events if e["type"] == "turn.completed"]
        commands = [e["item"]["command"] for e in events if e["type"] == "item.completed" and e["item"]["type"] == "command_execution"]
        observed = read(archive / "selection.json")["observed_skill_commands"]
        assert all(c in commands for c in observed)
        assert {c for c in commands if "go-concurrency-and-ownership/SKILL.md" in c} == {c for c in observed if "go-concurrency-and-ownership/SKILL.md" in c}
        if row["arm"] == "exposure" and row["case"] != "not-concurrency-work":
            assert any("go-concurrency-and-ownership/SKILL.md" in c for c in observed)
        assert usage == cost["usage"] and cost["elapsed_seconds"] == row["elapsed_seconds"]
        summary = next(c for c in cost_rows if c["number"] == row["number"])
        assert summary["usage"] == usage and summary["elapsed_seconds"] == row["elapsed_seconds"]
        for key in ["case", "arm", "profile", "model", "reasoning"]:
            assert summary[key] == row[key]
        assert summary["source_files"] == len(hashes)
        assert summary["source_patch_bytes"] == (archive / "source.patch").stat().st_size
        go_files = list((archive / "source").rglob("*.go"))
        assert summary["production_lines"] == sum(len(p.read_text().splitlines()) for p in go_files if not p.name.endswith("_test.go"))
        assert summary["test_lines"] == sum(len(p.read_text().splitlines()) for p in go_files if p.name.endswith("_test.go"))
        reconstructions.append({"attempt": row["number"], "files": len(hashes), "source_identity": True, "ordinary_statuses": [c["status"] for c in evidence["ordinary_checks"]], "held_statuses": [c["status"] for c in evidence["held_contract_checks"]], "supplementary_statuses": [c["status"] for c in evidence["supplementary_review_discovered_checks"]], "launch_and_cost_records_match": True})
    results["authors"] = {"candidate_count": 15, "group_count": 31, "remaining_candidate_slots": 1, "study_outcomes": reconstructions, "original_discovery_excluded_from_matched_credit": True}
    prompts = {r["number"]: (STUDY / r["directory"] / "prompt.txt").read_text() for r in m["attempts"]}
    for baseline, exposure in [(13, 4), (14, 5), (15, 6), (7, 8), (9, 10), (11, 12)]:
        without_candidate = re.sub(r"name: go-concurrency-and-ownership\ndescription: [^\n]*\nfile: [^\n]*\n\n", "", prompts[exposure])
        assert without_candidate == prompts[baseline], (baseline, exposure, "unmatched prompt")

    # Body routing is observed from native commands, never inferred from status=0.
    selection = read(STUDY / "selection-summary.json")
    assert len(selection["rows"]) == 12
    catalogs = read(STUDY / "selection-catalog-identity.json")
    assert len(catalogs) == 12 and all(c["catalog_files"] == m["skill_hashes"] for c in catalogs)
    selection_profiles = {}
    for row in selection["rows"]:
        selection_folder = STUDY / Path(row["completion"]).parent
        completion = read(STUDY / row["completion"])
        dispatch = read(selection_folder / "dispatch.json")
        assert dispatch["host_disable_sha256"] == m["host_disable_sha256"]
        assert tree(Path(dispatch["cwd"]) / ".agents/skills") == m["skill_hashes"]
        events = [json.loads(line) for line in (selection_folder / "events.jsonl").read_text().splitlines()]
        raw_commands = [e["item"]["command"] for e in events if e["type"] == "item.completed" and e["item"]["type"] == "command_execution"]
        assert completion["observed_commands"] == raw_commands
        assert completion["status"] == row["status"] == 0
        body = any("go-concurrency-and-ownership/SKILL.md" in c for c in completion["observed_commands"])
        assert body == row["candidate_body_read_observed"]
        assert row["matches_intended_selection"] == (body == row["request"].startswith("applicable"))
        assert "go-concurrency-and-ownership" not in (STUDY / Path(row["completion"]).parent / "prompt.txt").read_text()
        profile = selection_profiles.setdefault(row["profile"], {"matches": 0, "probes": 0, "deviations": []})
        profile["probes"] += 1
        profile["matches"] += int(row["matches_intended_selection"])
        if not row["matches_intended_selection"]:
            profile["deviations"].append(row["request"])
    assert sum(p["matches"] for p in selection_profiles.values()) == 10
    assert selection["selection_probes_spent_total"] == m["selection_probes_spent_total"] == 12
    assert selection["group_selection_probes_spent_total"] == m["group_selection_probes_spent_total"] == 26
    assert m["profiles"]["native-claude"]["available"] is False
    assert m["profiles"]["native-opencode"]["available_for_target_v2_loading"] is False
    results["routing"] = {"profiles": selection_profiles, "probes": 12, "matching_body_decisions": 10, "unavailable": ["native-claude unauthenticated", "native-opencode target v2 unavailable"], "status_zero_does_not_establish_selection": True}

    if args.final:
        mapping = read(STUDY / "neutral-packet-map.json")
        assert mapping["cards_frozen_before_mapping_disclosure"]
        groups = {}
        mapped_attempts = []
        for group, labels in mapping["mapping"].items():
            folder = STUDY / "neutral-reviews" / group
            freeze = read(STUDY / "neutral-reviews" / (group + "-freeze.json"))
            assert freeze["frozen_before_arm_disclosure"]
            for name, expected in freeze["hashes"].items():
                assert digest(folder / name) == expected, (group, name)
            for label, number in labels.items():
                assert tree(folder / "candidates" / label) == read(STUDY / f"attempt-{number:02d}/source-hashes.json")
                mapped_attempts.append(number)
                packet_path = STUDY / f"attempt-{number:02d}/reviews/packet.json"
                packet = read(packet_path)
                assert packet["candidate"] == label and packet["nine_topics_considered"]
                assert (packet_path.parent / packet["neutral_packet"]).resolve() == folder.resolve()
                assert (packet_path.parent / packet["freeze_index"]).resolve() == (STUDY / "neutral-reviews" / (group + "-freeze.json")).resolve()
            case = {"library": "library-shared-state", "control": "not-concurrency-work", "pipeline": "cli-pipeline-stop", "workers": "service-owned-workers", "service": "service-owned-workers"}[group]
            assert tree(folder / "original") == tree(EVALS / "files" / case)
            review_files = tree(folder / "review-guidance")
            assert len(review_files) == 23
            runtime_review = tree(ROOT / "plugins/go-quality-review/skills")
            assert review_files == runtime_review
            cards = [p for p in folder.rglob("*.md") if p.parent.name in labels and p.parent.parent.name == "cards"]
            assert len(cards) == 9 * len(labels)
            applicability = {}
            for label in labels:
                paths = [p for p in folder.rglob("applicability*.json") if p.name == f"applicability-{label}.json" or p.name == "applicability.json" and p.parent.name == label]
                assert len(paths) == 1, (group, label, paths)
                doc = read(paths[0])
                rows = doc if isinstance(doc, list) else doc.get("topics", doc.get("all_nine"))
                assert len(rows) == 9
                statuses = [r.get("coverage_status", r.get("coverage", r.get("status"))) for r in rows]
                assert all(s in ["complete", "not_applicable"] for s in statuses)
                assert all(r["reason"].strip() for r in rows)
                applicability[label] = {"complete_applicable": statuses.count("complete"), "justified_not_applicable": statuses.count("not_applicable"), "material_in_scope_gaps": 0}
            groups[group] = {"candidates": len(labels), "cards": len(cards), "frozen_artifacts_verified": len(freeze["hashes"]), "review_guidance_files": len(review_files), "applicability": applicability}
        assert sum(g["cards"] for g in groups.values()) == 108
        assert sorted(mapped_attempts) == sorted(r["number"] for r in m["attempts"])
        results["neutral_packets"] = {"groups": groups, "total_cards": 108, "mapping_sha256": digest(STUDY / "neutral-packet-map.json"), "comparison_sha256": digest(STUDY / "comparison.md")}
        capability_path = STUDY / "preparation/current-capability-checks.json"
        capability = read(capability_path)
        assert "not historical replay" in capability["captured"]
        by_argv = {tuple(r["argv"]): r for r in capability["rows"]}
        assert by_argv[("rtk", "proxy", "codex", "--version")]["stdout"].strip() == "codex-cli 0.157.0"
        assert by_argv[("rtk", "proxy", "claude", "--version")]["stdout"].strip() == "2.1.207 (Claude Code)"
        assert by_argv[("rtk", "proxy", "opencode", "--version")]["stdout"].strip() == "1.18.7"
        auth = by_argv[("rtk", "proxy", "claude", "auth", "status")]
        assert auth["status"] == 1 and json.loads(auth["stdout"])["loggedIn"] is False
        results["capability_provenance"] = {"current_read_only_capture_sha256": digest(capability_path), "current_versions_and_auth_corroborate_manifest": True, "historical_raw_versions_auth": "Unavailable: original live observations recorded in canonical/manifest but raw outputs not preserved.", "historical_model_help": "OpenCode model/help captures available in context-discovery/preparation/native-capability-checks.json", "current_capture_is_historical_replay": False}
    else:
        results["neutral_packets"] = {"status": "Final worker freeze, disclosed mapping and final comparison intentionally pending; no final readiness verdict."}
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    check()
