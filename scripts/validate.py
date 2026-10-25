#!/usr/bin/env python3
"""Validate skill portability, the shared review contract, and evaluation inputs.

This checks packaging and consistency, not the quality of an agent's review.
Run behavioral cases separately with their expected answers withheld.
"""

import argparse
from collections import Counter
import json
import re
from pathlib import Path

import yaml


def validate_harnesses(root):
    """Check that each canonical skill package is exposed to all three hosts."""
    root = root.resolve()
    errors = []

    def read_json(path, label):
        try:
            value = json.loads(path.read_text())
        except (OSError, ValueError) as exc:
            errors.append(f"{label}: cannot read JSON: {exc}")
            return None
        if not isinstance(value, dict):
            errors.append(f"{label}: expected a JSON object")
            return None
        return value

    packages = sorted(path for path in (root / "plugins").glob("*") if path.is_dir())
    if not packages:
        errors.append("No plugin packages found under plugins/")
    skill_ids = set()
    names = []
    for folder in packages:
        name = folder.name
        names.append(name)
        if folder.is_symlink() or not folder.resolve().is_relative_to(root):
            errors.append(f"{name}: package must be a real directory inside the repository")
        portable = read_json(folder / "plugin.json", f"{name} portable manifest")
        claude = read_json(folder / ".claude-plugin/plugin.json", f"{name} Claude manifest")
        if portable is not None:
            if portable.get("$schema") != "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json":
                errors.append(f"{name}: portable manifest schema mismatch")
            if portable.get("name") != name:
                errors.append(f"{name}: portable manifest name mismatch")
            if not isinstance(portable.get("version"), str) or not portable["version"]:
                errors.append(f"{name}: portable manifest version missing")
        if claude is not None:
            if claude.get("name") != name:
                errors.append(f"{name}: Claude manifest name mismatch")
            if portable is not None and claude.get("version") != portable.get("version"):
                errors.append(f"{name}: Claude and portable manifest version mismatch")

        skills_dir = folder / "skills"
        if skills_dir.is_symlink() or not skills_dir.resolve().is_relative_to(folder.resolve()):
            errors.append(f"{name}: skills directory must stay inside its package")
        skills = sorted(skills_dir.glob("*/SKILL.md"))
        if not skills:
            errors.append(f"{name}: no skills found")
        for entry in skills:
            skill_id = entry.parent.name
            if entry.parent.is_symlink() or entry.is_symlink():
                errors.append(f"{name}/{skill_id}: runtime skill must not be a symlink")
            if skill_id in skill_ids:
                errors.append(f"duplicate skill ID across packages: {skill_id}")
            skill_ids.add(skill_id)
            if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", skill_id) or len(skill_id) > 64:
                errors.append(f"{name}/{skill_id}: invalid portable skill ID")
            match = re.match(r"\A---\n(.*?)\n---\n", entry.read_text(), re.S)
            if not match:
                errors.append(f"{name}/{skill_id}: missing skill frontmatter")
                continue
            try:
                frontmatter = yaml.safe_load(match[1])
            except yaml.YAMLError as exc:
                errors.append(f"{name}/{skill_id}: invalid skill frontmatter: {exc}")
                continue
            if not isinstance(frontmatter, dict) or frontmatter.get("name") != skill_id:
                errors.append(f"{name}/{skill_id}: frontmatter name must match skill ID")
            if not isinstance(frontmatter, dict) or not isinstance(frontmatter.get("description"), str) or not frontmatter["description"].strip():
                errors.append(f"{name}/{skill_id}: skill description missing")
            for document in [entry, *entry.parent.glob("references/**/*.md")]:
                for target in re.findall(r"\]\(([^)]+)\)", document.read_text()):
                    if re.match(r"[a-zA-Z][a-zA-Z0-9+.-]*:", target) or target.startswith("#"):
                        continue
                    destination = (document.parent / target.split("#", 1)[0]).resolve()
                    if not destination.is_relative_to(entry.parent.resolve()) or not destination.exists():
                        errors.append(f"{name}/{skill_id}: missing skill reference {target} in {document}")

    expected_paths = {name: f"./plugins/{name}" for name in names}
    claude_market = read_json(root / ".claude-plugin/marketplace.json", "Claude marketplace")
    codex_market = read_json(root / ".agents/plugins/marketplace.json", "Codex marketplace")
    for label, catalog in (("Claude", claude_market), ("Codex", codex_market)):
        if catalog is None:
            continue
        if not isinstance(catalog.get("name"), str) or not catalog["name"]:
            errors.append(f"{label} marketplace name missing")
        if label == "Claude" and (not isinstance(catalog.get("owner"), dict) or not isinstance(catalog["owner"].get("name"), str) or not catalog["owner"]["name"]):
            errors.append("Claude marketplace owner missing")
        entries = catalog.get("plugins")
        if not isinstance(entries, list):
            errors.append(f"{label} marketplace plugins must be a list")
            continue
        observed = []
        for entry in entries:
            if not isinstance(entry, dict):
                errors.append(f"{label} marketplace entry must be an object")
                continue
            plugin_name = entry.get("name")
            observed.append(plugin_name)
            source = entry.get("source")
            if label == "Claude":
                path = source
            else:
                path = source.get("path") if isinstance(source, dict) else None
            if plugin_name not in expected_paths or path != expected_paths.get(plugin_name):
                errors.append(f"{label} marketplace target invalid for {plugin_name!r}")
            if label == "Codex":
                if not isinstance(source, dict) or source.get("source") != "local":
                    errors.append(f"Codex marketplace source invalid for {plugin_name!r}")
                policy = entry.get("policy")
                if not isinstance(policy, dict) or policy.get("installation") not in {"AVAILABLE", "INSTALLED_BY_DEFAULT", "NOT_AVAILABLE"} or policy.get("authentication") not in {"ON_INSTALL", "ON_USE"}:
                    errors.append(f"Codex marketplace policy invalid for {plugin_name!r}")
                if not isinstance(entry.get("category"), str) or not entry["category"]:
                    errors.append(f"Codex marketplace category missing for {plugin_name!r}")
        if sorted(observed, key=str) != sorted(names):
            errors.append(f"{label} marketplace inventory does not match packages")
    if claude_market is not None and codex_market is not None and claude_market.get("name") != codex_market.get("name"):
        errors.append("Claude and Codex marketplace names differ")

    opencode = read_json(root / "opencode.json", "OpenCode config")
    if opencode is not None:
        if opencode.get("$schema") != "https://opencode.ai/config.json":
            errors.append("OpenCode config schema mismatch")
        expected_skills = [f"./plugins/{name}/skills" for name in names]
        if not isinstance(opencode.get("skills"), list) or sorted(opencode["skills"], key=str) != expected_skills:
            errors.append("OpenCode skills source does not match packages")
    return errors


