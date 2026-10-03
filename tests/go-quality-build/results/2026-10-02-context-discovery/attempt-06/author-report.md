Repaired the upper-bound check in [range.go](/private/tmp/go-task-6nj9j2nc/range.go) and updated [range_test.go](/private/tmp/go-task-6nj9j2nc/range_test.go) to cover the inclusive upper endpoint, interior, and outside values. The signature and Go 1.22 minimum are unchanged.

**Check:** `rtk go test ./...` passed.

**Skill used:** `go-behavior-tests`. I skipped `go-api-contracts` because this private helper’s signature and public API are unchanged.