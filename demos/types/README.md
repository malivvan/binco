# Types

Serializing arrays, slices, maps and structs.

- [array](array/): serialize an array and skip it.
- [slice](slice/): serialize a slice, with the default and a custom length
  serializer.
- [map](map/): serialize a map.
- [struct](struct/): write a struct serializer by hand.

Every example follows the same shape:

```go
size := serializer.Size(v)          // exact number of bytes
bs := make([]byte, size)     // allocate once
n := serializer.Marshal(v, bs)      // no allocation inside Marshal
v, n, err := serializer.Unmarshal(bs)
```

`binco.Encode(serializer, v)` and `binco.Decode(serializer, bs)` are shorthands that allocate
for you; use them when buffer reuse does not matter.

## array

```
value:   [1 2 3]
size:    4 bytes
encoded: [3 2 4 6] (wrote 4)
decoded: [1 2 3] (read 4)
skipped: 4 bytes
shorter payload: binco: wrong length
```

An array is encoded exactly like a slice: `[length, elements...]`. Because the
array type has a fixed length, a payload with a different number of elements is
rejected (`binco.ErrWrongLength` when shorter, `binco.ErrTooLargeLength` when
longer) instead of being decoded into uninitialized memory.

## slice

```
value:     [1 2 3]
encoded:   [3 2 4 6] (4 bytes)
fixed len:   [3 0 0 0 0 0 0 0 2 4 6] (11 bytes)
nil=true → len=0, nil=false
nil=false → len=0, nil=false
```

The length is a positive varint by default. `slopts.WithLenSerializer` replaces it, for
example with the fixed-width `scalar.Int`. A nil slice and an empty slice encode
identically and decode to an empty, non-nil slice.

## map

```
value:    {1:"one", 2:"two", 3:"three"}
size:     18 bytes
encoded:  [3 6 5 116 104 114 101 101 2 3 111 110 101 4 3 116 119 111] (wrote 18)
decoded:  {1:"one", 2:"two", 3:"three"} (read 18)
```

Entries are written in Go's map iteration order, so the encoded bytes of a map
are not canonical - the exact byte sequence differs between runs. Decoding is
deterministic, and two payloads of the same map decode to equal maps. If a
byte-for-byte canonical payload matters, encode a sorted slice of key/value
pairs instead.

## struct

```
value:   {Str:hello, binco Sl:[1 2 3]}
size:    16 bytes
encoded: [11 104 101 108 108 111 44 32 98 105 110 120 3 2 4 6] (wrote 16)
decoded: {Str:hello, binco Sl:[1 2 3]} (read 16)
skipped: 16 bytes
binco.Marshal: [11 104 ... 3 2 4 6]
truncated payload: binco: too small byte slice
```

A struct is the concatenation of its fields in declaration order. `serializer.go`
implements `binco.Serializer[Foo]` by hand and also gives `Foo` the
`MarshalBinco`/`SizeBinco` methods, which enables `binco.Marshal(foo)`.

Field order is the wire format: never reorder or remove a field of a type that
is already on the wire. Add a new version and a DTM instead (see
[typed/versioning](../typed/versioning)).