def validate(root):
    contract = json.loads((root / "tests/go-quality-review/shared/review-contract.json").read_text())
    errors = []
    case_count = 0
    skills = sorted((root / "plugins/go-quality-review/skills").glob("go-*/SKILL.md"))
    for entry in skills:
        folder = entry.parent
        evaluation = root / "tests/go-quality-review" / folder.name / "evals"
        text = entry.read_text()
        match = re.match(r"\A---\n(.*?)\n---\n", text, re.S)
        if not match:
            errors.append(f"{folder.name}: missing YAML frontmatter")
            continue
        try:
            metadata = yaml.safe_load(match[1])
        except yaml.YAMLError as exc:
            errors.append(f"{folder.name}: {exc}")
            continue
        if not isinstance(metadata, dict):
            errors.append(f"{folder.name}: frontmatter must be a mapping")
            continue
        if metadata.get("name") != folder.name:
            errors.append(f"{folder.name}: name must match directory")
        if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", folder.name) or len(folder.name) > 64:
            errors.append(f"{folder.name}: invalid skill name")
        description = metadata.get("description")
        if not isinstance(description, str) or not 1 <= len(description) <= 1024:
            errors.append(f"{folder.name}: invalid description")
        extra = metadata.get("metadata", {})
        if not isinstance(extra, dict) or extra.get("review-contract") != contract["version"]:
            errors.append(f"{folder.name}: review contract version mismatch")
        if folder.name == "go-quality-report":
            grading = folder / "references/grading.md"
            routing = folder / "references/orchestration.md"
            if not grading.exists() or grading.read_text().count(contract["blocks"]["rubric"]) != 1:
                errors.append("go-quality-report: shared grading rubric mismatch")
            if not routing.exists() or any(name not in routing.read_text() for name in contract["topics"]):
                errors.append("go-quality-report: topic routing inventory incomplete")
            if not (folder / "scripts/grade.py").is_file():
                errors.append("go-quality-report: grade calculator missing")
            reporting = folder / "references/reporting.md"
            report_text = reporting.read_text() if reporting.exists() else ""
            template = contract["report_template"].replace("{topic}", "Go Quality Report")
            if report_text.count(template + contract["umbrella_suffix"]) != 1:
                errors.append("go-quality-report: report template mismatch")
            if report_text.count(contract["report_instructions"]) != 1:
                errors.append("go-quality-report: report instructions mismatch")
        else:
            for name, block in contract["blocks"].items():
                if text.count(block) != 1:
                    errors.append(f"{folder.name}: shared {name} block missing, duplicated, or changed")
            topic = contract["topics"].get(folder.name)
            if topic is None or contract["report_template"].replace("{topic}", topic) not in text:
                errors.append(f"{folder.name}: report template mismatch")
            if text.count(contract["report_instructions"]) != 1:
                errors.append(f"{folder.name}: report instructions mismatch")

        for document in [entry, *folder.glob("references/**/*.md")]:
            for target in re.findall(r"\]\(([^)]+)\)", document.read_text()):
                if re.match(r"[a-zA-Z][a-zA-Z0-9+.-]*:", target) or target.startswith("#"):
                    continue
                local = target.split("#", 1)[0]
                destination = (document.parent / local).resolve()
                if not destination.is_relative_to(folder.resolve()) or not destination.exists():
                    errors.append(f"{document}: missing or non-portable reference {target}")
        try:
            suite = json.loads((evaluation / "evals.json").read_text())
        except (ValueError, OSError) as exc:
            errors.append(f"{folder.name}: invalid evaluation file: {exc}")
            continue
        if suite.get("skill_name") != folder.name:
            errors.append(f"{folder.name}: evaluation skill_name mismatch")
        cases = suite.get("evals", [])
        if not isinstance(cases, list) or not cases:
            errors.append(f"{folder.name}: evaluation cases must be a nonempty list")
            continue
        seen = set()
        for case in cases:
            identifier = case.get("id")
            if not isinstance(identifier, str) or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", identifier):
                errors.append(f"{folder.name}: invalid evaluation id {identifier!r}")
                continue
            if identifier in seen:
                errors.append(f"{folder.name}: duplicate evaluation id {identifier}")
            seen.add(identifier)
            for key in ("prompt", "expected_output"):
                if not isinstance(case.get(key), str) or not case[key].strip():
                    errors.append(f"{folder.name}/{identifier}: missing {key}")
            assertions = case.get("assertions")
            if not isinstance(assertions, list) or not assertions or not all(isinstance(a, str) and a.strip() for a in assertions):
                errors.append(f"{folder.name}/{identifier}: assertions must be nonempty strings")
            for item in case.get("files", []):
                destination = (evaluation.parent / item).resolve()
                if not destination.is_relative_to(evaluation.resolve()) or not destination.exists():
                    errors.append(f"{folder.name}/{identifier}: missing or non-portable fixture {item}")
        case_count += len(cases)
    actual = {p.parent.name for p in skills}
    if actual != set(contract["topics"]) | {"go-quality-report"}:
        errors.append("Skill inventory does not match the shared contract")
    return errors, len(skills), case_count


