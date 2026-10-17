# Catalog lookup names

Implement `CanonicalName(name string) (string, error)` in `lib`, using the supplied `example.org/keytext` canonical-name policy. Return its canonical lookup name and preserve its validation error as a wrapped cause. Add focused public behavior tests for ordinary names, whitespace, and invalid input.

`lib` is the independently published `example.org/catalog` library. Preserve `Label` and its behavior, retain Go 1.22.0 support, and keep the local workspace useful for development. Approved dependency releases are `example.org/keytext` v1.0.0 and v1.1.0; their sources are supplied under `registry`. Do not add other dependencies or change the dependency's policy.

Names consist of ASCII letters and digits separated by whitespace. Canonical names lowercase letters, collapse whitespace to a hyphen, and reject empty names and other characters. For example, `"  North   Pier 7 "` becomes `"north-pier-7"`.

Run existing and added tests from `lib`. The workspace includes the local keytext development module.
