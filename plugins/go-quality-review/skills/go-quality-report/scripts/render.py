#!/usr/bin/env python3
"""Render a validated review record with computed grades. Python 3.10+, stdlib only."""

import argparse
from pathlib import Path
import re
import sys

from review_record import (DIMENSIONS, SEVERITIES, TOPICS, assessments, evidence_text,
                           load, project, severity, validate)


def inline(value):
    # Record strings may contain Markdown links/code, but cannot add structural lines.
    return " ".join(value.split())


def cell(value):
    return re.sub(r"(?<!\\)\|", r"\\|", inline(value))


def table(columns, rows):
    def line(values):
        return "| " + " | ".join(cell(v) for v in values) + " |"
    return [line(columns), line(["---"] * len(columns)), *(line(row) for row in rows)]


def topic_names(finding, topic):
    names = {a["topic"] for a in assessments(finding, topic)}
    return "; ".join(TOPICS[name] for name in TOPICS if name in names)


def consequences(finding, topic):
    by_topic = {a["topic"]: a for a in assessments(finding, topic)}
    result = []
    for name in TOPICS:
        if name not in by_topic:
            continue
        assessment = by_topic[name]
        line = f"{TOPICS[name]} ({assessment['severity']}): {assessment['consequence']}"
        if assessment.get("systemic"):
            line += " Systemic major."
        if "reach" in assessment:
            line += " Reach: " + assessment["reach"]
        result.append(line)
    return inline(" ".join(result))


