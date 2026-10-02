---
name: go-values-and-zero-values
description: Use when Go work changes zero/default or nil/empty state, borrowed or owned values, copying, aliasing, receivers, or interface conversion. Skip pure calculations and unchanged representations, even when they use structs or slices.
---

# Go values and zero values

Follow state through creation, mutation, copying and retention. A copied Go value can still share most of its mutable data.

Read the affected types, creation paths and consumers. Identify which values are borrowed, shared or independently owned; when borrowing expires; and what nil, empty, absent and explicitly zero mean. A useful zero value is desirable when the invariants permit it. Preserve supported zero behavior, but retain required construction when validation or mandatory inputs make it necessary.

| Changed state | Decision to make |
| --- | --- |
| Absent versus explicit zero | Represent presence when the meanings differ. Do not substitute a default merely because the supplied value is zero. |
| Nil versus empty collection | Choose from observable semantics, including the actual encoder and tags. Nil-map reads work; writes require an initialized map. Appending to a nil slice works. |
| Slice or map copy | Assignment shares storage. Cloning copies elements, not mutable references inside those elements. |
| Snapshot, fork or retained input | Copy exactly the reachable mutable state the ownership contract requires; document any intentionally shared state. |
| Borrowed view | Preserve intentional borrowing. Copy at retention when the producer may reuse the storage; do not add copies before every inspection. |
| Receiver or interface conversion | Check mutation, method sets, nil behavior and copy-sensitive state at real use sites. |

For a snapshot, enumerate reference-containing fields rather than stopping at the outer struct: slices, maps, pointers and interfaces may retain shared state. Setting capacity equal to length with `s[:len(s):len(s)]` prevents a nonempty append from overwriting spare capacity in the original backing array; existing elements still alias. `slices.Clone` and `maps.Clone` are shallow; nilness is preserved, but nested references still alias. If a graph's contract preserves shared targets or cycles, copying must preserve that internal relationship while separating the owned state from the original. Do not blindly clone external resources or synchronization objects.

For example, a snapshot of `[]Entry` where each entry contains `map[string][]byte` may need an outer slice copy, a map per entry and independent byte payloads. Preserve nil and non-nil empty values if their distinction is promised. A read-only borrowed view of the same entries may need no copy at all. The required behavior, rather than a universal deep-copy rule, determines the work.

Choose pointer/value receivers from semantics. A value receiver copies its fields, but mutations through a map, slice or pointer field still reach shared state. Pointer receivers can change whether `T` satisfies an interface; automatic address-taking for an addressable local variable can conceal that break. Compile a relevant interface assignment, method expression or non-addressable use such as a map index when that surface changes. Avoid fixed size thresholds and unsupported allocation claims.

Do not copy a used mutex or other type whose documented contract forbids copying. Inspect assignment, return, embedding and container paths as well as receivers; vet detects some copies but does not certify ownership. For interface nilness and concrete examples, see [value mechanics](references/value-mechanics.md).

Verify both directions of promised independence: mutate the returned copy and the original, including nested values and append paths. For retained borrowed buffers, reuse the producer's storage after capture. Exercise absent/default/explicit-zero creation and actual nil/empty serialization where affected. Verify required-constructor controls and relevant method sets. Use race checks when concurrent use is part of the contract; a passing race run does not establish a new safety guarantee.

This skill owns representation and value ownership. API contracts own supported evolution; composition owns construction design and resource lifetime; concurrency owns synchronization design; performance owns optimization evidence. Error success behavior belongs to error contracts. Apply the needed decisions without requiring another skill or redesigning unrelated code.
