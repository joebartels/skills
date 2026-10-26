"""Validate review record v1 and project it into the existing grading ledger."""

import json
import re

from grade import calculate


TOPICS = {
    "go-architecture-and-design": "Architecture & Design",
    "go-code-quality-and-idioms": "Code Quality & Go Idioms",
    "go-correctness-and-compatibility": "Correctness & Compatibility",
    "go-testing": "Testing",
    "go-security": "Security",
    "go-observability-and-resilience": "Observability & Resilience",
    "go-performance-and-resource-management": "Performance & Resource Management",
    "go-dependencies-and-reproducibility": "Dependencies & Reproducibility",
    "go-deployment-and-operations": "Deployment & Operations",
}
SEVERITIES = ("critical", "major", "moderate", "minor")
DIMENSIONS = ("Clarity", "Change locality", "Cohesion and boundaries", "Abstraction cost",
              "Rule ownership", "Testability", "Complexity hotspots")


def fields(value, required, optional=()):
    if not isinstance(value, dict):
        raise ValueError("expected an object")
    missing = set(required) - value.keys()
    extra = value.keys() - set(required) - set(optional)
    if missing or extra:
        raise ValueError(f"missing fields {sorted(missing)}; unknown fields {sorted(extra)}")


def text(value):
    if not isinstance(value, str) or not value.strip():
        raise ValueError("expected nonempty text")
    return value


def choice(value, allowed):
    if not isinstance(value, str) or value not in allowed:
        raise ValueError(f"expected one of {tuple(allowed)}, got {value!r}")


def array(value, nonempty=False):
    if not isinstance(value, list) or (nonempty and not value):
        raise ValueError("expected " + ("a nonempty" if nonempty else "an") + " array")
    return value


def unique(values):
    for value in values:
        text(value)
    if len(set(values)) != len(values):
        raise ValueError("duplicate identifier or reference")


def rows(value, id_key="id", pattern=None, nonempty=False):
    result = array(value, nonempty)
    ids = []
    for row in result:
        if not isinstance(row, dict):
            raise ValueError("expected object rows")
        identifier = text(row.get(id_key))
        if pattern and not re.fullmatch(pattern, identifier):
            raise ValueError(f"invalid {id_key}: {identifier}")
        ids.append(identifier)
    unique(ids)
    return result


def evidence(value):
    for item in array(value, nonempty=True):
        fields(item, ("kind", "reference", "text"))
        choice(item["kind"], ("inspected", "executed", "supplied"))
        text(item["reference"])
        text(item["text"])


def topic_refs(value, available):
    unique(array(value, nonempty=True))
    for topic in value:
        choice(topic, available)


def load(path):
    def pairs(items):
        result = {}
        for key, value in items:
            if key in result:
                raise ValueError(f"duplicate JSON key: {key}")
            result[key] = value
        return result

    def constant(value):
        raise ValueError(f"invalid JSON constant: {value}")

    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=pairs, parse_constant=constant)