def validate_build_evals(root: Path) -> tuple[list[str], int]:
    """Validate build cases independently of the fixed review contract."""
    errors = []
    case_count = 0
    installed = {entry.parent.name for entry in
                 (root / "plugins/go-quality-build/skills").glob("*/SKILL.md")}
    candidates = {suite.parent.parent.name for suite in
                  (root / "tests/go-quality-build").glob("*/evals/evals.json")}
    if (root / "tests/go-quality-build/combined/evals").is_dir():
        candidates.add("combined")
    for skill in sorted(installed | candidates):
        evaluation = root / "tests/go-quality-build" / skill / "evals"
        try:
            suite = json.loads((evaluation / "evals.json").read_text())
        except (OSError, ValueError) as exc:
            errors.append(f"{skill}: invalid evaluation file: {exc}")
            continue
        if not isinstance(suite, dict):
            errors.append(f"{skill}: evaluation suite must be an object")
            continue
        if suite.get("skill_name") != skill:
            errors.append(f"{skill}: evaluation skill_name mismatch")
        cases = suite.get("evals")
        if not isinstance(cases, list) or not cases:
            errors.append(f"{skill}: evaluation cases must be a nonempty list")
            continue
        seen = set()
        for case in cases:
            if not isinstance(case, dict):
                errors.append(f"{skill}: evaluation case must be an object")
                continue
            identifier = case.get("id")
            if not isinstance(identifier, str) or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", identifier):
                errors.append(f"{skill}: invalid evaluation id {identifier!r}")
                continue
            if identifier in seen:
                errors.append(f"{skill}: duplicate evaluation id {identifier}")
            seen.add(identifier)
            for key in ("prompt", "expected_output"):
                if not isinstance(case.get(key), str) or not case[key].strip():
                    errors.append(f"{skill}/{identifier}: missing {key}")
            assertions = case.get("assertions")
            if not isinstance(assertions, list) or not assertions or not all(isinstance(a, str) and a.strip() for a in assertions):
                errors.append(f"{skill}/{identifier}: assertions must be nonempty strings")
            files = case.get("files", [])
            if not isinstance(files, list):
                errors.append(f"{skill}/{identifier}: files must be a list")
                continue
            for item in files:
                if not isinstance(item, str) or not item.strip():
                    errors.append(f"{skill}/{identifier}: invalid fixture path")
                    continue
                destination = (evaluation / item).resolve()
                if not destination.is_relative_to(evaluation.resolve()) or not destination.is_file():
                    errors.append(f"{skill}/{identifier}: missing or non-portable fixture {item}")
        case_count += len(cases)
    return errors, case_count


