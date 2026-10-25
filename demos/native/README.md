# Zero-copy serialization (`zerocpy`)

```
zerocpy.String: [5 104 101 108 108 111]
composite.String:    [5 104 101 108 108 111]

processing "first"
processing "second"

aliased strings after the loop: ["secon" "second"]
copied strings after the loop:  ["first" "second"]

decoded bytes: [1 2 3]
after modifying the payload: [9 2 3] (the decoded slice changed)
composite.ByteSlice copied [9 2 3] and is unaffected by the payload change to [3 7 2 3]
zerocpy.Uint64 round trips: 0 (8 bytes), 40 (8 bytes), 1099511627776 (8 bytes), 18446744073709551615 (8 bytes)
native array: [7 8 9] → [7 8 9] (13 bytes)
```

The `zerocpy` package uses package `unsafe` for direct memory access. Two
consequences:

1. **Same bytes, faster code.** `zerocpy.String` and `composite.String` produce
   identical payloads, so the two implementations are interchangeable on the
   wire. Numbers are decoded with a single load.
2. **Decoded strings and byte slices alias the input buffer.** No copy is made,
   so the value is only valid while the buffer holds that data.

## The aliasing pitfall

The example simulates a reader that fills one buffer per message and keeps the
decoded strings:

```go
n := copy(buf, binco.Encode(zerocpy.String, msg)) // the same buffer every time
str, _, _ := zerocpy.String.Unmarshal(buf[:n])
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
* `zerocpy.NewArraySerializer[T, V]` is the only array serializer, because decoding into
  an array means reinterpreting the decoded elements as the array type. It
  rejects payloads whose length differs from the array type.
* The package assumes little-endian byte order and panics at init on a
  big-endian host (`binco.ErrUnsupportedArch`), so it can never silently emit a
  payload that violates the format.
* Generating code with `genopts.WithUnsafe()` (or `WithNotUnsafe()`, which uses
  `zerocpy` for everything except strings) picks these serializers automatically.