def validate(data):
    fields(data, ("version", "report", "mode", "scope", "rationale", "topics", "findings", "decisions",
                  "strengths", "safeguards", "checks", "limits"), ("maintainability",))
    if type(data["version"]) is not int or data["version"] != 1:
        raise ValueError("unsupported review record version; expected 1")
    choice(data["report"], ("overall", *TOPICS))
    choice(data["mode"], ("changeset", "code-area"))
    text(data["scope"])
    text(data["rationale"])
    topics = rows(data["topics"], nonempty=True)
    available = {t["id"] for t in topics}
    if data["report"] == "overall" and available != set(TOPICS):
        raise ValueError("overall reports require all nine topics, including justified not_applicable topics")
    if data["report"] != "overall" and available != {data["report"]}:
        raise ValueError("standalone records require exactly their named topic")
    applicable = set()
    for topic in topics:
        fields(topic, ("id", "rationale", "coverage", "cards"))
        choice(topic["id"], TOPICS)
        text(topic["rationale"])
        for card in array(topic["cards"]):
            text(card)
        for item in rows(topic["coverage"], nonempty=True):
            fields(item, ("id", "status", "reason"))
            choice(item["status"], ("complete", "not_applicable", "pending", "partial", "unavailable"))
            text(item["reason"])
            if item["status"] != "not_applicable":
                applicable.add(topic["id"])
    sources = []
    findings = rows(data["findings"], pattern=r"F[1-9]\d*")
    for finding in findings:
        fields(finding, ("id", "title", "cause", "status", "trigger", "evidence", "analysis", "assessments",
                         "owner", "correction", "verification", "sources"))
        for key in ("title", "cause", "trigger", "analysis", "owner", "correction", "verification"):
            text(finding[key])
        choice(finding["status"], ("introduced", "worsened", "existing-in-scope"))
        if data["mode"] == "changeset" and finding["status"] == "existing-in-scope":
            raise ValueError(f"{finding['id']}: unchanged debt cannot count in a changeset")
        evidence(finding["evidence"])
        sources.extend(array(finding["sources"], nonempty=True))
        for assessment in rows(finding["assessments"], id_key="topic", nonempty=True):
            fields(assessment, ("topic", "severity", "consequence"), ("systemic", "reach"))
            choice(assessment["topic"], applicable)
            choice(assessment["severity"], SEVERITIES)
            text(assessment["consequence"])
            systemic = assessment.get("systemic", False)
            if type(systemic) is not bool or (systemic and assessment["severity"] != "major"):
                raise ValueError("systemic must be boolean and only applies to major")
            if systemic or assessment["severity"] == "critical":
                text(assessment.get("reach"))
            elif "reach" in assessment:
                text(assessment["reach"])
    unique(sources)
    by_id = {f["id"]: f for f in findings}
    for decision in rows(data["decisions"], pattern=r"D[1-9]\d*"):
        fields(decision, ("id", "title", "disposition", "topics", "reason", "evidence"), ("finding_id",))
        choice(decision["disposition"], ("rejected", "unresolved", "merged", "related"))
        topic_refs(decision["topics"], TOPICS if decision["disposition"] == "related" else available)
        text(decision["title"])
        text(decision["reason"])
        evidence(decision["evidence"])
        if decision["disposition"] == "merged":
            target = decision.get("finding_id")
            choice(target, by_id)
            if not set(decision["topics"]) <= {a["topic"] for a in by_id[target]["assessments"]}:
                raise ValueError("merged decision topics must be covered by its canonical finding")
        elif "finding_id" in decision:
            raise ValueError("only merged decisions reference a finding_id")
        if decision["disposition"] == "unresolved" and not set(decision["topics"]) <= applicable:
            raise ValueError("unresolved material evidence conflicts with not_applicable coverage")
    for strength in rows(data["strengths"], pattern=r"G[1-9]\d*"):
        fields(strength, ("id", "topics", "description", "evidence"))
        topic_refs(strength["topics"], applicable)
        text(strength["description"])
        evidence(strength["evidence"])
    for safeguard in rows(data["safeguards"], pattern=r"S[1-9]\d*"):
        fields(safeguard, ("id", "topics", "risk", "evidence", "nonroutine"))
        topic_refs(safeguard["topics"], applicable)
        text(safeguard["risk"])
        evidence(safeguard["evidence"])
        if type(safeguard["nonroutine"]) is not bool:
            raise ValueError("nonroutine must be boolean")
    for check in array(data["checks"]):
        fields(check, ("kind", "command", "result"))
        choice(check["kind"], ("executed", "supplied", "not_run"))
        text(check["command"])
        text(check["result"])
    for limit in array(data["limits"]):
        text(limit)
    if data["report"] == "overall" or "maintainability" in data:
        profile = data.get("maintainability")
        fields(profile, ("rows", "cleanliness", "structure", "extensibility"))
        dimensions = rows(profile["rows"], id_key="dimension")
        if {d["dimension"] for d in dimensions} != set(DIMENSIONS):
            raise ValueError("maintainability requires all seven dimensions")
        for row in dimensions:
            fields(row, ("dimension", "rating", "confidence", "evidence", "limits"))
            choice(row["rating"], ("Weak", "Adequate", "Strong", "Not demonstrated"))
            choice(row["confidence"], ("High", "Medium", "Low"))
            text(row["evidence"])
            text(row["limits"])
        for key in ("cleanliness", "structure", "extensibility"):
            text(profile[key])


def assessments(finding, topic):
    return [a for a in finding["assessments"] if topic == "overall" or a["topic"] == topic]


def severity(finding, topic):
    return min((a["severity"] for a in assessments(finding, topic)), key=SEVERITIES.index)


def evidence_text(items):
    return "; ".join(f"{e['kind']}: {e['reference']} — {e['text']}" for e in items)


def project(data, topic):
    """Project validated canonical causes; never add together card counts or grades."""
    selected = [t for t in data["topics"] if topic == "overall" or t["id"] == topic]
    if not selected:
        raise ValueError(f"topic is absent from the record: {topic}")
    coverage = [{**c, "id": t["id"] + "/" + c["id"]} for t in selected for c in t["coverage"]]
    decisions = [d for d in data["decisions"] if topic == "overall" or topic in d["topics"]]
    coverage.extend({"id": "unresolved/" + d["id"], "status": "partial", "reason": d["reason"]}
                    for d in decisions if d["disposition"] == "unresolved")
    findings = []
    for finding in data["findings"]:
        scoped = assessments(finding, topic)
        if not scoped:
            continue
        level = severity(finding, topic)
        strongest = [a for a in scoped if a["severity"] == level]
        findings.append({"id": finding["id"], "severity": level, "cause": finding["cause"],
                         "evidence": evidence_text(finding["evidence"]), "correction": finding["correction"],
                         "systemic": any(a.get("systemic", False) for a in strongest),
                         "reach": "; ".join(a["reach"] for a in strongest if "reach" in a)})
    ledger = {"scope": data["scope"], "coverage": coverage, "findings": findings}
    for key in ("strengths", "safeguards"):
        ledger[key] = [{**row, "evidence": evidence_text(row["evidence"])} for row in data[key]
                       if topic == "overall" or topic in row["topics"]]
    return ledger, calculate(ledger)