REPORT_GRADES = {"F", "C-", "C", "C+", "B-", "B", "B+", "A-", "A", "A+"}
REPORT_STATES = {"Not applicable", "Insufficient evidence"}
REPORT_SEVERITIES = ("critical", "major", "moderate", "minor")
REPORT_SECTIONS = ("Findings summary", "Finding details", "Strengths", "Checks and limits")
SUMMARY_COLUMNS = ("ID", "Severity", "Topic", "Finding / consequence", "Evidence", "Correction")


def report_table(body, columns, errors, label):
    """Read one pipe table; escaped pipes remain part of their cell."""
    lines = [line.strip() for line in body.splitlines() if line.strip().startswith("|")]
    rows = []
    for line in lines:
        if not line.endswith("|"):
            errors.append(f"{label}: table rows must end with a pipe")
            continue
        cells = [cell.strip() for cell in re.split(r"(?<!\\)\|", line[1:-1])]
        if len(cells) != len(columns):
            errors.append(f"{label}: expected {len(columns)} cells; escape literal pipes")
            continue
        rows.append(cells)
    if len(rows) < 2 or tuple(rows[0]) != tuple(columns):
        errors.append(f"{label}: missing or invalid table header")
        return []
    if any(not re.fullmatch(r":?-{3,}:?", cell) for cell in rows[1]):
        errors.append(f"{label}: invalid table separator")
    for row in rows[2:]:
        if not all(row):
            errors.append(f"{label}: empty table cell")
    return rows[2:]


def report_field(body, field, errors, label):
    values = re.findall(r"^" + re.escape(field) + r":[ \t]*(.*)$", body, re.M)
    if len(values) != 1 or not values[0].strip():
        errors.append(f"{label}: missing, empty or repeated {field}")
        return ""
    return values[0].strip()


def possible_report_grades(counts):
    critical, major, moderate, minor = (counts[s] for s in REPORT_SEVERITIES)
    if critical:
        return {"F"}
    if major >= 2 or (major == 1 and moderate >= 2):
        return {"C-"}
    if major == 1:
        # A systemic-major explanation still needs semantic review.
        return {"C", "C-"}
    if moderate >= 2:
        return {"C+"}
    if moderate == 1:
        return {"B-" if minor >= 2 else "B"}
    if minor >= 2:
        return {"B+"}
    if minor == 1:
        return {"A-"}
    # Independent, verified safeguards still need semantic review.
    return {"A", "A+"}


