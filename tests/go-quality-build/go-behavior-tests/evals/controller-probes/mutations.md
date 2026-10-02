# Frozen semantic mutations

Freeze these meanings before author dispatch. Adapt the smallest candidate-specific
source patch after inspecting its implementation; preserve the exact patch and
clean/compiled/test outcome. Run candidate tests alone, without controller probes.
Compile errors and uninterpretable hangs do not count as regression detection.

| ID | Case | Semantic regression |
| --- | --- | --- |
| codec-coordinated-swap | library-codec | Encode emits Name/Region and Decode interprets Name/Region: round trips still work but independent wire contracts break. |
| cli-success-exit | cli-partial-failure | main prints the real failure but exits zero instead of status 2. |
| cli-discard-prefix | cli-partial-failure | A later invalid row causes accepted-prefix files to be discarded or never published. |
| mirror-no-label-validation | service-publication | Omit Label nonblank rejection independently from Code validation. |
| mirror-direct-truncate | service-publication | Publish the otherwise-correct new snapshot by truncating/writing the existing path rather than replacing it. |
| host-no-join | worker-host | Cancel work but call resource release/return before waiting for callback cleanup. |
| worker-start-relative | worker-host | Use ticker/start-relative recurrence so a slow callback's next cycle can start without a full post-completion interval. |

The old-open-handle publication observation assumes supported Unix replacement
semantics. A skipped Windows observation is not evidence of Windows behavior.
