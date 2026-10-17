# Module and build evaluation tasks

These fixtures exercise separately consumed libraries, workspace and replacement resolution, generated inputs, supported build variants, and a local-calculation non-selection control. They contain incomplete requested features; unchanged starting fixtures are expected to fail the new-contract probes.

Give an author the case prompt, its exact `files` inventory from `evals/evals.json`, and the same selected neighboring-skill catalog across comparison arms. Withhold `expected_output`, assertions and `probes/`. Registry sources are read-only dependency releases; local development modules model legitimate workspace or replacement behavior.

The checker constructs an offline file module proxy from `registry/<name>/<version>/`, uses fresh temporary module/build caches, and tests disposable copies. For author runs, `build_proxy(registry, destination)` in `check.py` supplies the same proxy format; set `GOPROXY` to its returned URI and `GOSUMDB=off` for these synthetic modules. Those settings apply to fixtures, not production dependency policy.

Check a completed fixture tree with the actual supported Go executable:

```sh
python3 tests/go-quality-build/go-modules-and-builds/check.py workspace-library /path/to/candidate --go /path/to/go --json /tmp/module-check.json
```

Other case IDs are `replacement-transfer`, `generated-targets` and `pure-control`. Library trees retain their `lib/`, `go.work`, development module and registry. The checker distinguishes workspace tests, standalone library tests and a fresh external consumer that replaces only the candidate library. Generation checks verify committed tables before removing them in a disposable copy, regenerate both build variants, compare repeated output bytes and compile Linux amd64/arm64 tests. A valid schema change must also appear in both regenerated variants; only the dedicated schema probe runs under that changed input. Cross-compilation does not execute those target binaries.

`--go` selects the toolchain; automatic switching and user Go configuration are disabled. Run each promised version separately. Test/generator code executes cooperatively in temporary directories; this checker is not a security sandbox. Inspect untrusted executable inputs before running them.