def check_report(text, contract, expected_topic=None):
    """Check presentation and internal consistency, never factual evidence truth."""
    errors = []
    known_topics = set(contract["topics"].values())
    header = re.match(r"\A\s*# ([^\n]+) — ([^\n]+)\n", text)
    title, grade = (header[1].strip(), header[2].strip()) if header else ("", "")
    umbrella = title == "Go Quality Report"
    if title not in known_topics | {"Go Quality Report"} or grade not in REPORT_GRADES | REPORT_STATES:
        errors.append("invalid topic or grade heading")
    if expected_topic is not None and title != expected_topic:
        errors.append(f"expected topic heading {expected_topic}")

    headings = list(re.finditer(r"^## ([^\n]+)\s*$", text, re.M))
    preamble = text[:headings[0].start()] if headings else text
    for field in ("Scope", "Coverage", "Rationale"):
        report_field(preamble, field, errors, "report")
    sections = {}
    positions = {}
    for index, heading in enumerate(headings):
        name = heading[1].strip()
        if name in sections:
            errors.append(f"repeated section {name}")
        end = headings[index + 1].start() if index + 1 < len(headings) else len(text)
        sections[name] = text[heading.end():end].strip()
        positions[name] = index
    required = REPORT_SECTIONS + (("Report cards", "Maintainability profile") if umbrella else ())
    for name in required:
        if not sections.get(name):
            errors.append(f"missing or empty {name}")
    ordered = [positions[name] for name in required if name in positions]
    if ordered != sorted(ordered):
        errors.append("invalid section order")

    summary = report_table(sections.get("Findings summary", ""), SUMMARY_COLUMNS, errors, "Findings summary")
    if not summary and not re.search(r"^None found\.[ \t]*$", sections.get("Findings summary", ""), re.M):
        errors.append("empty Findings summary must say None found.")
    summary_by_id = {}
    for row in summary:
        fid, severity, topic = row[:3]
        if not re.fullmatch(r"F[1-9]\d*", fid):
            errors.append(f"invalid summary ID {fid}")
        if fid in summary_by_id:
            errors.append(f"duplicate summary ID {fid}")
        if severity not in REPORT_SEVERITIES:
            errors.append(f"{fid}: invalid severity {severity}")
        topics = {part.strip() for part in topic.split(";")}
        if not topics or not topics <= known_topics:
            errors.append(f"{fid}: unknown topic")
        if not umbrella and title in known_topics and topics != {title}:
            errors.append(f"{fid}: standalone finding must use its assessed topic")
        summary_by_id[fid] = (severity, topic)
    if all(re.fullmatch(r"F[1-9]\d*", row[0]) and row[1] in REPORT_SEVERITIES for row in summary):
        order = [(REPORT_SEVERITIES.index(row[1]), int(row[0][1:])) for row in summary]
        if order != sorted(order):
            errors.append("invalid finding order: use severity, then numeric ID")

    detail_body = sections.get("Finding details", "")
    details = list(re.finditer(r"^### (F[1-9]\d*) — ([^\n]+)$", detail_body, re.M))
    all_detail_headings = re.findall(r"^### .+$", detail_body, re.M)
    if len(details) != len(all_detail_headings):
        errors.append("invalid finding detail heading")
    detail_ids = set()
    for index, detail in enumerate(details):
        fid = detail[1]
        if fid in detail_ids:
            errors.append(f"duplicate detail ID {fid}")
        detail_ids.add(fid)
        end = details[index + 1].start() if index + 1 < len(details) else len(detail_body)
        body = detail_body[detail.end():end]
        fields = {field: report_field(body, field, errors, fid) for field in
                  ("Severity", "Topic", "Status", "Trigger", "Evidence", "Consequence", "Correction", "Verification")}
        if fields["Status"] not in {"introduced", "worsened", "existing-in-scope"}:
            errors.append(f"{fid}: invalid Status")
        if fid in summary_by_id:
            severity, topic = summary_by_id[fid]
            if fields["Severity"] != severity:
                errors.append(f"{fid}: severity differs between summary and detail")
            if fields["Topic"] != topic:
                errors.append(f"{fid}: topic differs between summary and detail")
    if set(summary_by_id) != detail_ids:
        errors.append("summary/detail IDs differ")
    if not details and detail_body != "None found.":
        errors.append("empty Finding details must say None found.")

    tally_lines = re.findall(r"^Finding counts:.*$", preamble, re.M)
    tally = re.search(r"^Finding counts: critical=(\d+), major=(\d+), moderate=(\d+), minor=(\d+)[ \t]*$", preamble, re.M)
    if grade in REPORT_GRADES and not tally:
        errors.append("missing or invalid severity counts")
    if len(tally_lines) > 1 or (tally_lines and not tally):
        errors.append("invalid or repeated severity counts")
    observed = Counter(row[1] for row in summary)
    if tally:
        counts = Counter(dict(zip(REPORT_SEVERITIES, map(int, tally.groups()))))
        if observed != counts:
            errors.append("finding severities do not match counts")
        if grade in REPORT_GRADES and grade not in possible_report_grades(counts):
            errors.append(f"grade {grade} conflicts with counts")
        if grade == "Insufficient evidence" and not sum(counts.values()):
            errors.append("Insufficient evidence must not imply a clean assessment with zero counts")
    if grade == "Not applicable" and (tally_lines or summary or details):
        errors.append("Not applicable must not contain findings or counts")
    assessed_lines = re.findall(r"^Assessed-scope defect grade:[ \t]*(.*)$", preamble, re.M)
    if assessed_lines:
        assessed = re.match(r"([^\s(]+)", assessed_lines[0].strip())
        assessed_grade = assessed[1] if assessed else ""
        if len(assessed_lines) != 1 or assessed_grade not in REPORT_GRADES:
            errors.append("invalid, empty or repeated assessed-scope grade")
        if grade != "Insufficient evidence":
            errors.append("assessed-scope grade requires an Insufficient evidence result")
        elif assessed_grade in {"A", "A+"}:
            errors.append("incomplete coverage cannot support an assessed-scope A or A+")
        elif assessed_grade in REPORT_GRADES and assessed_grade not in possible_report_grades(observed):
            errors.append("assessed-scope grade conflicts with confirmed finding counts")

    if umbrella:
        cards = sections.get("Report cards", "")
        card_header = next((line for line in cards.splitlines() if line.strip().startswith("|")), "")
        columns = tuple(cell.strip() for cell in card_header.strip().strip("|").split("|"))
        if len(columns) != 4 or columns[:2] != ("Topic", "Grade/state"):
            errors.append("Report cards: invalid table header")
        else:
            rows = report_table(cards, columns, errors, "Report cards")
            topics = [row[0] for row in rows]
            if len(topics) != len(known_topics) or set(topics) != known_topics:
                errors.append("Report cards must contain every topic exactly once")
            if any(row[1] not in REPORT_GRADES | REPORT_STATES for row in rows):
                errors.append("Report cards: invalid topic grade/state")
    return errors


