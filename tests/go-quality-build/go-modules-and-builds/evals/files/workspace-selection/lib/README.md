# Streamkit SDK

`example.org/streamkit` is a Go library for stream service clients. Install it into your application with:

```sh
go get example.org/streamkit@latest
```

Go 1.22 and later are supported. The SDK is published from this directory; applications import `example.org/streamkit` without the repository's development tools.

```go
path, err := streamkit.TopicPath("orders.eu-west")
```

`TopicPath` returns `/topics/` followed by the topic. Topic names contain dot-separated labels, each starting with an ASCII lowercase letter and continuing with lowercase letters, digits, or hyphens. The complete name is between 1 and 128 bytes. Names are case-sensitive and are never trimmed or rewritten. Invalid topics return contextual errors wrapping `streamwire.ErrTopic`.

The service's opaque resume cursors use streamwire's `EncodeCursor` and `DecodeCursor` protocol. Consumers store tokens unchanged; offsets range from zero through the largest `uint64`.
