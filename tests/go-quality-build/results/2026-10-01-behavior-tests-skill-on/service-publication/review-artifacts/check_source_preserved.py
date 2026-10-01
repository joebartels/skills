import hashlib, json, pathlib
base = pathlib.Path("/private/tmp/go-independent-review-kolmdafa")
before = json.loads((base / "output" / "source-hashes-before.json").read_text())
after = {str(f.relative_to(base)): hashlib.sha256(f.read_bytes()).hexdigest() for area in ("original", "candidate") for f in sorted((base / area).rglob("*")) if f.is_file()}
(base / "output" / "source-hashes-after.json").write_text(json.dumps(after, indent=2) + "\n")
assert before == after, "original or candidate changed"
print(f"All {len(after)} original/candidate source files retain exact SHA-256 hashes.")

