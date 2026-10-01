# Fielddesk

Fielddesk is a small service with two features. Check-ins are submitted by field staff and stored as individual text files. Bulletins are read by staff from a local snapshot. One server process currently owns both HTTP routes.

## Supported caller contract

Go callers use `fielddesk.New(root)` to obtain a service, `SubmitCheckin(ctx, id, note)` to record a check-in, `Bulletins(ctx)` to read the current bulletin snapshot, and `Handler()` to serve HTTP. Keep those symbols and signatures usable by existing callers. `SubmitCheckin` accepts lowercase letters, digits and hyphens in a nonempty ID and a nonblank note. It writes `note` followed by one newline to `root/checkins/<id>.txt`; a duplicate ID replaces that file. Bad input returns an error without writing a file. `Bulletins` returns an empty, non-nil slice when the snapshot is absent.

`POST /v1/checkins` accepts JSON `{"id":"...","note":"..."}`. A successful response is status 201 with JSON `{"id":"..."}` followed by a newline. Invalid check-ins receive status 400. `GET /v1/bulletins` returns status 200, content type `application/json`, and a JSON array of objects with `id` and `text` fields; an absent snapshot returns `[]` followed by a newline. Other methods on these routes return 405. Keep these request and response shapes while evolving the service.

The snapshot is `root/bulletins.json`, a JSON array of `Bulletin` values. Operators currently place it there outside the service. Existing deployments set `FIELDDESK_ROOT` for `cmd/server` (default `./data`). The server's HTTP routes and the Go API above are in use by separate consumers.

## Requested evolution

The new change is described in the evaluation prompt. There is no required package tree: separate responsibilities only where the new consumers and independently changing integration warrant it. The existing Go and HTTP contract remains supported.
