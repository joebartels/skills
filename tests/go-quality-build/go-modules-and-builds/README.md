# Module and build evaluation tasks

These implementation tasks compare candidate module/build guidance with an explicitly chosen baseline. They are not reviews or individual effectiveness tests of existing skills. Keep fixture bytes, model settings, tools and any neighboring skills equal across comparison arms; label a no-guidance comparison separately from adding guidance to an existing catalog. Passing this checker alone does not establish a skill's usefulness.

Give an author the prompt and exact `files` inventory from `evals/evals.json`. Withhold `expected_output`, assertions and `probes/`. Registry sources are read-only dependency releases. Starting fixtures contain incomplete requested features and are expected to fail the new-contract probes.

## Prepare and check

Use `build_proxy(registry, destination)` from `check.py` to supply the offline dependency proxy during author runs. Set `GOPROXY` to its returned URI and `GOSUMDB=off` for these synthetic modules, with private module/build caches. These fixture settings are not production dependency policy.

For `release-inputs`, copy the complete project to its own Git repository and stage the initial files, then run `make preview` before the author starts. This models a normal development tree with ignored preview assets. Keep its index when checking: deliberately staged ignored inputs belong to its source archive. Other cases need no warm preparation.

Check a completed tree with an actual supported Go executable:

```sh
python3 tests/go-quality-build/go-modules-and-builds/check.py workspace-selection /path/to/candidate --go /path/to/go --json /tmp/module-check.json
```

The checker uses disposable copies and fresh caches without changing the candidate or global Go configuration. `--go` selects the toolchain; automatic switching is disabled. Run each promised version separately. Test/generator code executes cooperatively in temporary directories; inspect untrusted executable inputs before running them.

| Cases | Checks |
| --- | --- |
| `workspace-library`, `replacement-transfer`, `workspace-selection` | Workspace tests, standalone library tests with unchanged metadata, and an external consumer replacing only the candidate library. `workspace-selection` additionally requires make and runs its documented root `make test` and `make build` for the companion tool. Keep the complete tree, including companion modules and registry. Missing library sums can break standalone tests while consumer installation still succeeds. |
| `generated-targets` | Existing committed tables, clean and repeated regeneration, both tags, a valid schema-source variation, and Linux amd64/arm64 compilation of committed and regenerated tables. Only the dedicated schema probe runs after varying its input. Cross-compilation does not execute target binaries. |
| `release-inputs` | Warm preview tests and the actual source archive's API/CLI behavior. Export inventory includes tracked and unignored intended source, excludes warm ignored assets, and never inherits an ancestor repository's index. Requires Git and make. |
| `toolchain-variants` | Feature and preserved behavior in the implementation selected by the supplied toolchain. Run both supported series to exercise both source paths. |
| `pure-control` | Local calculation and the module skill's non-selection boundary. |