def check_reports(root, reports):
    """Check each current topic evaluation report against the shared structure."""
    errors = []
    contract = json.loads((root / "tests/go-quality-review/shared/review-contract.json").read_text())
    count = 0
    for skill, topic in contract["topics"].items():
        cases = json.loads((root / "tests/go-quality-review" / skill / "evals/evals.json").read_text())["evals"]
        for case in cases:
            path = reports / skill / f"{case['id']}.md"
            label = f"{skill}/{case['id']}"
            if not path.exists():
                errors.append(f"{label}: missing report")
                continue
            count += 1
            errors.extend(f"{label}: {error}" for error in check_report(path.read_text(), contract, topic))
    return errors, count


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--reports", type=Path, help="Check current topic evaluation reports at <skill>/<case-id>.md")
    parser.add_argument("--report-file", type=Path, action="append", default=[], help="Check one topic or full report; repeat for multiple files")
    args = parser.parse_args()
    try:
        errors, skill_count, case_count = validate(args.root)
        errors.extend(validate_harnesses(args.root))
        build_errors, build_case_count = validate_build_evals(args.root)
        errors.extend(build_errors)
        report_count = None
        if args.reports:
            report_errors, report_count = check_reports(args.root, args.reports)
            errors.extend(report_errors)
        if args.report_file:
            contract = json.loads((args.root / "tests/go-quality-review/shared/review-contract.json").read_text())
            report_count = (report_count or 0) + len(args.report_file)
            for path in args.report_file:
                errors.extend(f"{path}: {error}" for error in check_report(path.read_text(), contract))
    except (OSError, ValueError, TypeError) as exc:
        parser.exit(1, f"Validation could not complete: {exc}\n")
    for error in errors:
        print(f"FAIL: {error}")
    if errors:
        parser.exit(1, f"{len(errors)} validation errors\n")
    print(f"PASS: {skill_count} skills; matching grading/report contracts; {case_count} valid evaluation cases")
    print(f"PASS: {build_case_count} valid build evaluation cases")
    if report_count is not None:
        print(f"PASS: {report_count} report structures, summary/detail consistency, severity tallies, and grade arithmetic")
        print("Factual findings, severity judgments, systemic-major explanations, and A+ evidence still require semantic review.")


if __name__ == "__main__":
    main()
