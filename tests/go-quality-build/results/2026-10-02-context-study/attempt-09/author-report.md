Updated `inRange` to include `high` and added a focused regression test for the upper boundary and the value just above it. The API, Go 1.22 minimum, and small implementation are preserved.

**Skills used:** `go-behavior-tests`. The API-contract skill was not applicable because `inRange` is private.

**Check:** `rtk go test ./...` passed (2 tests, 1 package).