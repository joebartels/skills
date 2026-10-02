import hashlib, json, pathlib, shutil
base = pathlib.Path("/private/tmp/go-independent-review-kolmdafa")
source = base / "candidate"
dest = base / "verification"
shutil.copytree(source, dest)
hashes = {str(f.relative_to(base)): hashlib.sha256(f.read_bytes()).hexdigest() for area in ("original", "candidate") for f in sorted((base / area).rglob("*")) if f.is_file()}
(base / "output" / "source-hashes-before.json").write_text(json.dumps(hashes, indent=2) + "\n")
print("Created unmodified verification copy; recorded SHA-256 hashes for original and candidate.")

