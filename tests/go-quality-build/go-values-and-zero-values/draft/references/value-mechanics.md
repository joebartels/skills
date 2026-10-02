# Value mechanics that affect contracts

Use these facts when the corresponding state changes. They are not prescriptions to convert every API to pointers, initialize every slice, or copy every input.

## Nil, empty and presence

A nil slice has length and capacity zero and can be ranged over or appended to. A non-nil empty slice has the same length but can produce a different external representation. With ordinary `encoding/json` slice fields, nil produces `null` and non-nil empty produces `[]`; `omitempty`, custom encoders and `[]byte` change the question. Assert the actual promised bytes instead of inferring them from a broad slice rule.

A nil map permits reads, lookup, range and delete. Assignment into it panics. Initialize it where mutation requires storage; that does not imply every map-returning function must return an empty map.

When zero means both “unset” and a meaningful configuration, keep presence separately. A private presence flag, pointer, option or existing schema mechanism may fit. Required inputs can still justify a constructor. Choose the smallest representation that expresses the real invariant.

## Copies and borrowed storage

Copying a slice copies its descriptor, leaving its backing array shared. A full slice expression such as `s[:len(s):len(s)]` limits capacity but still shares the existing elements. Copying a map value retains the same map; copying a pointer retains the same pointee.

Clone helpers copy elements by assignment. For `[]struct{ Labels map[string]string }`, cloning the slice leaves each Labels map shared. For `map[string][]byte`, cloning the map leaves the byte arrays shared. Copy those nested values only if the contract promises independence. Preserve internal identity when callers depend on shared targets or cycles; a memo of original-to-copy nodes is one possible implementation, not a required API.

A producer can return a borrowed byte view valid only until its next call. Inspect or filter while it is valid, and copy accepted bytes before retention. A documented borrowed predicate can receive the original view; copying rejected input may add unnecessary work. A zero-length copy made with `append([]byte(nil), input...)` turns non-nil empty input into nil; use a representation-preserving copy when nilness matters.

## Receivers, locks and interfaces

Methods on `T` belong to the method sets of both `T` and `*T`; methods on `*T` belong to `*T`. A compiler may take an address for a call on an addressable `T`, but that does not give `T` the pointer method set. A map index or temporary value may reveal the difference. Check actual interfaces and function/method types before a receiver change.

A value receiver containing a map or slice does not make that data immutable. A type containing a used `sync.Mutex` must not be copied; moving it through an assignment, return or container can violate that contract even when every method has a pointer receiver. Honor other copy restrictions such as `strings.Builder`'s; the presence of any pointer or channel alone does not make copying forbidden.

An interface is nil only when it holds neither a dynamic type nor a value:

```go
var concrete *Problem = nil
var err error = concrete // non-nil interface, if *Problem implements error
```

A nil receiver may be a valid documented implementation. Do not use a generic reflection-based “nil normalization” wrapper. For successful error returns, explicitly return a nil error interface rather than converting a nil error pointer.

These facts follow the [Go specification](https://go.dev/ref/spec), [nil-error FAQ](https://go.dev/doc/faq#nil_error), [slices.Clone](https://pkg.go.dev/slices#Clone), [maps.Clone](https://pkg.go.dev/maps#Clone), [JSON documentation](https://pkg.go.dev/encoding/json#Marshal) and [sync contracts](https://pkg.go.dev/sync). Clone helpers require Go 1.21; respect the project's supported toolchain before choosing them.
