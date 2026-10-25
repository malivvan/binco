# Zero-copy serialization (`zerocopy`)

```
zerocopy.String: [5 104 101 108 108 111]
composite.String:    [5 104 101 108 108 111]

processing "first"
processing "second"

aliased strings after the loop: ["secon" "second"]
copied strings after the loop:  ["first" "second"]

decoded bytes: [1 2 3]
after modifying the payload: [9 2 3] (the decoded slice changed)
composite.ByteSlice copied [9 2 3] and is unaffected by the payload change to [3 7 2 3]
zerocopy.Uint64 round trips: 0 (8 bytes), 40 (8 bytes), 1099511627776 (8 bytes), 18446744073709551615 (8 bytes)
native array: [7 8 9] → [7 8 9] (13 bytes)
```

The `zerocopy` package uses package `unsafe` for direct memory access. Two
consequences:

1. **Same bytes, faster code.** `zerocopy.String` and `composite.String` produce
   identical payloads, so the two implementations are interchangeable on the
   wire. Numbers are decoded with a single load.
2. **Decoded strings and byte slices alias the input buffer.** No copy is made,
   so the value is only valid while the buffer holds that data.

## The aliasing pitfall

The example simulates a reader that fills one buffer per message and keeps the
decoded strings:

```go
n := copy(buf, binco.Encode(zerocopy.String, msg)) // the same buffer every time
str, _, _ := zerocopy.String.Unmarshal(buf[:n])
aliased = append(aliased, str)             // wrong: aliases buf
copied = append(copied, strings.Clone(str)) // right: an independent copy
```

Both strings end up pointing into `buf`, which by then holds `"second"`. The
first string has length 5, so it reads `"secon"`:

```
aliased strings after the loop: ["secon" "second"]
copied strings after the loop:  ["first" "second"]
```

Fixes, in order of cost:

* process the value before reading the next message (zero cost),
* copy it (`strings.Clone`, `bytes.Clone`),
* use `composite.String` / `composite.ByteSlice`, which copy while decoding.

## Other notes

* Byte slices have the same rule: modifying the payload changes the decoded
  slice.
* `zerocopy.NewArraySerializer[T, V]` is the only array serializer, because decoding into
  an array means reinterpreting the decoded elements as the array type. It
  rejects payloads whose length differs from the array type.
* Scalars are converted with an explicit little-endian order, so the package is
  correct on any host byte order and at an unaligned offset. The array serializer
  checks at construction that its type parameters really describe an array of the
  element type, and panics with `binco.ErrWrongArrayType` if they do not.
* Generating code with `genopts.WithUnsafe()` (or `WithNotUnsafe()`, which uses
  `zerocopy` for everything except strings) picks these serializers automatically.