def render(data, topic=None):
    validate(data)
    topic = topic or data["report"]
    if topic != "overall" and topic not in TOPICS:
        raise ValueError(f"unknown topic: {topic}")
    if topic == "overall" and data["report"] != "overall":
        raise ValueError("a standalone record cannot produce an overall report")
    ledger, grade = project(data, topic)
    scoped_topics = [t for t in data["topics"] if topic == "overall" or t["id"] == topic]
    rationale = data["rationale"] if topic == "overall" else scoped_topics[0]["rationale"]
    title = "Go Quality Report" if topic == "overall" else TOPICS[topic]
    lines = [f"# {title} — {grade['overall']}",
             "Scope: " + inline(data["mode"] + "; " + data["scope"]),
             "Coverage: " + inline("; ".join(f"{c['id']}: {c['status']} — {c['reason']}" for c in ledger["coverage"])),
             "Rationale: " + inline(rationale)]
    if ledger["findings"] or grade["overall"] not in ("Not applicable", "Insufficient evidence"):
        lines.append("Finding counts: " + ", ".join(f"{s}={grade['unique_finding_counts'][s]}" for s in SEVERITIES))
    if grade["assessed_scope_grade"]:
        lines.append("Assessed-scope defect grade: " + grade["assessed_scope_grade"] + " (reviewed areas only)")
    findings = sorted((f for f in data["findings"] if assessments(f, topic)),
                      key=lambda f: (SEVERITIES.index(severity(f, topic)), int(f["id"][1:])))
    lines += ["", "## Findings summary", ""]
    lines += table(("ID", "Severity", "Topic", "Finding / consequence", "Evidence", "Correction"), [
        (f["id"], severity(f, topic), topic_names(f, topic), f["title"] + " — " + consequences(f, topic),
         "; ".join(e["reference"] for e in f["evidence"]), f["correction"]) for f in findings])
    if not findings:
        lines += ["", "None found."]
    lines += ["", "## Finding details", ""]
    if not findings:
        lines.append("None found.")
    for f in findings:
        lines += [f"### {f['id']} — {inline(f['title'])}", "",
                  "Severity: " + severity(f, topic), "Topic: " + topic_names(f, topic),
                  "Status: " + f["status"], "Trigger: " + inline(f["trigger"]),
                  "Evidence: " + inline(evidence_text(f["evidence"])),
                  "Consequence: " + consequences(f, topic),
                  "Correction: " + inline(f["owner"] + " — " + f["correction"]),
                  "Verification: Proposed: " + inline(f["verification"]),
                  "Analysis: " + inline(f["cause"] + ". " + f["analysis"]),
                  "Sources: " + inline("; ".join(f["sources"])), ""]
    lines += ["", "## Strengths", ""]
    strengths = [s for s in data["strengths"] if topic == "overall" or topic in s["topics"]]
    for s in sorted(strengths, key=lambda s: int(s["id"][1:])):
        lines.append(f"- [{s['id']}] " + inline(s["description"] + " — " + evidence_text(s["evidence"])))
    if not strengths:
        lines.append("None verified.")
    for s in ledger["safeguards"]:
        lines.append(f"- [{s['id']}] " + inline(f"{'Nonroutine' if s['nonroutine'] else 'Routine'} safeguard for {s['risk']} — {s['evidence']}"))
    lines += ["", "## Checks and limits", ""]
    if not data["checks"]:
        lines.append("No checks recorded; inspected and supplied evidence is labeled above.")
    for check in data["checks"]:
        lines.append("- " + inline(f"{check['kind']}: {check['command']} — {check['result']}"))
    for limit in data["limits"]:
        lines.append("- " + inline(limit))
    lines.append("Grades and structure are computed; evidence, admission, severity and coverage require reviewer judgment.")
    if topic == "overall":
        lines += ["", "## Report cards", ""]
        by_topic = {t["id"]: t for t in data["topics"]}
        card_rows = []
        for name in TOPICS:
            topic_ledger, result = project(data, name)
            entry = by_topic[name]
            coverage = "; ".join(f"{c['status']}: {c['reason']}" for c in topic_ledger["coverage"])
            card_rows.append((TOPICS[name], result["overall"], entry["rationale"] + " " + coverage,
                              "; ".join(entry["cards"]) or "Reviewed in this record"))
        lines += table(("Topic", "Grade/state", "Assessed chunks and coverage limits", "Cards"), card_rows)
        lines += ["", "## Maintainability profile", ""]
        profile = data["maintainability"]
        by_dimension = {r["dimension"]: r for r in profile["rows"]}
        lines += table(("Dimension", "Rating", "Confidence", "Evidence / counterevidence", "Scope / next validation"), [
            (name, by_dimension[name]["rating"], by_dimension[name]["confidence"], by_dimension[name]["evidence"], by_dimension[name]["limits"])
            for name in DIMENSIONS])
        for key in ("cleanliness", "structure", "extensibility"):
            lines += ["", key.capitalize() + ": " + inline(profile[key])]
    decisions = [d for d in data["decisions"] if topic == "overall" or topic in d["topics"] or
                 (data["report"] != "overall" and d["disposition"] == "related")]
    if decisions:
        lines += ["", "## Ungraded decisions", ""]
        for d in sorted(decisions, key=lambda d: int(d["id"][1:])):
            target = " → " + d["finding_id"] if "finding_id" in d else ""
            topics = "; ".join(TOPICS[name] for name in TOPICS if name in d["topics"])
            lines.append(f"- {d['id']} ({d['disposition']}{target}; {topics}): " + inline(d["title"] + " — " + d["reason"] + " " + evidence_text(d["evidence"])))
    return "\n".join(lines).rstrip() + "\n"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("record", type=Path)
    parser.add_argument("--topic", choices=tuple(TOPICS), help="render one topic from an overall record")
    parser.add_argument("--output", type=Path, help="write Markdown here; default stdout")
    args = parser.parse_args()
    try:
        if args.output and (args.output.resolve() == args.record.resolve() or
                            (args.output.exists() and args.output.samefile(args.record))):
            raise ValueError("output must not overwrite the input record")
        rendered = render(load(args.record), args.topic)
        if args.output:
            args.output.write_text(rendered, encoding="utf-8")
        else:
            sys.stdout.write(rendered)
    except (ValueError, OSError) as exc:
        parser.exit(2, f"error: {exc}\n")


if __name__ == "__main__":
    main()
