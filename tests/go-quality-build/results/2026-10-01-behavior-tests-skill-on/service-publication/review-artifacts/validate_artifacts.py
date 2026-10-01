import json, pathlib
base = pathlib.Path("/private/tmp/go-independent-review-kolmdafa/output")
records = json.loads((base / "checks.json").read_text())
expected = {"setup":0, "toolchain":0, "baseline":1, "baseline-local-listeners":0, "race-shuffle-repeat":0, "setup-mutations":0, "mutation-in-place-publication":1, "mutation-accept-blank-fields":1, "mutation-ignore-late-read-error":1, "mutation-omit-response-close":1, "mutation-accept-trailing-json":1, "minimum-api-vet":0, "sources-preserved":0, "linux-compile":0, "windows-compile":0}
by_label = {record["label"]: record for record in records}
assert len(records) == len(by_label), "Duplicate check records"
for label, code in expected.items():
    assert by_label[label]["exit_code"] == code, label
    assert not by_label[label]["timed_out"], label
report = (base / "review.md").read_text()
for heading in ("## Testing — A+", "## Correctness & Compatibility — A+", "## Architecture & Design — A"):
    assert heading in report, heading
assert json.loads((base / "source-hashes-before.json").read_text()) == json.loads((base / "source-hashes-after.json").read_text())
print("All 15 named verification records, report-card headings, and source-preservation hashes validated.")

