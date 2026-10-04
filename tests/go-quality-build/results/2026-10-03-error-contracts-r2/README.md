# Verify this archive

From the repository root, run:

```sh
rtk proxy python3 -B tests/go-quality-build/verify_error_contract_archive.py
```

The read-only checker verifies the original seal, archived inputs and guidance,
patch reconstruction, review mappings and recorded results. It prints a fresh
result without replacing the historical `integrity.json`.

`controller.py` and `verify_integrity.py` are sealed snapshots of the original
execution tools. Their machine-local paths describe that execution. External
guidance hashes are historical attestations; contents absent from the archive
are not verified by the portable checker. Recorded Go tests are not rerun.
