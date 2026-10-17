# SDK resume cursors

Add `ResumeToken(topic string, offset uint64) (string, error)` to `lib` so clients can persist a subscription position and resume it later. Use the streamwire cursor format already understood by `streamctl`. A successful token preserves the topic and the full unsigned offset; invalid topics return an error with context that wraps `streamwire.ErrTopic`.

Add focused public behavior tests. Existing SDK operations must keep their behavior.

- [SDK usage and installation](lib/README.md)
- [Workspace development](docs/development.md)
- [Cursor inspection tool](tools/streamctl/README.md)
