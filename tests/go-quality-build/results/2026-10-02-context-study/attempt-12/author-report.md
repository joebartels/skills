Repaired `inRange` to include the upper bound and added a focused regression test for `value == high`. The function signature and Go 1.22 minimum are unchanged.

**Skill used:** `go-behavior-tests`.

**Checks:** `rtk go test ./...` passed. `git diff` checks could not run because this directory is not a Git repository.