import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

module = Path("/private/tmp/go-combined-author-vsgga4k8/task-4/module")
report = Path("/private/tmp/go-combined-author-vsgga4k8/task-4/report")
source = module / "index.go"
original = source.read_bytes()
original_hash = hashlib.sha256(original).hexdigest()
text = original.decode()
offset = text.index("func replaceFile(path string, data []byte) error {")
mutated = text[:offset] + "func replaceFile(path string, data []byte) error {\n\treturn os.WriteFile(path, data, 0600)\n}\n"
command = ["rtk", "proxy", "env", "GOCACHE=/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN=local", "go", "test", "-timeout=20s", "-run=^TestPartialWriteFailureRetainsPriorData$", "."]
started = time.monotonic()
try:
    source.write_text(mutated)
    result = subprocess.run(command, cwd=module, env=os.environ.copy(), capture_output=True, text=True, timeout=30)
finally:
    source.write_bytes(original)
entry = {
    "command": command,
    "cwd": str(module),
    "environment": {"GOCACHE": "/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN": "local"},
    "execution_context": "workspace sandbox; synthetic transport and child-local file-size limit",
    "stdout": result.stdout,
    "stderr": result.stderr,
    "exit_code": result.returncode,
    "elapsed_seconds": time.monotonic() - started,
    "expected_assertion_failure": True,
    "controlled_mutation": "replace atomic staging/rename with direct os.WriteFile to the live destination",
    "source_sha256_before": original_hash,
    "source_sha256_after_restoration": hashlib.sha256(source.read_bytes()).hexdigest(),
}
log = report / "checks.json"
records = json.loads(log.read_text())
records.append(entry)
log.write_text(json.dumps(records, indent=2) + "\n")
print(json.dumps({**entry, "stdout": entry["stdout"][:1400]}))
assert result.returncode != 0, "retention mutation unexpectedly passed"
assert "--- FAIL: TestPartialWriteFailureRetainsPriorData" in result.stdout, "failure was not an assertion from the retention test"
for mode in ("put", "csv", "refresh"):
    assert f"--- FAIL: TestPartialWriteFailureRetainsPriorData/{mode}" in result.stdout
assert original_hash == entry["source_sha256_after_restoration"], "source restoration failed"
