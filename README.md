# binco [![Go Reference](https://pkg.go.dev/badge/github.com/malivvan/binco.svg)](https://pkg.go.dev/github.com/malivvan/binco) [![Release](https://img.shields.io/github/v/release/malivvan/binco.svg?sort=semver)](https://github.com/malivvan/binco/releases/latest) ![test](https://github.com/malivvan/binco/workflows/ci/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/malivvan/binco/badge.svg)](https://coveralls.io/github/malivvan/binco) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`binco` is a binary codec toolkit for Go. It gives you codecs for every
built-in type, a zero-copy fast path, pointer mapping for graphs, type
identifiers for polymorphic data, a streaming API, and a code generator that
writes serializers for your own types.

The format it implements is schema-driven: sender and receiver share the same
type definitions, so a payload contains values only - no field names, no type
tags, no alignment padding. Everything the receiver needs is in the code, not in
the bytes:

```go
type Foo struct {
	A int
	B bool
	C string
}

Foo{A: 300, B: true, C: "hi"}    ->    [216 4 1 2 104 105]
```

That is the whole idea. The rest of this document explains how to use it for
every kind of data you are likely to serialize. If you are working on the library
itself, [`AGENTS.md`](./AGENTS.md) covers the architecture, conventions and
commands.

## Contents

**Getting started**

* [Why binco](#1-why-binco)
* [Install and requirements](#2-install-and-requirements)
* [Core concepts](#3-core-concepts)
  * [Codecs](#31-codecs)
  * [Buffers and streams](#32-buffers-and-streams)
  * [The schema is the contract](#33-the-schema-is-the-contract)
  * [Errors and panics](#34-errors-and-panics)
  * [Sizes skipping and the exactly-N-bytes invariant](#35-sizes-skipping-and-the-exactly-n-bytes-invariant)
* [Quick start](#4-quick-start)
  * [A number](#41-a-number)
  * [A string](#42-a-string)
  * [A struct by hand](#43-a-struct-by-hand)
  * [The same struct from the generator](#44-the-same-struct-from-the-generator)
  * [Over a stream](#45-over-a-stream)

**The format**

* [The wire format](#5-the-wire-format)
  * [Anatomy of a payload](#51-anatomy-of-a-payload)
  * [Booleans and bytes](#52-booleans-and-bytes)
  * [Varint](#53-varint)
  * [Fixed width](#54-fixed-width)
  * [Strings and byte slices](#55-strings-and-byte-slices)
  * [Slices and arrays](#56-slices-and-arrays)
  * [Maps](#57-maps)
  * [Pointers](#58-pointers)
  * [Time](#59-time)
  * [Type identifiers](#510-type-identifiers)
  * [A payload dissected](#511-a-payload-dissected)
  * [Compatibility rules](#512-compatibility-rules)
* [Choosing codecs](#6-choosing-codecs)

**Reference**

* [Package guide](#7-package-guide)
  * [binco](#71-binco)
  * [varnum](#72-varnum)
  * [scalar](#73-scalar)
  * [composite](#74-composite)
  * [zerocopy](#75-zerocopy)
  * [ptrmap](#76-ptrmap)
  * [typeid](#77-typeid)
  * [config](#78-config)
  * [wirefmt](#79-wirefmt)

**Use cases**

* [Use cases](#8-use-cases)
  * [Numbers and enums](#81-numbers-and-enums)
  * [Strings and byte blobs](#82-strings-and-byte-blobs)
  * [Structs by hand](#83-structs-by-hand)
  * [Structs from the generator](#84-structs-from-the-generator)
  * [Collections](#85-collections)
  * [Optional fields and nil](#86-optional-fields-and-nil)
  * [Shared pointers and identity](#87-shared-pointers-and-identity)
  * [Cycles linked lists trees graphs](#88-cycles-linked-lists-trees-graphs)
  * [Polymorphism with interfaces](#89-polymorphism-with-interfaces)
  * [Versioning and migration](#810-versioning-and-migration)
  * [Streaming protocols](#811-streaming-protocols)
  * [Validating untrusted input](#812-validating-untrusted-input)
  * [Partial decoding and out-of-order reads](#813-partial-decoding-and-out-of-order-reads)
  * [Zero-copy hot paths](#814-zero-copy-hot-paths)
  * [Custom codecs](#815-custom-codecs)
  * [Composition recursion and generics](#816-composition-recursion-and-generics)
  * [Zero-allocation encoding](#817-zero-allocation-encoding)
  * [Concurrency](#818-concurrency)
  * [Testing your codecs](#819-testing-your-codecs)
  * [Interoperability](#820-interoperability)

**The generator**

* [Code generation](#9-code-generation)
  * [Getting started](#91-getting-started)
  * [Generator configuration](#92-generator-configuration)
  * [Generator methods](#93-generator-methods)
  * [Generated code and naming](#94-generated-code-and-naming)
  * [Serialization options](#95-serialization-options)
  * [Interfaces](#96-interfaces)
  * [Versioned types](#97-versioned-types)
  * [Multi-package and cross-package generation](#98-multi-package-and-cross-package-generation)
  * [Streaming mode](#99-streaming-mode)
  * [Workflow and troubleshooting](#910-workflow-and-troubleshooting)

**Operating it**

* [Performance guide](#10-performance-guide)
* [Correctness guarantees and security](#11-correctness-guarantees-and-security)
* [Error reference](#12-error-reference)
* [Cheat sheet](#13-cheat-sheet)
* [FAQ](#14-faq)
* [Working on binco](#15-working-on-binco)
* [License](#license)

## 1. Why binco

### Small payloads by construction

There is no per-field metadata, so the cost of a payload is the cost of its
values. Numbers use a variable-length encoding when that saves space, a
fixed-width encoding when speed matters more, and you pick per field. The same
struct above occupies five bytes, of which four carry actual information.

### Fast paths, chosen at compile time

Every encoding decision is static: which codec, which width, which length
prefix. The generator inlines the whole structure of a type, and the `zerocopy`
package turns scalar work into single machine-word reads and writes. Nothing in
the hot path uses reflection, interfaces you did not ask for, or map lookups.

### Zero allocations when you want them

`Size` reports the exact number of bytes a value needs, and `Marshal` writes
into a buffer you own. A steady-state encoder can therefore reuse one buffer and
allocate nothing at all. `binco.Encode` is the convenient version that allocates
for you.

### Complete coverage of Go's type system

Numbers of every width, `bool`, `byte`, `string`, `[]byte`, arrays, slices,
maps, pointers, `time.Time`, defined types, interfaces and multiple versions of
a type - in memory and over streams, with validation, skipping, and optional
zero-copy decoding.

### Predictable on the wire

The format is documented below byte by byte. Buffer and streaming codecs emit
identical bytes, the zero-copy codecs are byte-compatible with the portable
ones, and the generator's output is stable. Two processes using different codec
families still interoperate.

### What binco does not do

* No schema evolution magic: the schema lives in your code, so changing it is a
  code change with the usual compatibility rules (see
  [Compatibility rules](#512-compatibility-rules)).
* No framing: a payload has no length prefix or delimiter. If you need framing,
  add it yourself or use a type identifier, as shown in
  [Streaming protocols](#811-streaming-protocols).
* No reflection at run time: reflection exists only in the code generator.
* No canonical map order by default: a Go map iterates in a random order, so
  encoding one is not deterministic. `composite.NewSortedMapSerializer` writes
  the entries sorted by their encoded key bytes when the bytes must be
  reproducible; see [Collections](#85-collections).

## 2. Install and requirements

```bash
go get github.com/malivvan/binco@latest
```

v1.0.0 is the first stable release. The module follows
[semantic versioning](https://semver.org/spec/v2.0.0.html), and the wire format
described in [The wire format](#5-the-wire-format) is frozen for the v1 series.
Release notes live in [CHANGELOG.md](./CHANGELOG.md).

* Go 1.27.3 or newer.
* The module has **no third-party dependencies**: `go.mod` requires nothing, and
  there is no `go.sum` or `vendor` directory. The generator's import fixing and
  path validation live in `internal/` as small local implementations that replace
  `golang.org/x/tools` and `golang.org/x/mod` (see [`AGENTS.md`](./AGENTS.md)).

Add the import you need:

```go
import (
	"github.com/malivvan/binco"            // core: interfaces, helpers, errors, DTM
	"github.com/malivvan/binco/varnum"     // variable-length numbers
	"github.com/malivvan/binco/scalar"     // fixed-width numbers and time
	"github.com/malivvan/binco/composite"  // bool, string, []byte, pointers, slices, maps
	"github.com/malivvan/binco/zerocopy"    // zero-copy scalars and containers
	"github.com/malivvan/binco/ptrmap"     // pointer mapping
	"github.com/malivvan/binco/typeid"     // type identifiers
	"github.com/malivvan/binco/wirefmt"    // streaming API
)
```

## 3. Core concepts

### 3.1 Codecs

Everything in the module is a codec implementing one interface:

```go
type Serializer[T any] interface {
	Marshal(t T, bs []byte) (n int)
	Unmarshal(bs []byte) (t T, n int, err error)
	Size(t T) (size int)
	Skip(bs []byte) (n int, err error)
}
```

* `Marshal` fills `bs` and returns the number of bytes written. It panics with
  `binco.ErrTruncated` when `bs` is too small - a caller bug, not a data
  error.
* `Unmarshal` parses one value and never reads past it, so the rest of the
  buffer can be decoded by the next codec.
* `Size` returns exactly the number of bytes `Marshal` will write.
* `Skip` advances over a value without building it.

Codecs are values, not functions with hidden state. The exported ones are
package-level variables (`varnum.Int`, `composite.String`, `scalar.Uint64`,
`zerocopy.Int32`, ...) and constructors (`composite.NewSliceSerializer`, `ptrmap.NewPtrSerializer`,
`typeid.NewSerializer`, ...). Because the interface is generic, codecs compose: a slice
codec takes an element codec, a map codec takes a key and a value codec, a typed
codec wraps any of them.

```go
ints := composite.NewSliceSerializer(varnum.Int)                               // []int
rows := composite.NewSliceSerializer(composite.NewSliceSerializer(varnum.Int)) // [][]int
byName := composite.NewMapSerializer(composite.String, varnum.Int)             // map[string]int
```

### 3.2 Buffers and streams

The module ships two parallel APIs with the same wire format:

| | Buffer API | Streaming API |
|---|---|---|
| package | `binco` and its subpackages | `binco/wirefmt` and its subpackages |
| destination | `[]byte` you own | `io.Writer` |
| source | `[]byte` | `io.Reader` |
| interface | `Serializer[T]` | `wirefmt.Serializer[T]` |
| errors | only data errors | data errors **and** I/O errors |
| use when | the payload is a value (message, file, cache entry) | the payload is a flow (connection, pipe, large file) |

Bytes written by a buffer codec are read by the matching stream codec, and vice
versa. Only the plumbing differs; the format does not.

### 3.3 The schema is the contract

A payload is the concatenation of the encoded fields of a type, in declaration
order. Nothing marks where a field ends: the receiver's code knows. Three
consequences:

1. **Field order is part of the format.** Reordering struct fields changes the
   wire format. Add new versions instead (see
   [Versioning and migration](#810-versioning-and-migration)).
2. **Both sides must agree on every field's codec.** A `varnum.Int` field read
   by a `scalar.Int` parser produces garbage, not an error.
3. **You can drop data by not reading it.** Unknown trailing fields are not an
   error; a reader that knows fewer fields simply stops early.

The generator writes a `SchemaFingerprint` constant next to every struct: a
SHA-256 of the field names and types in declaration order. Comparing it with the
value that produced a payload turns a silent misread (two `int` fields swapped)
into a clear failure. See
[Compatibility rules](#512-compatibility-rules).

### 3.4 Errors and panics

`Marshal` panics with `binco.ErrBufferTooSmall` when the output buffer is too
small; that is the only panic in normal operation and it signals a bug in the
caller (`binco.Encode` sizes the buffer for you and never triggers it). Every
codec panics with that same sentinel rather than with a raw index error, so one
`recover` can recognise the case. `Unmarshal` and `Skip` never panic on malformed
input: they return one of the errors in the
[error reference](#12-error-reference), all prefixed with `binco: ` and all
comparable with `errors.Is`.
A value that exceeds `binco.MaxDepth` levels of nesting reports
`binco.ErrTooDeep` instead of exhausting the stack.

```go
name, err := binco.Decode(composite.String, bs)
switch {
case errors.Is(err, binco.ErrTruncated):
	// the payload ended inside the value
case errors.Is(err, binco.ErrNegativeLength):
	// a length field decoded as a negative number
}
```

Streaming codecs additionally return whatever the `Writer` or `Reader` returned,
unchanged and unwrapped, with one rule: `io.EOF` before the first byte of a value
is a clean end of stream, and `io.ErrUnexpectedEOF` after it is a truncated
message. The two are distinguishable.

`Skip` applies the same range checks as `Unmarshal`: a value that `Unmarshal`
would reject as out of range is not silently skipped either.

### 3.5 Sizes skipping and the exactly-N-bytes invariant

Three guarantees make zero-allocation encoding safe and let you index into a
payload without decoding it:

* `Size(v)` equals the number of bytes `Marshal(v, bs)` writes.
* `Unmarshal` consumes exactly `Size` bytes of the value it returns.
* `Skip` advances exactly as far as `Unmarshal` would.

The test suite asserts these invariants for every codec, including the
generated ones.

## 4. Quick start

### 4.1 A number

```go
package main

import (
	"fmt"

	"github.com/malivvan/binco/varnum"
)

func main() {
	num := 100

	// 1. Ask for the exact size, then encode into a buffer you own.
	bs := make([]byte, varnum.Int.Size(num))
	n := varnum.Int.Marshal(num, bs)
	fmt.Println(n, bs) // 2 [200 1]

	// 2. Decode it back.
	got, n, err := varnum.Int.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Println(got, n) // 100 2

	// 3. Advance over it without decoding.
	n, err = varnum.Int.Skip(bs)
	fmt.Println(n, err) // 2 <nil>
}
```

### 4.2 A string

```go
import "github.com/malivvan/binco/composite"

bs := binco.Encode(composite.String, "hello world")   // [11 104 101 ... 100]
text, err := binco.Decode(composite.String, bs)        // "hello world"
```

`binco.Encode` allocates the exact buffer and returns it; `binco.Decode` returns
the value and the error. Use them when convenience matters more than buffer
reuse.

### 4.3 A struct by hand

A codec for a struct is the concatenation of its fields' codecs:

```go
type Person struct {
	Name string
	Age  int
	Tags []string
}

type PersonCodec struct{}

func (PersonCodec) Marshal(v Person, bs []byte) (n int) {
	n = composite.String.Marshal(v.Name, bs)
	n += varnum.Int.Marshal(v.Age, bs[n:])
	return n + composite.NewSliceSerializer(composite.String).Marshal(v.Tags, bs[n:])
}

func (PersonCodec) Unmarshal(bs []byte) (v Person, n int, err error) {
	v.Name, n, err = composite.String.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Age, n1, err = varnum.Int.Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Tags, n1, err = composite.NewSliceSerializer(composite.String).Unmarshal(bs[n:])
	n += n1
	return
}

func (PersonCodec) Size(v Person) (size int) {
	return composite.String.Size(v.Name) +
		varnum.Int.Size(v.Age) +
		composite.NewSliceSerializer(composite.String).Size(v.Tags)
}

func (PersonCodec) Skip(bs []byte) (n int, err error) {
	n, err = composite.String.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = varnum.Int.Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = composite.NewSliceSerializer(composite.String).Skip(bs[n:])
	return n + n1, err
}

var PersonSerializer = PersonCodec{}

// Usage:
p := Person{Name: "ada", Age: 36, Tags: []string{"math", "code"}}
bs := binco.Encode(PersonSerializer, p)      // 13 bytes
back, err := binco.Decode(PersonSerializer, bs)
```

The pattern is always the same: encode field by field into `bs[n:]`, decode
field by field from `bs[n:]`, and accumulate `n`.

### 4.4 The same struct from the generator

Writing that by hand does not scale. The generator writes it for you:

```go
//go:generate go run ./gen

type Person struct {
	Name string
	Age  int
	Tags []string
}
```

```go
// gen/main.go
package main

import (
	"log"
	"os"
	"reflect"

	"example.com/app"

	"github.com/malivvan/binco/codegen"
	genopts "github.com/malivvan/binco/codegen/config/gen"
)

func main() {
	g, err := codegen.NewGenerator(genopts.WithPath("example.com/app"))
	if err != nil {
		log.Fatal(err)
	}
	if err = g.AddStruct(reflect.TypeFor[app.Person]()); err != nil {
		log.Fatal(err)
	}
	bs, err := g.Generate()
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile("person_binco.gen.go", bs, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

`go generate` now produces `PersonBinco`, a `binco.Serializer[Person]` with the
same behaviour as the hand-written codec. The full story is in
[Code generation](#9-code-generation).

### 4.5 Over a stream

```go
import (
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
)

// Writing.
if _, err := wirefmt.Encode(wcomposite.String, conn, "hello"); err != nil {
	return err
}

// Reading.
msg, n, err := wirefmt.Decode(wcomposite.String, conn)
```

`wirefmt.Encode` and `wirefmt.Decode` accept any `io.Writer`/`io.Reader`.
`Encode` buffers through a pooled `bufio.Writer`; `Decode` reads exactly the
bytes it needs and never reads ahead, so it cannot swallow the next message of a
shared stream. For a long-lived stream use `wirefmt.NewEncoder` and
`wirefmt.NewDecoder`, which own one buffer each, or pass your own
`*bufio.Writer`/`*bufio.Reader`.

## 5. The wire format

### 5.1 Anatomy of a payload

A payload is a sequence of encoded values with no separators:

```
[ field 1 ][ field 2 ][ field 3 ] ...
```

There is no header, no version, no length, no checksum and no padding. The
receiver reads values in the order the schema declares them. Every value is
self-delimiting: a number by its width or its LEB128 continuation bits, a
container by its length prefix, a pointer by its flag byte.

### 5.2 Booleans and bytes

`bool` and `byte` occupy exactly one byte.

```
true     ->  [1]
false    ->  [0]
byte 7   ->  [7]
```

A `bool` byte other than 0 or 1 is rejected with `binco.ErrWrongFormat`.

### 5.3 Varint

`package varnum` encodes unsigned integers in LEB128 form: seven bits per byte,
least significant group first, the high bit set on every byte but the last.

```
40          ->  [40]
127         ->  [127]
128         ->  [128 1]
300         ->  [172 2]
2^64-1      ->  [255 255 255 255 255 255 255 255 255 1]
```

Signed integers are ZigZag-mapped first, so that small negative numbers stay
small:

```
 0          ->  [0]
-1          ->  [1]
 1          ->  [2]
-2          ->  [3]
 63         ->  [126]
-64         ->  [127]
 64         ->  [128 1]
 300        ->  [216 4]
-300        ->  [215 4]
```

`varnum.PositiveInt` and its sized variants skip the ZigZag step. They are meant
for values that cannot be negative: lengths, type identifiers, pointer ids.

An encoding must be *canonical*: the shortest form of the value. A multi-byte
encoding whose final byte is zero (`[128 128 128 0]`) is rejected with
`binco.ErrNonCanonical`, so every value has exactly one byte representation and
hashing, signing or de-duplicating payloads by their bytes is meaningful.

An encoding may not exceed the width of its type. When it does, decoding fails
with `binco.ErrOverflow`:

| Type | Longest encoding | Largest final byte |
|------|------------------|--------------------|
| `uint8` | 2 bytes | 1 |
| `uint16` | 3 bytes | 3 |
| `uint32` | 5 bytes | 15 |
| `uint64` | 10 bytes | 1 |

Floats are stored as the unsigned varint of their IEEE 754 bit pattern (64 or 32
bits), which keeps `0.0` and small denormals short and `1.5` at nine bytes.

### 5.4 Fixed width

`package scalar` writes numbers in little-endian byte order using the width of
their Go type: eight bytes for 64-bit values, four for 32-bit, two for 16-bit,
one for 8-bit. `int` and `uint` are always encoded as eight bytes, whatever the
host word size is, so a payload written on amd64 reads back on a 32-bit build; a
value the host cannot hold is reported as `binco.ErrOverflow`.

```
uint64 40     ->  [40 0 0 0 0 0 0 0]
uint64 max    ->  [255 255 255 255 255 255 255 255]
uint16 0x1234 ->  [52 18]
int64 -1      ->  [255 255 255 255 255 255 255 255]
float64 1.0   ->  [0 0 0 0 0 0 240 63]
float32 1.5   ->  [0 0 192 63]
byte 7        ->  [7]
```

Raw beats varint in size only for large values:

| Type | Raw becomes smaller above |
|------|---------------------------|
| `uint64` | 2^56 |
| `uint32` | 2^28 |
| `uint16` | 2^14 |
| `uint8` | never |

Its advantage is speed: no shifts, no branches, and in `package zerocopy` a single
machine-word store.

### 5.5 Strings and byte slices

Both are a length followed by raw bytes. The length is a positive integer
(`varnum.PositiveInt` by default).

```
"hello world"   ->  [11 104 101 108 108 111 32 119 111 114 108 100]
[]byte{1,2,3}   ->  [3 1 2 3]
""              ->  [0]
nil []byte      ->  [0]
```

A nil string/byte slice and an empty one produce the same bytes and both decode
to an empty, non-nil value. Byte slices carry no UTF-8 meaning, so invalid UTF-8
survives a round trip inside `[]byte` but not necessarily inside `string`
(Go does not validate on conversion, so in practice it survives there too).

### 5.6 Slices and arrays

A slice is a length followed by its elements. An array is encoded the same way,
but its type fixes the length.

```
[]int{1,2,3}        ->  [3 2 4 6]
[3]int{1,2,3}       ->  [3 2 4 6]
[]string{"hello", "world"}
                    ->  [2 5 104 101 108 108 111 5 119 111 114 108 100]
```

Decoding into an array requires the length prefix to match the array type: a
payload of a different length fails with `binco.ErrWrongLength`. That keeps the
decode memory-safe.

### 5.7 Maps

A map is a length followed by key/value pairs:

```
map[int]string{1: "hello"}  ->  [1 2 5 104 101 108 108 111]
```

Entries appear in Go's map iteration order, which is random. Two marshals of the
same map may produce different bytes, though both decode to equal maps. If you
need canonical bytes, encode a sorted slice of pairs instead; see
[Collections](#85-collections).

### 5.8 Pointers

A pointer starts with a one-byte flag:

| Flag | Name | Meaning |
|------|------|---------|
| `0` | `binco.PtrNotNil` | the pointed-to value follows |
| `1` | `binco.PtrNil` | nil, nothing follows |
| `2` | `binco.PtrMapping` | an identifier follows; data only the first time |

The plain pointer codec (`composite.NewPtrSerializer`) writes flags 0 and 1 and rejects flag 2:

```
nil *string        ->  [1]
&"hello world"     ->  [0 11 104 101 ... 100]
```

Pointer mapping (`package ptrmap`) uses flag 2 to write a pointer graph once:

```go
ptr1, ptr2 := &10, &20
list := []*int{ptr1, ptr1, ptr2}

list               ->  [3 2 0 20 2 0 2 1 40]
```

```
[3]      number of elements
[2 0]    mapping flag, id 0, then the value 10 (ZigZag 20)
[2 0]    the same pointer again: flag and id only
[2 1]    mapping flag, id 1, then the value 20 (ZigZag 40)
```

Ids start at zero in encounter order and are shared by every `ptrmap` codec of one
value. An id remembers the type it was decoded as, and a reference is only
resolved when that type matches: a payload cannot make a `*string` out of an
`*int`, and the mismatch is reported as `binco.ErrWrongType`.

Both pointer codecs bound their recursion. `ptrmap` counts depth per decoded
value and stops at `binco.MaxDepth` (1000); the plain pointer codec shares one
atomic counter per serializer and stops at `binco.MaxPointerDepth` (10000). A
payload of nested pointers therefore returns `binco.ErrTooDeep` instead of
exhausting the goroutine stack. See
[Shared pointers and identity](#87-shared-pointers-and-identity).

### 5.9 Time

`time.Time` is stored as a Unix timestamp, using a fixed eight bytes:

```
time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
  with scalar.TimeUnixUTC  ->  [217 129 56 102 0 0 0 0]
```

Four units are available: seconds, milliseconds, microseconds and nanoseconds.
Every variant decodes to UTC, so the same payload compares equal on machines in
different zones; the `...UTC` names are kept for compatibility and mean the same
thing. Precision the unit cannot represent is lost, and `UnixNano` overflows
silently outside roughly 1678-2262 (`UnixMilli` has a far wider range).

### 5.10 Type identifiers

A DTM (data type metadata) is a positive integer written in front of a payload
to name its type:

```
DTM 7 + int 300    ->  [7 216 4]
```

It is written with `varnum.PositiveInt`, so ids below 128 cost one byte. The
receiver reads the id first and dispatches to the matching codec:

```go
dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
switch dtm {
case LoginDTM:
	// decode the rest with the Login codec
case ChatDTM:
	// decode the rest with the Chat codec
default:
	err = binco.NewUnexpectedDTMError(dtm)
}
```

DTMs are the only "framing" the format has, which makes them the tool of choice
for streams and files of heterogeneous values.

Ids are positive: a zero or negative id would be written as a huge unsigned
number, so `typeid.DTMSerializer` rejects it with `binco.ErrOverflow`. In the
generator, pin the id of every registered type with `genopts.WithDTM`; without it
the generator numbers the types in registration order, and adding or removing one
renumbers the others. Generated constants carry literal values, duplicate ids are
rejected, and `Generator.DTMRegistry` returns the mapping for a golden file.

### 5.11 A payload dissected

Consider a record encoded as its three fields, using a varint for the number, an
8-byte raw integer for the source port and a string for the host:

```go
bs := []byte{
	216, 4, // 300 as a ZigZag varint
	0xD4, 0x3D, 0, 0, 0, 0, 0, 0, // 15828 as an 8-byte raw integer
	9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', // length + bytes
}
```

Reading it back is a matter of knowing the schema:

```go
id, n1, err := varnum.Int.Unmarshal(bs)
port, n2, err := scalar.Uint64.Unmarshal(bs[n1:])
host, n3, err := composite.String.Unmarshal(bs[n1+n2:])
// id = 300, port = 15828, host = "localhost", n1+n2+n3 == len(bs)
```

Nothing in the bytes says which is which; the decode order does.

### 5.12 Compatibility rules

A payload is compatible with a reader when the reader decodes the same field
sequence with the same codecs. In practice:

**Safe changes**

* Adding a trailing field, if old readers may stop early and new readers get a
  payload that contains it. Old payloads lack the bytes, so new readers must
  know the format version some other way (a DTM, an envelope, or a version field
  at the front).
* Changing a codec's *implementation* (`composite.String` to `zerocopy.String`, a
  generator mode switch) - the bytes are identical.
* Adding a new type with a new DTM.

**Breaking changes**

* Reordering or removing fields.
* Changing a field's codec or unit (`varnum` to `scalar`, seconds to milliseconds).
* Changing a length codec.
* Reusing a DTM for different data.

When a break is unavoidable, introduce a new version and keep the reader able to
decode the old one; see [Versioning and migration](#810-versioning-and-migration).

The generated `SchemaFingerprint` constant is the cheap check: store it beside a
payload (or compare it at deployment time) and a reordered field is caught before
the data is misread.

## 6. Choosing codecs

Three families encode numbers, and they are interchangeable byte-wise only
within a family:

| Package | Encoding | Best for |
|---------|----------|----------|
| `varnum` | LEB128, ZigZag for signed | small and medium numbers, ids, counters |
| `scalar` | fixed width, little-endian | large numbers, random-access data, speed |
| `zerocopy` | fixed width through memory access | hot paths, zero-copy blobs |

`zerocopy` is byte-compatible with `scalar` for numbers and with `composite` for
strings and byte slices, so switching codecs never changes the format, only the
code.

**Points to weigh**

* Space: varint wins below the thresholds in [Fixed width](#54-fixed-width).
* Speed: `scalar` and `zerocopy` win, because they avoid branches and bit assembly.
* Variance: `scalar` has constant size, which simplifies fixed-size records and
  random access into a stream of records.

**Length codecs** default to `varnum.PositiveInt`. Replacing them with a
fixed-width codec (`scalar.Int`) makes every length cost eight bytes instead of one
to five, but decoding is a single load. Use it only when you need that.

**Time units** are chosen by the codec: `scalar.TimeUnix` (seconds) for coarse
timestamps, `scalar.TimeUnixNano` for precision. The `UTC` variants cost nothing
extra and remove the local-time surprise.

**Generator modes** apply the same choice per type:

| Go type | safe (default) | `WithUnsafe` | `WithNotUnsafe` |
|---------|----------------|--------------|-----------------|
| `int`, `int64`, `int32`, `int16`, `uint`, ... | `varnum` | `zerocopy` | `zerocopy` |
| `int8`, `uint8`, `byte`, `float32`, `float64` | `scalar` | `zerocopy` | `zerocopy` |
| `bool`, `string`, `[]byte` | `composite` | `zerocopy` | `zerocopy` (`composite` for strings) |
| `time.Time` | `scalar.TimeUnixUTC` | `zerocopy.TimeUnixUTC` | `zerocopy.TimeUnixUTC` |
## 7. Package guide

### 7.1 binco

The root package holds everything the packages share.

**The codec contract**

```go
type Serializer[T any] interface {
	Marshal(t T, bs []byte) (n int)
	Unmarshal(bs []byte) (t T, n int, err error)
	Size(t T) (size int)
	Skip(bs []byte) (n int, err error)
}
```

**Self-marshalling types**

```go
type Marshaller interface {
	MarshalBinco(bs []byte) (n int)
	SizeBinco() (size int)
}

type MarshallerTyped interface {
	MarshalTypedBinco(bs []byte) (n int)
	SizeTypedBinco() (size int)
}
```

Types implementing them can be encoded without naming a codec:

```go
bs := binco.Marshal(v)        // allocates SizeBinco bytes, then MarshalBinco
bs = binco.MarshalTyped(v)    // the same, with the type identifier in front
```

The method names carry the package name so a type can also satisfy
`encoding/json`, `encoding/binary` and friends without collisions.

**Helpers**

```go
bs := binco.Encode(varnum.Int, 42)       // allocate + Marshal
v, err := binco.Decode(varnum.Int, bs)   // Unmarshal without the byte count
```

**Type identifiers**

```go
type DTM int
```

**Validators**

```go
type Validator[T any] interface{ Validate(t T) error }
type ValidatorFn[T any] func(t T) error

limited := binco.ValidatorFn[int](func(n int) error {
	if n > 100 {
		return errors.New("too many")
	}
	return nil
})
```

**Pointer maps** (used by `ptrmap`)

```go
ptrMap := binco.NewPtrMap()           // Put(unsafe.Pointer, reflect.Type) int,
                                      // Get(unsafe.Pointer, reflect.Type), Len, Clear
revPtrMap := binco.NewReversePtrMap() // Put(int, unsafe.Pointer, reflect.Type),
                                      // Get(int), Len, Clear
```

**Limits and widths**

```go
binco.Num8RawSize, binco.Num16RawSize, binco.Num32RawSize, binco.Num64RawSize
binco.Uint8MaxVarintLen ... binco.Uint64MaxVarintLen
binco.Uint8MaxLastByte  ... binco.Uint64MaxLastByte
```

**Type constraints** for writing generic codecs: `Signed`, `Unsigned`,
`Integer8`, `Integer16`, `Integer32`, `Integer64`, `Num32`, `Num64`.

**Errors**: see the [error reference](#12-error-reference), plus
`UnexpectedDTMError` (`DTM()`) and `WrongDTMError` (`Expected()`, `Actual()`).

### 7.2 varnum

Variable-length numbers. The full list:

```go
varnum.Byte

varnum.Uint8, varnum.Uint16, varnum.Uint32, varnum.Uint64, varnum.Uint
varnum.Int8,  varnum.Int16,  varnum.Int32,  varnum.Int64,  varnum.Int

varnum.PositiveInt8, varnum.PositiveInt16, varnum.PositiveInt32,
varnum.PositiveInt64, varnum.PositiveInt      // no ZigZag, non-negative

varnum.Float32, varnum.Float64

varnum.EncodeZigZag[T Signed](t T) T
varnum.DecodeZigZag[T Unsigned](t T) T
```

* Use `Uint*`/`Int*` for general numbers.
* Use `PositiveInt*` for lengths, ids and counts.
* `Uint`/`Int` use a 64-bit wire width on every host; a value the host cannot
  hold is `binco.ErrOverflow`.

### 7.3 scalar

Fixed-width numbers and timestamps, little-endian.

```go
scalar.Byte
scalar.Uint8, scalar.Uint16, scalar.Uint32, scalar.Uint64, scalar.Uint
scalar.Int8,  scalar.Int16,  scalar.Int32,  scalar.Int64,  scalar.Int
scalar.Float32, scalar.Float64

scalar.TimeUnix, scalar.TimeUnixMilli, scalar.TimeUnixMicro, scalar.TimeUnixNano
scalar.TimeUnixUTC, scalar.TimeUnixMilliUTC, scalar.TimeUnixMicroUTC, scalar.TimeUnixNanoUTC

scalar.SkipInteger8, scalar.SkipInteger16, scalar.SkipInteger32, scalar.SkipInteger64
```

### 7.4 composite

Bools, strings, byte slices, pointers and the container codecs.

```go
composite.Bool
composite.String
composite.ByteSlice

composite.NewStringSerializer(opts ...stropts.SetOption)         // string
composite.NewByteSliceSerializer(opts ...bslopts.SetOption)      // []byte
composite.NewPtrSerializer[T any](baseSerializer binco.Serializer[T]) // *T (flags 0 and 1)

composite.NewSliceSerializer[T any](elemSerializer binco.Serializer[T], opts ...slopts.SetOption[T])
composite.NewMapSerializer[T comparable, V any](keySerializer binco.Serializer[T],
	valueSerializer binco.Serializer[V], opts ...mapopts.SetOption[T, V])
```

Every constructor honours the validators it is given, so there is one way to
build a codec: pass the validator as an option. The constructors return
`binco.Serializer[T]`, which is all a caller needs.

Building blocks for hand-written codecs:

```go
composite.MarshalString, composite.SizeString, composite.SkipString
composite.MarshalByteSlice, composite.SizeByteSlice, composite.SkipByteSlice
composite.MarshalSlice, composite.UnmarshalSlice, composite.SizeSlice, composite.SkipSlice
composite.UnmarshalValidSlice
composite.SkipBool
```

### 7.5 zerocopy

Same wire format as `scalar`/`composite`, faster, and zero-copy for strings and byte
slices. Every scalar is available, plus arrays:

```go
zerocopy.Bool, zerocopy.String, zerocopy.ByteSlice
zerocopy.Byte, zerocopy.Uint8 ... zerocopy.Uint64, zerocopy.Uint
zerocopy.Int8 ... zerocopy.Int64, zerocopy.Int
zerocopy.Float32, zerocopy.Float64
zerocopy.TimeUnix ... zerocopy.TimeUnixNanoUTC

zerocopy.NewStringSerializer(opts ...stropts.SetOption)
zerocopy.NewByteSliceSerializer(opts ...bslopts.SetOption)
zerocopy.NewArraySerializer[T, V any](elemSerializer binco.Serializer[V], opts ...arropts.SetOption[V])
```

Rules:

* A decoded `string` or `[]byte` **aliases the input buffer**. It is valid while
  the buffer is, and it must not be modified. Reusing the buffer (for example
  from a `sync.Pool`) silently mutates values that were already decoded, and any
  small string keeps the whole buffer alive.
* Scalars are converted through `encoding/binary` with an explicit little-endian
  order, so the codecs are correct on any host byte order and at an unaligned
  offset. Only the array codec reinterprets memory, and it checks with reflection
  that its two type parameters really describe an array of the element type
  (panicking with `binco.ErrWrongArrayType` if they do not).
* Arrays live only here: decoding into an array means reinterpreting the decoded
  elements as the array type. The length is checked before any element is read,
  and decoding fills the array in place.

### 7.6 ptrmap

Pointer mapping: pointer identity, shared references and cycles.

```go
ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()

ptrCodec := ptrmap.NewPtrSerializer(ptrMap, revPtrMap, baseCodec) // *Base
wrapped := ptrmap.Wrap(ptrMap, revPtrMap, anyCodec)       // resets the maps per call
```

Wiring rules:

1. Every `ptrmap` codec of one value shares the same pair of maps.
2. The outermost codec is wrapped with `ptrmap.Wrap`, which clears both maps
   around every operation (including `Size`, which assigns ids). Calling `Size` or
   `Marshal` on an inner codec directly leaves ids behind and breaks the
   `Size`/`Marshal` invariant.
3. A wrapped codec is not safe for concurrent use; give each goroutine its own
   maps (see [Concurrency](#818-concurrency)).

An identifier is stored together with the type it was decoded as, and the encoder
keys its table by `(address, type)`. Two pointers that share an address (a struct
and its first field, or any two zero-size values) therefore get different ids, and
a reference read as a different type fails with `binco.ErrWrongType` instead of
being cast with `unsafe.Pointer`. Recursive references stop at
`binco.MaxDepth` with `binco.ErrTooDeep`.

### 7.7 typeid

Type identifiers in front of payloads.

```go
typeid.DTMSerializer                                 // binco.Serializer[binco.DTM]
typeid.NewSerializer[T any](dtm binco.DTM, serializer binco.Serializer[T]) typeid.Serializer[T]

serializer.DTM() binco.DTM
serializer.UnmarshalData(bs []byte) (T, int, error)  // data only
serializer.SkipData(bs []byte) (int, error)
```

`typeid.Serializer[T]` itself implements `binco.Serializer[T]`, so it drops into any position
where a codec is expected, including inside slices and maps.

### 7.8 config

Each container codec has a config package. Options are function values
applied to an `Options` struct; `Apply` is exported so generated code and your
own helpers can build option slices.

| Package | Import alias | Options |
|---------|--------------|---------|
| `binco/config/array` | `arropts` | `WithLenSerializer`, `WithElemValidator` |
| `binco/config/byte_slice` | `bslopts` | `WithLenSerializer`, `WithLenValidator` |
| `binco/config/map` | `mapopts` | `WithLenSerializer`, `WithLenValidator`, `WithKeyValidator`, `WithValueValidator` |
| `binco/config/slice` | `slopts` | `WithLenSerializer`, `WithLenValidator`, `WithElemValidator` |
| `binco/config/string` | `stropts` | `WithLenSerializer`, `WithLenValidator` |

```go
bounded := composite.NewStringSerializer(
	stropts.WithLenValidator(binco.ValidatorFn[int](func(n int) error {
		if n > 64 {
			return errors.New("string too long")
		}
		return nil
	})),
)

fixedLengths := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](scalar.Int))
```

The streaming tree has the same options under `binco/wirefmt/config/*`. In both
trees the plain constructors honour the validators, so `WithLenValidator` is never
silently ignored.

### 7.9 wirefmt

The streaming API mirrors the buffer API:

```go
type Writer interface { io.ByteWriter; io.Writer; io.StringWriter }
type Reader interface { io.ByteReader; io.Reader }

type Serializer[T any] interface {
	Marshal(t T, w Writer) (n int, err error)
	Unmarshal(r Reader) (t T, n int, err error)
	Size(t T) (size int)
	Skip(r Reader) (n int, err error)
}
```

`*bufio.Writer`, `*bytes.Buffer` and `*strings.Builder` satisfy `Writer`;
`*bufio.Reader` and `*bytes.Reader` satisfy `Reader`. For any other stream:

```go
w := wirefmt.NewWriter(conn)   // writes one byte at a time through Write
r := wirefmt.NewReader(conn)   // never reads past the bytes it was asked for
```

The adapters are correct but slow; prefer `bufio` at the call site. The helpers
do the buffering for you:

```go
n, err := wirefmt.Encode(wcomposite.String, w, "hello")
v, n, err := wirefmt.Decode(wcomposite.String, r)

// Reuse one buffer for a whole stream.
enc := wirefmt.NewEncoder(w)
dec := wirefmt.NewDecoder(r)
n, err = wirefmt.EncodeTo(enc, wcomposite.String, "hello")
v, n, err = wirefmt.DecodeFrom(dec, wcomposite.String)
```

`Decode` never buffers a reader it cannot read one byte from, because it cannot
know whether the caller still needs the bytes it would read ahead; buffering per
call used to swallow the next message of a shared stream. `NewDecoder` owns one
`bufio.Reader` for the lifetime of the stream, buffers once and is the fast
option. `Encode` reuses a pooled `bufio.Writer`.

Truncation is `io.ErrUnexpectedEOF`; only a stream that ends before the first
byte of a value returns `io.EOF`. `wirefmt.SliceWriter` adapts a `[]byte` to a
`Writer`, which is useful for testing and for embedding a streaming codec.

Marshalling and helper packages:

```go
type Marshaller interface {
	MarshalBinco(w Writer) (n int, err error)
	SizeBinco() (size int)
}

type MarshallerTyped interface {
	MarshalTypedBinco(w Writer) (n int, err error)
	SizeTypedBinco() (size int)
}
```

Subpackages: `wirefmt/varnum`, `wirefmt/scalar`, `wirefmt/composite`, `wirefmt/zerocopy`,
`wirefmt/ptrmap`, `wirefmt/typeid`, `wirefmt/config/*`. `wirefmt/zerocopy` has no byte slice
codec: a wire cannot hand out a view of bytes it has not read yet, so byte slices
go through `wirefmt/composite`.

## 8. Use cases

Each recipe shows a working shape: the helper types and codecs it uses are
declared in the same recipe or in a neighbouring one, and the snippets compile
as written once those declarations exist.

### 8.1 Numbers and enums

**Plain numbers.** Pick a family once and use it everywhere for that field:

```go
type Metrics struct {
	Requests uint64  // counters: fixed is constant-size and fast
	Latency  int64   // milliseconds: varint keeps small values short
	Load     float32 // ratio: fixed
}

```

**Defined numeric types.** A defined type is not assignable to its underlying
type, so wrap the codec:

```go
type UserID uint64

type userIDCodec struct{}

func (userIDCodec) Marshal(v UserID, bs []byte) (n int) {
	return varnum.Uint64.Marshal(uint64(v), bs)
}
func (userIDCodec) Unmarshal(bs []byte) (v UserID, n int, err error) {
	raw, n, err := varnum.Uint64.Unmarshal(bs)
	return UserID(raw), n, err
}
func (userIDCodec) Size(v UserID) (size int) { return varnum.Uint64.Size(uint64(v)) }
func (userIDCodec) Skip(bs []byte) (n int, err error) { return varnum.Uint64.Skip(bs) }
```

**Enums.** Encode the smallest integer that covers the values and validate the
range while decoding:

```go
type Level uint8

var ErrUnknownLevel = errors.New("unknown level")

type levelCodec struct{}

func (levelCodec) Marshal(v Level, bs []byte) (n int) {
	return scalar.Byte.Marshal(byte(v), bs)
}

func (levelCodec) Unmarshal(bs []byte) (v Level, n int, err error) {
	b, n, err := scalar.Byte.Unmarshal(bs)
	if err != nil {
		return 0, n, err
	}
	if b > 3 {
		return 0, n, fmt.Errorf("%w: %d", ErrUnknownLevel, b)
	}
	return Level(b), n, nil
}

func (levelCodec) Size(v Level) (size int) { return 1 }
func (levelCodec) Skip(bs []byte) (n int, err error) { return scalar.Byte.Skip(bs) }
```

Unknown enum values are a fact of life in evolving systems; rejecting them is a
policy decision. To keep reading, map unknown values to a sentinel instead.

**Durations.** Store the unit you need, in the smallest type that holds it:

```go
type Millis time.Duration

type millisCodec struct{}

func (millisCodec) Marshal(v Millis, bs []byte) (n int) {
	return varnum.Int64.Marshal(int64(v), bs)
}
func (millisCodec) Unmarshal(bs []byte) (v Millis, n int, err error) {
	m, n, err := varnum.Int64.Unmarshal(bs)
	return Millis(m), n, err
}
func (millisCodec) Size(v Millis) (size int) { return varnum.Int64.Size(int64(v)) }
func (millisCodec) Skip(bs []byte) (n int, err error) { return varnum.Int64.Skip(bs) }
```

### 8.2 Strings and byte blobs

```go
name := "hello world"
bs := binco.Encode(composite.String, name)     // [11 104 101 ... 100]

blob := []byte{1, 2, 3}
bs = binco.Encode(composite.ByteSlice, blob)   // [3 1 2 3]
```

Strings are UTF-8 by convention, byte slices are fixed. Both are length-prefixed,
so `nil` and `""`/`[]byte{}` encode identically and decode to an empty value.

**Bounding a length.** Never trust a length from the network: allocate only
after a validator accepted it.

```go
boundedText := composite.NewStringSerializer(
	stropts.WithLenValidator(binco.ValidatorFn[int](func(n int) error {
		if n > 1<<16 {
			return errors.New("string too long")
		}
		return nil
	})),
)

boundedBlob := composite.NewByteSliceSerializer(
	bslopts.WithLenValidator(binco.ValidatorFn[int](func(n int) error {
		if n > 1<<20 {
			return errors.New("blob too large")
		}
		return nil
	})),
)
```

**Zero-copy variants.** If the buffer outlives the value, `zerocopy` avoids the
copy:

```go
s, n, err := zerocopy.String.Unmarshal(bs)   // s aliases bs
```

Read [Zero-copy hot paths](#814-zero-copy-hot-paths) before using it.

### 8.3 Structs by hand

The full pattern, with a nested container:

```go
type Order struct {
	ID     UserID
	Items  []Item
	Note   string
	Total  int64
}

type orderCodec struct{}

func (orderCodec) Marshal(v Order, bs []byte) (n int) {
	n = userIDCodec{}.Marshal(v.ID, bs)
	n += composite.NewSliceSerializer[Item](ItemCodec{}).Marshal(v.Items, bs[n:])
	n += composite.String.Marshal(v.Note, bs[n:])
	return n + varnum.Int64.Marshal(v.Total, bs[n:])
}

func (orderCodec) Unmarshal(bs []byte) (v Order, n int, err error) {
	v.ID, n, err = userIDCodec{}.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Items, n1, err = composite.NewSliceSerializer[Item](ItemCodec{}).Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Note, n1, err = composite.String.Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Total, n1, err = varnum.Int64.Unmarshal(bs[n:])
	n += n1
	return
}

func (orderCodec) Size(v Order) (size int) {
	return userIDCodec{}.Size(v.ID) +
		composite.NewSliceSerializer[Item](ItemCodec{}).Size(v.Items) +
		composite.String.Size(v.Note) +
		varnum.Int64.Size(v.Total)
}

func (orderCodec) Skip(bs []byte) (n int, err error) {
	n, err = userIDCodec{}.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = composite.NewSliceSerializer[Item](ItemCodec{}).Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = composite.String.Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = varnum.Int64.Skip(bs[n:])
	return n + n1, err
}
```

**Required errors when no data is consumed.** Returning `err` from
`Unmarshal` stops the caller immediately; `Skip` must propagate errors the same
way. A codec that ignores an error produces confusing failures far from the
cause.

**Trailing fields.** If you append a field to the schema, old payloads do not
contain it. Either version the type, or make the reader tolerate a truncated
tail by checking `n` against `len(bs)`:

```go
func (c orderCodec) Unmarshal(bs []byte) (v Order, n int, err error) {
	// ... decode the original fields ...
	if n == len(bs) {
		return v, n, nil // an old payload without the new field
	}
	v.Total, n1, err = varnum.Int64.Unmarshal(bs[n:])
	n += n1
	return
}
```

This works because every value is self-delimiting and `Unmarshal` never reads
past the buffer.

### 8.4 Structs from the generator

The generator produces exactly the code above. For a type

```go
type Order struct {
	ID    UserID
	Items []Item
	Note  string
	Total int64
}
```

call `g.AddStruct(reflect.TypeFor[Order]())` and you get `OrderBinco` with the
four methods. Nested structs, slices, maps and pointers are resolved
automatically. Read [Code generation](#9-code-generation) for the options and
the workflow.

### 8.5 Collections

**Slices**

```go
ints := composite.NewSliceSerializer(varnum.Int)
names := composite.NewSliceSerializer(composite.String)
matrix := composite.NewSliceSerializer(composite.NewSliceSerializer(varnum.Int))

bs := binco.Encode(names, []string{"hello", "world"})
// [2 5 104 101 108 108 111 5 119 111 114 108 100]
```

**Arrays** need the `zerocopy` package (the length must match the array type):

```go
arrSerializer := zerocopy.NewArraySerializer[[3]int, int](varnum.Int)
bs := binco.Encode(arrSerializer, [3]int{1, 2, 3})   // [3 2 4 6]

// A payload with a different length is rejected:
_, _, err := arrSerializer.Unmarshal(binco.Encode(composite.NewSliceSerializer(varnum.Int), []int{1, 2}))
// err: binco: wrong length
```

**Maps**

```go
byName := composite.NewMapSerializer(composite.String, varnum.Int)
bs := binco.Encode(byName, map[string]int{"a": 1}) // [1 1 97 2]
```

Map iteration order is random, so the bytes of a multi-entry map vary between
runs. Decoding is deterministic. If a payload must be canonical (content
addressing, signatures, golden files), sort the entries and encode a slice:

```go
type Entry struct {
	Key   string
	Value int
}

type entryCodec struct{}

func (entryCodec) Marshal(v Entry, bs []byte) (n int) {
	n = composite.String.Marshal(v.Key, bs)
	return n + varnum.Int.Marshal(v.Value, bs[n:])
}

func (entryCodec) Unmarshal(bs []byte) (v Entry, n int, err error) {
	v.Key, n, err = composite.String.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Value, n1, err = varnum.Int.Unmarshal(bs[n:])
	n += n1
	return
}

func (entryCodec) Size(v Entry) (size int) {
	return composite.String.Size(v.Key) + varnum.Int.Size(v.Value)
}

func (entryCodec) Skip(bs []byte) (n int, err error) {
	n, err = composite.String.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = varnum.Int.Skip(bs[n:])
	return n + n1, err
}

func CanonicalMap(m map[string]int) []byte {
	entries := make([]Entry, 0, len(m))
	for k, v := range m {
		entries = append(entries, Entry{Key: k, Value: v})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	return binco.Encode(composite.NewSliceSerializer[Entry](entryCodec{}), entries)
}
```

**Sets.** A set is a map with an empty value, or a sorted slice of keys:

```go
setSerializer := composite.NewSliceSerializer(composite.String) // sorted on the writer side
```

**Length codecs.** If you need random access over a stream of slices, a
fixed-width length lets a reader seek:

```go
seekable := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](scalar.Int))
```

### 8.6 Optional fields and nil

There is no "optional" marker in the format; presence is decided by the schema.
Three common patterns:

**Pointer (cheap, explicit)**

```go
type Profile struct {
	Nickname *string   // nil -> flag byte 1
	Bio      string    // always present
}

var profileSerializer = composite.NewPtrSerializer(composite.String) // use for Nickname
```

A nil pointer costs one byte, a non-nil pointer costs one byte plus the value.

**Sentinel value.** When the domain has a spare value (empty string, zero id),
use it and document it. No extra byte on the wire.

**Presence flag plus value.** When you need "set to empty" distinct from
"unset", encode a bool:

```go
type Maybe struct {
	Set   bool
	Value string
}
```

Decode with a validator to keep the pair consistent.

### 8.7 Shared pointers and identity

Two fields pointing at the same string must still point at the same string after
decoding. `ptrmap` writes the value once and refers to it by id:

```go
type Shared struct {
	A *string
	B *string
}

func SharedSerializer() binco.Serializer[Shared] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	strPtr := ptrmap.NewPtrSerializer(ptrMap, revPtrMap, composite.String)
	return ptrmap.Wrap(ptrMap, revPtrMap, sharedCodec{strPtr})
}

type sharedCodec struct {
	strPtr binco.Serializer[*string]
}

func (s sharedCodec) Marshal(v Shared, bs []byte) (n int) {
	n = s.strPtr.Marshal(v.A, bs)
	return n + s.strPtr.Marshal(v.B, bs[n:])
}

func (s sharedCodec) Unmarshal(bs []byte) (v Shared, n int, err error) {
	v.A, n, err = s.strPtr.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.B, n1, err = s.strPtr.Unmarshal(bs[n:])
	n += n1
	return
}

func (s sharedCodec) Size(v Shared) (size int) {
	return s.strPtr.Size(v.A) + s.strPtr.Size(v.B)
}

func (s sharedCodec) Skip(bs []byte) (n int, err error) {
	n, err = s.strPtr.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.strPtr.Skip(bs[n:])
	return n + n1, err
}
```

```go
s := "hello"
v := Shared{A: &s, B: &s}
bs := binco.Encode(SharedSerializer(), v)
back, _ := binco.Decode(SharedSerializer(), bs)
fmt.Println(back.A == back.B) // true
```

The encoded form is `[2 0 5 hello 2 0]`: flag 2, id 0, the value, then flag 2
and the same id with no data. See [Pointers](#58-pointers).

**When not to use ptrmap.** If your data is a tree with no sharing, `composite.NewPtrSerializer`
is smaller and faster: no ids, no maps, no bookkeeping.

### 8.8 Cycles linked lists trees graphs

A cycle is a pointer graph where following pointers returns to a node. Plain
pointer codecs cannot represent that; `ptrmap` can, because the second visit is
written as an id.

**Doubly linked list.** Each element has `Prev` and `Next`; the list has `Head`,
`Tail` and `Len`:

```go
type Elem struct {
	Value int
	Prev  *Elem
	Next  *Elem
}

type List struct {
	Head *Elem
	Tail *Elem
	Len  int
}

type elemCodec struct {
	next binco.Serializer[*Elem]
}

func (s *elemCodec) Marshal(v Elem, bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Value, bs)
	n += s.next.Marshal(v.Prev, bs[n:])
	return n + s.next.Marshal(v.Next, bs[n:])
}

func (s *elemCodec) Unmarshal(bs []byte) (v Elem, n int, err error) {
	v.Value, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Prev, n1, err = s.next.Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Next, n1, err = s.next.Unmarshal(bs[n:])
	n += n1
	return
}

func (s *elemCodec) Size(v Elem) (size int) {
	return varnum.Int.Size(v.Value) + s.next.Size(v.Prev) + s.next.Size(v.Next)
}

func (s *elemCodec) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.next.Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = s.next.Skip(bs[n:])
	return n + n1, err
}

func ListSerializer() binco.Serializer[List] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	codec := &elemCodec{}
	codec.next = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, codec)

	serializer := listSerializer{codec.next}
	return ptrmap.Wrap(ptrMap, revPtrMap, serializer)
}

type listSerializer struct {
	elemPtr binco.Serializer[*Elem]
}

func (s listSerializer) Marshal(v List, bs []byte) (n int) {
	n = s.elemPtr.Marshal(v.Head, bs)
	n += s.elemPtr.Marshal(v.Tail, bs[n:])
	return n + varnum.Int.Marshal(v.Len, bs[n:])
}

func (s listSerializer) Unmarshal(bs []byte) (v List, n int, err error) {
	v.Head, n, err = s.elemPtr.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Tail, n1, err = s.elemPtr.Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Len, n1, err = varnum.Int.Unmarshal(bs[n:])
	n += n1
	return
}
```

`Size` and `Skip` follow the same pattern. The cycle between elements is
captured by the pointer codec: each element is written once, and the back
pointers become ids.

**Trees.** A tree with parent pointers is the same shape:

```go
type Node struct {
	Value    int
	Children []*Node
	Parent   *Node
}
```

Use `ptrmap.NewPtrSerializer` for `Parent` and `Children`; the root has `Parent == nil`.

**Graphs.** Vertices and edges form cycles (edge → vertex → edge). The wiring is
the interesting part: the codecs refer to each other, so build them with
variables and assign in dependency order:

```go
var (
	edgesSerializer    binco.Serializer[map[int]*Edge]
	verticesSerializer binco.Serializer[map[int]*Vertex]
	edgePtrSerializer  binco.Serializer[*Edge]
	vertexPtrSerializer binco.Serializer[*Vertex]
)
edge := edgeCodec{vertexPtr: &vertexPtrSerializer}
vertex := vertexCodec{edges: &edgesSerializer}
edgePtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, edge)
vertexPtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, vertex)
edgesSerializer = composite.NewMapSerializer(varnum.Int, edgePtrSerializer)
verticesSerializer = composite.NewMapSerializer(varnum.Int, vertexPtrSerializer)
```

A runnable version lives in [`demos/pm/graph`](./demos/pm/graph).

**Depth.** `ptrmap` writes data the first time it meets a pointer, so a deeply
nested structure is written once and read once. Nesting is bounded by
`binco.MaxDepth`; a pathologically deep graph reports `binco.ErrTooDeep` instead
of exhausting the stack.
### 8.9 Polymorphism with interfaces

An interface field can hold several concrete types. The payload must say which
one, so each implementation gets a DTM and the interface codec dispatches on it.

**Marshalling without a type switch.** If every implementation satisfies
`binco.MarshallerTyped`, the interface codec just delegates:

```go
type Command interface {
	Run() string
}

type Add struct{ Value int }

func (a Add) Run() string { return "add" }

func (a Add) MarshalTypedBinco(bs []byte) (n int) { return AddSerializer.Marshal(a, bs) }
func (a Add) SizeTypedBinco() (size int)           { return AddSerializer.Size(a) }

var AddSerializer = typeid.NewSerializer(AddDTM, addCodec{})
```

```go
type commandCodec struct{}

func (commandCodec) Marshal(v Command, bs []byte) (n int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%T does not implement binco.MarshallerTyped", v))
	}
	return m.MarshalTypedBinco(bs)
}

func (commandCodec) Size(v Command) (size int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%T does not implement binco.MarshallerTyped", v))
	}
	return m.SizeTypedBinco()
}
```

The win: a new implementation needs no change to the interface codec.

**Unmarshalling needs the table.** The concrete type is unknown until the DTM is
read:

```go
func (commandCodec) Unmarshal(bs []byte) (v Command, n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case AddDTM:
		v, n1, err = AddSerializer.UnmarshalData(bs[n:])
	case RemoveDTM:
		v, n1, err = RemoveSerializer.UnmarshalData(bs[n:])
	default:
		err = binco.NewUnexpectedDTMError(dtm)
	}
	n += n1
	return
}
```

`Skip` mirrors it with `SkipData`. A finite switch is the fast, type-safe
option; a `map[binco.DTM]func([]byte) (T, int, error)` is the extensible one. The
generator writes the switch for you (see [Interfaces](#96-interfaces)).

**Interface inside a struct.** Nothing special: the field's codec is the
interface codec, and the DTM is written where the field sits.

**"Oneof" with a marker field.** If you prefer explicit tags over DTMs, encode a
small enum before the value:

```go
type Shape struct {
	Kind  ShapeKind // 0 circle, 1 square
	Value any       // interpreted according to Kind
}
```

DTMs cost the same byte and are less error-prone, so prefer them.

### 8.10 Versioning and migration

Type versions are DTMs plus a dispatcher. There are three practical strategies.

**Strategy 1: one DTM per version (recommended).** Each version has its own
codec and its own type id. The current codec reads all versions and converts old
ones:

```go
type AddV1 struct{ Value int16 }
type AddV2 struct{ Value int }
type AddLatest AddV2

func MigrateAddV1(v AddV1) AddLatest { return AddLatest{Value: int(v.Value)} }

type versionedCodec struct{}

func (versionedCodec) Marshal(v AddLatest, bs []byte) (n int) {
	return AddV2Serializer.Marshal(AddV2(v), bs) // always writes the current version
}

func (versionedCodec) Unmarshal(bs []byte) (v AddLatest, n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case AddV1DTM:
		var old AddV1
		old, n1, err = AddV1Serializer.UnmarshalData(bs[n:])
		if err != nil {
			return
		}
		v = MigrateAddV1(old)
	case AddV2DTM:
		var cur AddV2
		cur, n1, err = AddV2Serializer.UnmarshalData(bs[n:])
		v = AddLatest(cur)
	default:
		err = binco.NewUnexpectedDTMError(dtm)
	}
	n += n1
	return
}
```

Old payloads stay readable forever; new payloads only ever contain the current
version. The generator automates this with `AddVersioned`/`RegisterVersioned`
(see [Versioned types](#97-versioned-types)).

**Strategy 2: a version field in the payload.** Add an explicit `uint8` version
as the first field of a struct and branch on it. Cheaper than a DTM if the type
is already inside a typed envelope, but it cannot express a change of the
payload's overall shape.

**Strategy 3: side-by-side types.** Keep `UserV1` and `UserV2` as separate,
complete types with their own DTMs and no migration. Suitable when the two
versions are semantically different rather than successive revisions.

**Migration rules of thumb**

* Never change the meaning of an existing DTM.
* Migrations run while decoding; keep them allocation-free where possible.
* Keep the old version's codec, not just its data, for as long as payloads can
  exist (logs, queues, archives).
* A migration that cannot fail is easier to reason about; validate afterwards
  if needed.

### 8.11 Streaming protocols

A stream has no structure of its own, so the codec supplies it. Three patterns
cover most protocols.

**Fixed sequence.** Both sides know the order, like a buffer payload:

```go
// client
wirefmt.Encode(wvarnum.Int, conn, version)
wirefmt.Encode(wcomposite.String, conn, user)
wirefmt.Encode(wcomposite.ByteSlice, conn, payload)

// server
version, _, err := wirefmt.Decode(wvarnum.Int, conn)
user, _, err := wirefmt.Decode(wcomposite.String, conn)
payload, _, err := wirefmt.Decode(wcomposite.ByteSlice, conn)
```

**Type-tagged messages (DTS).** Read a DTM per message and dispatch:

```go
func Serve(conn net.Conn) error {
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	for {
		dtm, _, err := wtypeid.DTMSerializer.Unmarshal(r)
		if errors.Is(err, io.EOF) {
			return nil // clean end of stream
		}
		if err != nil {
			return err
		}
		switch dtm {
		case LoginDTM:
			msg, n, err := LoginSerializer.UnmarshalData(r)
			if err != nil {
				return err
			}
			if _, err := wcomposite.String.Marshal("welcome "+msg.User, w); err != nil {
				return err
			}
			if err := w.Flush(); err != nil { // publish this message now
				return err
			}
		case UploadDTM:
			// Known, but not needed: skip the payload without allocating.
			if _, err := UploadSerializer.SkipData(r); err != nil {
				return err
			}
		default:
			// The size of an unknown message is unknown, so the stream cannot be
			// resynchronized: report the id and stop.
			return fmt.Errorf("unknown message: %w", binco.NewUnexpectedDTMError(dtm))
		}
	}
}
```

Three rules that make this robust:

* Treat `io.EOF` **between** messages as a clean shutdown; `io.ErrUnexpectedEOF`
  **inside** a message is a truncated wire.
* Flush after each message if the peer should react to it immediately.
* Never ignore an unknown DTM: its length is unknowable, so skipping it would
  desynchronize the wire.

**Length-prefixed frames.** When you need to buffer a message before decoding
it (for example to hand it to another goroutine), prefix it with its size:

```go
payload := binco.Encode(OrderSerializer, order)
wirefmt.Encode(wvarnum.PositiveInt, conn, len(payload))
conn.Write(payload)
```

The receiver reads the size, allocates once, and calls `io.ReadFull`. This
trades one extra length field for buffering freedom.

A runnable DTS example lives in
[`demos/typed/stream`](./demos/typed/stream).

### 8.12 Validating untrusted input

A length read from the network is an allocation request. Validating codecs check
it **before** the allocation happens, which turns a hostile payload into an
error instead of an out-of-memory kill.

```go
var (
	ErrTooLong  = errors.New("too long")
	ErrTooMany  = errors.New("too many elements")
	ErrBadKey   = errors.New("bad key")
)

boundedText := composite.NewStringSerializer(
	stropts.WithLenValidator(binco.ValidatorFn[int](func(n int) error {
		if n > 4096 {
			return fmt.Errorf("%w: %d", ErrTooLong, n)
		}
		return nil
	})),
)

boundedList := composite.NewSliceSerializer(varnum.Int,
	slopts.WithLenValidator[int](binco.ValidatorFn[int](func(n int) error {
		if n > 1000 {
			return ErrTooMany
		}
		return nil
	})),
	slopts.WithElemValidator[int](binco.ValidatorFn[int](func(v int) error {
		if v < 0 {
			return errors.New("negative element")
		}
		return nil
	})),
)

boundedMap := composite.NewMapSerializer(composite.String, composite.ByteSlice,
	mapopts.WithLenValidator[string, []byte](binco.ValidatorFn[int](func(n int) error {
		if n > 100 {
			return ErrTooMany
		}
		return nil
	})),
	mapopts.WithKeyValidator[string, []byte](binco.ValidatorFn[string](func(k string) error {
		if len(k) > 64 {
			return ErrBadKey
		}
		return nil
	})),
	mapopts.WithValueValidator[string, []byte](binco.ValidatorFn[[]byte](func(v []byte) error {
		if len(v) > 1<<20 {
			return ErrTooLong
		}
		return nil
	})),
)
```

**Order of checks**: length, then keys, then values, each as soon as the item is
decoded. So a bad length never leads to a large allocation, and a bad key stops
the decode before its value is read.

**Validators are reader-side only.** `Marshal` never validates; producing valid
data is the writer's job. A validating codec therefore costs nothing on the hot
encode path.

**What validation cannot do.** It cannot bound the total size of nested
containers (each level is checked independently) and it cannot reject a payload
that is merely *semantically* wrong (an impossible date, an unknown enum). Add
domain checks through `WithValidator`, or after decoding.

**Skipping does not validate.** `Skip` performs structural checks (flags, varint
widths, available bytes) but never calls validators, because it does not decode
values. Do not use `Skip` as a trust boundary.

### 8.13 Partial decoding and out-of-order reads

Because `Skip` is exact and `Unmarshal` never reads past a value, you can index
into a payload without decoding it.

**Read one field only.** Given a record encoded as `ID`, `Name`, `Score`:

```go
idEnd, err := varnum.Int.Skip(bs)
if err != nil {
	return err
}
nameEnd, err := composite.String.Skip(bs[idEnd:])
if err != nil {
	return err
}

// Score is the last field: jump straight to it.
score, _, err := scalar.Int32.Unmarshal(bs[idEnd+nameEnd:])
```

**Reverse order.** Decode the last field first, then the middle, then the first:

```go
score, _, err := scalar.Int32.Unmarshal(bs[idEnd+nameEnd:])
name, _, err := composite.String.Unmarshal(bs[idEnd:])
id, _, err := varnum.Int.Unmarshal(bs)
```

`demos/out_of_order` runs exactly this.

**Unknown trailing fields.** A reader for an older schema stops after the fields
it knows; the remainder is simply not read. This is how a forward-compatible
reader tolerates extra data. When a protocol carries exactly one value per
message, `binco.DecodeExact` is the strict form: it returns
`binco.ErrTrailingBytes` if anything follows the value, instead of silently
dropping it.

**Forward compatibility for containers.** If a slice may grow new element
shapes, keep the element schema fixed and add optional trailing fields inside
each element; a reader that knows fewer fields can skip the rest of each
element's bytes because the element boundary is known from its own length
prefix.

**Counting bytes.** `Unmarshal` returns `n`, the number of bytes it consumed.
Summing `n` while decoding is how a reader verifies that a payload was fully
understood:

```go
var consumed int
for _, codec := range codecs {
	n, err := codec.Skip(bs[consumed:])
	if err != nil {
		return err
	}
	consumed += n
}
if consumed != len(bs) {
	return fmt.Errorf("trailing %d bytes", len(bs)-consumed)
}
```

### 8.14 Zero-copy hot paths

`zerocopy` decodes strings and byte slices without copying: the returned value
points into the input buffer.

```go
bs := binco.Encode(zerocopy.String, "hello")
s, n, err := zerocopy.String.Unmarshal(bs) // s aliases bs, no allocation
```

This is the fastest option and the easiest to get wrong. Rules:

1. **The value lives only as long as the buffer.** If the buffer is reused, the
   string changes under you:

   ```go
   var buf []byte
   for _, msg := range messages {
   	buf = binco.Encode(zerocopy.String, msg)
   	s, _, _ := zerocopy.String.Unmarshal(buf)
   	_ = s // fine here: processed before the next iteration
   	keep = append(keep, s) // wrong: every kept string aliases the same buffer
   }
   ```

2. **Never write through the returned slice.** It is the payload.

3. **Copy when it must outlive the buffer**: `strings.Clone(s)`,
   `bytes.Clone(b)`, or use `composite.String`/`composite.ByteSlice`, which copy while
   decoding.

4. **Do not mix modes for one value by accident.** Importing `zerocopy` and `composite`
   together is legitimate (bytes are identical), but be explicit about which
   fields are views.

`demos/native` demonstrates the pitfall and its fixes.

### 8.15 Custom codecs

A codec is any type with four methods, so domain types are easy:

**Fixed-size identifier (UUID).**

```go
type UUID [16]byte

type uuidCodec struct{}

func (uuidCodec) Marshal(v UUID, bs []byte) (n int) {
	return copy(bs[:16], v[:])
}

func (uuidCodec) Unmarshal(bs []byte) (v UUID, n int, err error) {
	if len(bs) < 16 {
		return v, 0, binco.ErrTruncated
	}
	copy(v[:], bs[:16])
	return v, 16, nil
}

func (uuidCodec) Size(v UUID) (size int) { return 16 }

func (uuidCodec) Skip(bs []byte) (n int, err error) {
	if len(bs) < 16 {
		return 0, binco.ErrTruncated
	}
	return 16, nil
}
```

**Decimal-ish amounts.** Store integer minor units; never store floats for money.

```go
type Cents int64

type centsCodec struct{}

func (centsCodec) Marshal(v Cents, bs []byte) (n int) {
	return varnum.Int64.Marshal(int64(v), bs)
}
func (centsCodec) Unmarshal(bs []byte) (v Cents, n int, err error) {
	c, n, err := varnum.Int64.Unmarshal(bs)
	return Cents(c), n, err
}
func (centsCodec) Size(v Cents) (size int) { return varnum.Int64.Size(int64(v)) }
func (centsCodec) Skip(bs []byte) (n int, err error) { return varnum.Int64.Skip(bs) }
```

**Net addresses.** Prefer the fixed-size form (4 or 16 bytes) over text:

```go
type IPv4 [4]byte // 4 bytes, validated on decode if you need a mask
```

**Wrapper codecs** are useful for cross-cutting behaviour. A codec that logs
sizes, counts calls, or enforces a budget composes without touching the payload:

```go
type countingCodec[T any] struct {
	inner binco.Serializer[T]
	bytes *int64
}

func (c countingCodec[T]) Marshal(v T, bs []byte) (n int) {
	n = c.inner.Marshal(v, bs)
	*c.bytes += int64(n)
	return n
}

func (c countingCodec[T]) Unmarshal(bs []byte) (T, int, error) {
	return c.inner.Unmarshal(bs)
}

func (c countingCodec[T]) Size(v T) int { return c.inner.Size(v) }

func (c countingCodec[T]) Skip(bs []byte) (int, error) { return c.inner.Skip(bs) }
```

### 8.16 Composition recursion and generics

Codecs are values, and Go lets you build them with functions. That covers most
"not invented here" needs without a DSL.

**Recursive types** need a serializer that can refer to itself. Because a struct
cannot hold its own codec by value, hold it behind a pointer and assign after
construction:

```go
type Tree struct {
	Value    int
	Children []Tree
}

type treeCodec struct {
	self *binco.Serializer[Tree] // set after construction
}

func (c treeCodec) childrenSerializer() binco.Serializer[[]Tree] {
	return composite.NewSliceSerializer(*c.self)
}

func (c treeCodec) Marshal(v Tree, bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Value, bs)
	return n + c.childrenSerializer().Marshal(v.Children, bs[n:])
}

func (c treeCodec) Unmarshal(bs []byte) (v Tree, n int, err error) {
	v.Value, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Children, n1, err = c.childrenSerializer().Unmarshal(bs[n:])
	n += n1
	return
}

func (c treeCodec) Size(v Tree) (size int) {
	return varnum.Int.Size(v.Value) + c.childrenSerializer().Size(v.Children)
}

func (c treeCodec) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = c.childrenSerializer().Skip(bs[n:])
	return n + n1, err
}

func NewTreeSerializer() binco.Serializer[Tree] {
	c := &treeCodec{}
	self := binco.Serializer[Tree](c)
	c.self = &self
	return c
}
```

**Generics.** A codec for `Pair[A, B]` is a function of two codecs:

```go
type Pair[A, B any] struct {
	First  A
	Second B
}

func NewPairSerializer[A, B any](first binco.Serializer[A], second binco.Serializer[B]) binco.Serializer[Pair[A, B]] {
	return pairCodec[A, B]{first, second}
}
```

**Adapters.** Converting a value before encoding is sometimes cleaner than
writing a codec. For a defined type whose underlying type is `int`:

```go
type Int int

type intCodec struct{}

func (intCodec) Marshal(v Int, bs []byte) (n int) { return varnum.Int.Marshal(int(v), bs) }
func (intCodec) Unmarshal(bs []byte) (v Int, n int, err error) {
	i, n, err := varnum.Int.Unmarshal(bs)
	return Int(i), n, err
}
func (intCodec) Size(v Int) (size int) { return varnum.Int.Size(int(v)) }
func (intCodec) Skip(bs []byte) (n int, err error) { return varnum.Int.Skip(bs) }
```

The generator does exactly this for `AddDefinedType`.

### 8.17 Zero-allocation encoding

The buffer API exists so a hot path can allocate nothing.

**Reuse one buffer per goroutine.** `Size` is exact, so a buffer grown once is
enough forever after:

```go
type Encoder struct {
	buf []byte
}

func (e *Encoder) Encode(v Order) []byte {
	size := OrderSerializer.Size(v)
	if cap(e.buf) < size {
		e.buf = make([]byte, size)
	}
	e.buf = e.buf[:size]
	OrderSerializer.Marshal(v, e.buf)
	return e.buf // valid until the next call
}
```

The returned slice aliases the internal buffer: copy it (or write it out)
before the next call.

**Avoid `binco.Encode` in loops.** It allocates by design:

```go
for _, order := range orders {
	bs := binco.Encode(OrderSerializer, order) // one allocation per iteration
	conn.Write(bs)
}
```

is convenient, while

```go
for _, order := range orders {
	size := OrderSerializer.Size(order)
	if cap(buf) < size {
		buf = make([]byte, size)
	}
	buf = buf[:size]
	OrderSerializer.Marshal(order, buf)
	conn.Write(buf)
}
```

does not allocate once the buffer has grown.

**sync.Pool** helps when encoders live in short-lived handlers. Return the
buffer to the pool only after it has been written out:

```go
var bufPool = sync.Pool{New: func() any { b := make([]byte, 0, 512); return &b }}

func writeOrder(w io.Writer, v Order) error {
	bp := bufPool.Get().(*[]byte)
	defer bufPool.Put(bp)

	buf := *bp
	if size := OrderSerializer.Size(v); cap(buf) < size {
		buf = make([]byte, size)
	} else {
		buf = buf[:size]
	}
	OrderSerializer.Marshal(v, buf)
	*bp = buf
	_, err := w.Write(buf)
	return err
}
```

**Decoding allocations.** Slices, maps and strings allocate by construction.
`zerocopy.String`/`zerocopy.ByteSlice` are the only zero-allocation decoders, and
`Skip` is the only way to decode nothing. For a fixed-size message, decode into
reused fields (a slice's backing array can be reused by decoding into
`v.Items[:0]` and appending, which is what `composite` cannot do for you - it
allocates a new slice each time).

### 8.18 Concurrency

* **Buffer codecs are stateless** (except `ptrmap`), so they are safe to share
  across goroutines. A single `varnum.Int` in a global is fine.
* **The plain pointer codec counts nesting in an atomic.** `composite.NewPtrSerializer`
  keeps one atomic counter per serializer to bound recursion; it stays free of
  data races, and only extremely deep *concurrent* nesting (tens of thousands of
  levels at once) could reach the limit. `ptrmap` counts per decoded value
  instead.
* **`ptrmap` codecs are stateful.** The maps are per-value state, so each goroutine
  needs its own maps and its own wrapper:

  ```go
  func perRequest() binco.Serializer[Graph] {
      ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
      return ptrmap.Wrap(ptrMap, revPtrMap, newGraphSerializer(ptrMap, revPtrMap))
  }
  ```

* **Buffers are not shared.** One buffer per goroutine, or a `sync.Pool`.
* **Writers are not safe for concurrent use.** Serialize writes (or give each
  goroutine its own buffer and a channel to the writer).
* **A `codegen.Generator` is not safe for concurrent use.** It records the types
  it renders, the packages it has seen and the names it has handed out; create one
  per generation run.

### 8.19 Testing your codecs

**Round trip.** For every codec, test values that stress the encoding: zero,
one, boundaries of the varint width, empty containers, nil pointers.

```go
func TestOrderCodec(t *testing.T) {
	orders := []Order{
		{},
		{ID: 1, Items: []Item{{Name: "a"}}, Note: "", Total: -1},
		{ID: 1 << 40, Items: make([]Item, 100), Note: "note", Total: 1 << 50},
	}
	for _, want := range orders {
		bs := binco.Encode(orderCodec{}, want)
		if len(bs) != orderCodec{}.Size(want) {
			t.Fatalf("size mismatch")
		}
		got, err := binco.Decode(orderCodec{}, bs)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
		n, err := orderCodec{}.Skip(bs)
		if err != nil || n != len(bs) {
			t.Fatalf("skip = %d, %v", n, err)
		}
	}
}
```

**Truncation.** Every prefix of a valid payload must fail cleanly:

```go
for i := 0; i < len(bs); i++ {
	if _, _, err := orderCodec{}.Unmarshal(bs[:i]); err == nil {
		t.Fatalf("prefix %d decoded without error", i)
	}
}
```

**Fuzzing.** The library's own tests use `testing.F` round trips; the same shape
works for your codecs:

```go
func FuzzOrder(f *testing.F) {
	f.Add("", 0, int64(0))
	f.Fuzz(func(t *testing.T, note string, id int, total int64) {
		v := Order{Note: note, Total: total}
		if id > 0 {
			v.ID = UserID(id)
		}
		bs := binco.Encode(orderCodec{}, v)
		got, err := binco.Decode(orderCodec{}, bs)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if got.Note != v.Note || got.Total != v.Total {
			t.Fatalf("mismatch: %+v vs %+v", got, v)
		}
	})
}
```

**Deterministic maps.** When asserting bytes, encode a single-entry map or the
sorted-slice form; multi-entry maps have random order.

### 8.20 Interoperability

**With JSON.** The two encodings solve different problems; a service can return
JSON to people and binco to services. Keep a single source of truth for the
struct and two marshallers:

```go
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

bs := binco.Encode(UserBinco, user)         // wire format
js, _ := json.Marshal(user)               // human-facing
```

**Protobuf-style encoding.** The primitives map directly: varint for numeric
fields, length-prefixed values for `string`/`bytes`, a varint tag in
front of each field if you want Protobuf's wire layout. The library does not
implement Protobuf, but its codecs are a building block for such formats.

**Files and object storage.** Prefix a file with a DTM if it can contain several
types, or write one type per file and keep the type in the file name or
metadata. A common envelope is:

```
[DTM][compressed payload][checksum]
```

where the checksum covers the DTM and the payload.

**Caches.** Cache entries are values: encode the key's scope, the payload and
an expiry timestamp, and decode defensively (validators, bounded lengths).

**Database blobs.** The same applies: store a DTM plus the payload, and version
the schema with the DTM strategy. Keep the codec of every version for as long as
rows can exist.

**Across languages.** The format rules are in
[The wire format](#5-the-wire-format): LEB128 varint, ZigZag for signed values,
little-endian raw numbers, positive varints for lengths, and the pointer flags.
Any implementation that follows them interoperates.
## 9. Code generation

`binco/codegen` writes the serializer code you would otherwise write by hand. It
inspects types through reflection **once**, at generation time; the generated
code contains no reflection.

### 9.1 Getting started

```
app/
 |--- go.mod
 |--- order.go
 |--- gen/
      |--- main.go
```

**order.go**

```go
//go:generate go run ./gen
package app

type Order struct {
	ID    uint64
	Items []Item
	Note  string
	Total int64
}

type Item struct {
	SKU   string
	Count int
}
```

**gen/main.go**

```go
package main

import (
	"log"
	"os"
	"reflect"

	"example.com/app"

	"github.com/malivvan/binco/codegen"
	genopts "github.com/malivvan/binco/codegen/config/gen"
)

func main() {
	g, err := codegen.NewGenerator(
		genopts.WithPath("example.com/app"), // required
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, add := range []error{
		g.AddStruct(reflect.TypeFor[app.Item]()),
		g.AddStruct(reflect.TypeFor[app.Order]()),
	} {
		if add != nil {
			log.Fatal(add)
		}
	}
	bs, err := g.Generate()
	if err != nil {
		// The source is returned even on failure, so it can be inspected.
		if writeErr := os.WriteFile("failed.gen.go", bs, 0o644); writeErr != nil {
			log.Fatal(writeErr)
		}
		log.Fatal(err)
	}
	if err = os.WriteFile("order_binco.gen.go", bs, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

```bash
go generate        # writes order_binco.gen.go
go build
```

The generated code contains `ItemBinco` and `OrderBinco`, both implementing
`binco.Serializer[Item]` and `binco.Serializer[Order]`.

### 9.2 Generator configuration

```go
g, err := codegen.NewGenerator(
	genopts.WithPath("github.com/user/project/pkg"), // required
	genopts.WithPackage("pkg"),                      // package name
	genopts.WithImport("time"),                      // extra import
	genopts.WithImportAlias("github.com/x/y", "y"),  // extra import + alias
	genopts.WithSerializer(reflect.TypeFor[other.Type](), "other.CustomSerializer"),
	genopts.WithDTM(reflect.TypeFor[event.Login](), 7), // pin a wire id
	genopts.WithLogger(logger),                          // destination of warnings
	genopts.WithStream(), // generate for io.Writer/io.Reader
	genopts.WithUnsafe(), // use the zero-copy codecs
	// genopts.WithNotUnsafe(), // zero-copy except for strings
)
```

| Option | Effect |
|--------|--------|
| `WithPath` | import path of the package the generated file belongs to (required) |
| `WithPackage` | name of that package (default: the last path element) |
| `WithImport` | add an import path |
| `WithImportAlias` | add an import path with an alias; also used to resolve cross-package type names |
| `WithSerializer` | custom serializer name for a type (may carry a package) |
| `WithDTM` | pin the wire identifier of a versioned type or interface implementation |
| `WithLogger` | receive the generator's warnings; the default discards them |
| `WithStream` | generate streaming codecs |
| `WithUnsafe` | generate with `binco/zerocopy` everywhere |
| `WithNotUnsafe` | generate with `binco/zerocopy`, except for strings |

**Modes**

* *Portable* (default): `varnum`, `scalar`, `composite`. No `unsafe` package at all.
* *Unsafe*: `binco/zerocopy` for numbers, bools, strings, byte slices and time.
* *Not unsafe*: like unsafe, but strings use the copying `composite` codec, which
  removes the zero-copy side effects.

`Generate` wraps `ErrCodeGenFailed` around the real cause, so
`errors.Is(err, codegen.ErrCodeGenFailed)` matches and the message names the
template, the invalid expression or the name collision that failed. The returned
bytes are still useful to see what went wrong. Bad validator and migration
expressions, an unexported field of a foreign struct and two types that would be
generated under the same name are all reported when the type is added, not as a
compile error inside the output.

### 9.3 Generator methods

```go
g.AddDefinedType(t reflect.Type, opts ...tpopts.SetOption) error
g.AddStruct(t reflect.Type, opts ...stopts.SetOption) error
g.AddTyped(t reflect.Type, opts ...tpopts.SetOption) error
g.AddInterface(t reflect.Type, opts ...intropts.SetOption) error
g.RegisterInterface(t reflect.Type, opts ...intropts.SetRegisterOption) error
g.AddVersioned(t reflect.Type, opts ...veropts.SetOption) error
g.RegisterVersioned(t reflect.Type, opts ...veropts.SetOption) error
g.Generate() ([]byte, error)
g.DTMRegistry() (map[string]binco.DTM, error)   // type name -> wire id, for a golden file
```

**AddDefinedType** - a type whose underlying type is a number, string, array,
slice, map or pointer:

```go
type Int int
type Names []string
type Lookup map[string]int
type UintPtr *uint

g.AddDefinedType(reflect.TypeFor[Int](),
	tpopts.WithNumEncoding(tpopts.NumEncodingRaw), // scalar.Int for the underlying int
	tpopts.WithValidator("ValidateInt"),           // called after decoding
)
```

**AddStruct** - a defined struct:

```go
g.AddStruct(reflect.TypeFor[Order](),
	stopts.WithValidator("ValidateOrder"),
	// One option set per field, in declaration order:
	stopts.WithField(),
	stopts.WithField(fldopts.WithIgnore()),
	stopts.WithField(fldopts.WithType(
		tpopts.WithLenValidator("ValidateNoteLength"),
	)),
)
```

Passing any `stopts.WithField` means passing exactly one per field; the
generator rejects a mismatch.

A `time.Time` wrapper is declared with `WithUnderlyingTime`:

```go
type Timestamp time.Time

g.AddStruct(reflect.TypeFor[Timestamp](),
	stopts.WithUnderlyingTime(
		stopts.WithUnderlyingTimeUnit(tpopts.TimeUnitMilli),
	),
)
```

**AddTyped** - a type identifier in front of an existing serializer:

```go
g.AddStruct(reflect.TypeFor[Order]())
g.AddTyped(reflect.TypeFor[Order]()) // declares OrderDTM and OrderTypedBinco
```

**AddInterface** - a codec for an interface, from a list of implementations that
already have typed serializers:

```go
g.AddStruct(reflect.TypeFor[Circle]())
g.AddTyped(reflect.TypeFor[Circle]())
g.AddInterface(reflect.TypeFor[Shape](),
	intropts.WithImpl(reflect.TypeFor[Circle]()),
)
```

**RegisterInterface** - the convenience form: adds each implementation (struct
or defined type), its typed serializer, the DTM constants and the interface
codec:

```go
g.RegisterInterface(reflect.TypeFor[Shape](),
	intropts.WithStructImpl(reflect.TypeFor[Circle]()),
	intropts.WithDefinedTypeImpl(reflect.TypeFor[Square]()),
	// intropts.WithRegisterMarshaller(), // delegate to MarshalTypedBinco
)
```

**AddVersioned / RegisterVersioned** - a codec that reads several versions and
migrates old ones:

```go
type FooV1 int
type FooV2 string
type Foo FooV2

func MigrateFooV1(v FooV1) Foo { return Foo(FooV2("migrated")) }

g.RegisterVersioned(reflect.TypeFor[Foo](),
	veropts.WithVersion(reflect.TypeFor[FooV1](), "MigrateFooV1"),
	veropts.WithCurrentVersion(reflect.TypeFor[FooV2]()),
)
```

Exactly one version must be the current one.

### 9.4 Generated code and naming

For each added type the generator writes:

```go
var OrderBinco = orderBinco{}

type orderBinco struct{}

func (s orderBinco) Marshal(v Order, bs []byte) (n int)              { ... }
func (s orderBinco) Unmarshal(bs []byte) (v Order, n int, err error) { ... }
func (s orderBinco) Size(v Order) (size int)                         { ... }
func (s orderBinco) Skip(bs []byte) (n int, err error)               { ... }
```

| Symbol | Name |
|--------|------|
| serializer variable | `<Type>Binco` |
| serializer type | `<type>Binco` (unexported) |
| typed serializer | `<Type>TypedBinco` |
| type identifier | `<Type>DTM` |
| anonymous container codec | hashed name, one per distinct type + options |

The file starts with

```go
// Code generated by binco/codegen. DO NOT EDIT.
```

so tooling treats it as generated. Anonymous codecs (the serializer of a
`[]Item` field, for example) are declared once in a single `var (...)` block at
the top of the file.

### 9.5 Serialization options

Options are attached to a type, then to a field with
`fldopts.WithType(...)`.

```go
tpopts.WithNumEncoding(tpopts.NumEncodingVarint)          // varnum.Int
tpopts.WithNumEncoding(tpopts.NumEncodingVarintPositive)  // varnum.PositiveInt
tpopts.WithNumEncoding(tpopts.NumEncodingRaw)             // scalar.Int
tpopts.WithTimeUnit(tpopts.TimeUnitMilli)
tpopts.WithValidator("Validate")
tpopts.WithLenEncoding(tpopts.NumEncodingRaw)
tpopts.WithLenValidator("ValidateLength")
tpopts.WithElemValidator("ValidateElem")
tpopts.WithKeyValidator("ValidateKey")
```

| Type | Applicable options |
|------|--------------------|
| numbers | `WithNumEncoding`, `WithValidator` |
| `string` | `WithLenEncoding`, `WithLenValidator`, `WithValidator` |
| array | `WithLenEncoding`, `WithElemValidator`, `WithValidator` |
| slice | `WithLenEncoding`, `WithLenValidator`, `WithElemValidator`, `WithValidator` |
| map | `WithLenEncoding`, `WithLenValidator`, `WithKeyValidator`, `WithElemValidator`, `WithValidator` |
| `time.Time` | `WithTimeUnit`, `WithValidator` |

Options that do not apply to a type are ignored, so a single options slice can
be passed around. Validators are referenced by Go expression, which means the
function or variable must be visible in the target package:

```go
func ValidateNoteLength(n int) error {
	if n > 1024 {
		return errors.New("note too long")
	}
	return nil
}
```

### 9.6 Interfaces

The generated interface codec writes a DTM and dispatches on it:

```go
var ShapeBinco = shapeBinco{}

func (s shapeBinco) Marshal(v Shape, bs []byte) (n int) {
	switch t := v.(type) {
	case Circle:
		return CircleTypedBinco.Marshal(t, bs)
	case Square:
		return SquareTypedBinco.Marshal(t, bs)
	default:
		panic(fmt.Sprintf(binco.ErrorPrefix+"unexpected %v type", t))
	}
}
```

With `intropts.WithMarshaller()` (or `WithRegisterMarshaller()` in the register
form) the switch is replaced by a delegation to `MarshalTypedBinco`, which lets
you add implementations without regenerating the interface codec:

```go
if m, ok := v.(binco.MarshallerTyped); ok {
	return m.MarshalTypedBinco(bs)
}
panic(...)
```

Either way the receiver switches on the DTM and reports `UnexpectedDTMError` for
unknown ids. A nil interface value is written as the reserved identifier `0` and
decodes back to nil, so `Marshal(nil)` no longer panics.

### 9.7 Versioned types

The generated versioned codec is exactly the hand-written one from
[Versioning and migration](#810-versioning-and-migration): `Marshal` writes the
current version, `Unmarshal` accepts every version and migrates, `Skip` advances
over every version.

Pin the identifiers with `genopts.WithDTM` in anything that is stored or sent.
Without it the generator numbers the types in registration order, which is stable
within one run but renumbers the payloads as soon as a type is added or removed.
Generated constants carry literal values, a duplicate id or a negative id fails
generation with `binco.ErrDuplicateDTM`, and mixing pinned and automatic ids is
rejected. `g.DTMRegistry()` returns the mapping so you can commit it as a golden
file and compare it in CI.

### 9.8 Multi-package and cross-package generation

By default the generator assumes that `bar.Bar`'s serializer lives in package
`bar` and is called `BarBinco`. If it lives elsewhere, say so:

```go
g, _ := codegen.NewGenerator(
	genopts.WithPath("example.com/foo"),
	genopts.WithImportAlias("example.com/serializers", "serializers"),
	genopts.WithSerializer(reflect.TypeFor[bar.Bar](), "serializers.CustomBar"),
)
```

Generated references then use `serializers.CustomBarBinco`.

To generate a serializer **for** a type from another package, point
`WithPath` at the destination and add the type as usual. The generated import
block is built from the packages of every type the generator resolved, so a
`time.Time` field or a type from another package imports itself. Add
`WithImport` or `WithImportAlias` for a package that appears only inside a
validator or migration expression, since no type refers to it. Unused imports are
pruned and the result is gofmt-formatted.

### 9.9 Streaming mode

`genopts.WithStream()` changes the signatures:

```go
func (s orderBinco) Marshal(v Order, w wirefmt.Writer) (n int, err error)
func (s orderBinco) Unmarshal(r wirefmt.Reader) (v Order, n int, err error)
func (s orderBinco) Size(v Order) (size int)
func (s orderBinco) Skip(r wirefmt.Reader) (n int, err error)
```

and the imports become `binco/wirefmt/...`. Byte slices are encoded with
`binco/wirefmt/composite` (a stream cannot hand out a view of unread bytes); everything
else mirrors the buffer codecs. A streaming codec can also be used in the buffer
API by copying through a `bytes.Buffer`, but normally you pick one mode per
package.

### 9.10 Workflow and troubleshooting

**Recommended workflow**

1. Add the `//go:generate go run ./gen` directive to the file that declares the
   generated types.
2. Write `gen/main.go`, run `go generate ./...`.
3. Commit the generated file: builds stay hermetic and diffs show schema
   changes.
4. Run the generator in CI and fail when the output changes unexpectedly.

**Common problems**

| Symptom | Cause and fix |
|---------|---------------|
| `code generation failed` and the source looks unfinished | a missing import; add `genopts.WithImport(...)` and inspect the returned bytes |
| `use the Generator.AddStruct method` | a struct was passed to `AddDefinedType` (or vice versa) |
| `not enough structops.WithField() calls` | one `WithField` per field, always |
| generated code refers to a package that does not exist | the target package path is wrong, or the field type needs `WithImportAlias` |
| `cannot get a serializer for ...` error | a field type has no codec; add a defined type, a `WithSerializer`, or change the schema |
| `generated name collision` | two types with the same bare name; give one `genopts.WithSerializer` |
| `... is not a Go expression` | a validator or migration option contains a typo |
| `... has the unexported field ...` | a struct from another package must export every field to be generated |

## 10. Performance guide

**Plan for the buffer.** `Size` is exact and cheap; use it to allocate once and
reuse. The `Encoder` pattern in
[Zero-allocation encoding](#817-zero-allocation-encoding) is the recommended
shape for hot paths.

**Encode fixed-width slices in bulk.** A slice whose element codec reports a
fixed width (`scalar` and `zerocopy` scalars) is copied in one operation instead of
one interface call per element, provided the host is little-endian. The fast path
is automatic; element validators keep the per-element path.

**Pick the codec family per field.**

* `varnum` for small and medium integers, ids, counters, lengths.
* `scalar` for values above the thresholds in [Fixed width](#54-fixed-width), for constant-size
  records, and when CPU matters more than bytes.
* `zerocopy` for scalars and blobs on paths where the aliasing rules are
  acceptable; it is correct on every byte order.

**Skip what you do not need.** `Skip` costs a few branches and no allocation.
Decoding a field you immediately discard is pure overhead.

**Prefer the streaming API for large payloads.** It never materializes the whole
message. Use a `bufio` reader/writer at the boundary: the adapters write one
byte at a time.

**Batch small writes.** A stream of tiny messages is dominated by `Write`
syscalls; buffer (bufio) and flush on message boundaries or when the buffer
fills.

**Avoid reflection.** The generated code is allocation-free and inlinable. If
you must use reflection, do it once at startup to build codecs, not per value.

**Benchmark the real shape.** The cost of a payload is dominated by memory
traffic; measure with your structs rather than synthetic scalars. `go test
-bench` with `-benchmem` is enough to see whether a change allocates.

## 11. Correctness guarantees and security

**Guarantees**

* `Size(v)` equals the bytes `Marshal(v, bs)` writes; `Unmarshal` consumes
  exactly that many; `Skip` advances exactly as far.
* Encoding is deterministic except for maps (iteration order) and pointer ids
  (assignment order). Everything else is a pure function of the value.
* Buffer codecs and streaming codecs emit identical bytes; `scalar`/`zerocopy` agree
  on numbers; `composite`/`zerocopy` agree on strings and byte slices.
* Decoding never reads past the value it returns and never panics on malformed
  input.
* Lengths are checked without overflow: a decoded `MaxInt64` length returns
  `ErrTruncated` instead of wrapping into a slice panic.
* A decoded length never drives an allocation larger than the input that could
  contain it; a streaming blob grows with the bytes that actually arrive.
* Varints are canonical: every value has exactly one byte representation.
* The generator's output compiles and behaves like hand-written code; the test
  suite regenerates, compiles and runs it, and CI fails when the committed output
  differs from a fresh run.

**Security notes**

* Treat every length as hostile: the codecs already cap what a length can make
  them allocate, and the validating codecs let you reject implausible ones before
  they are read.
* Reject unknown DTMs in protocols; do not attempt to skip them.
* A payload can legitimately be empty; distinguish "empty message" from
  "connection closed" by protocol design, not by guessing.
* Pointer identifiers are typed: a payload cannot reinterpret one type as another
  (`ErrWrongType`), and two values that share an address get different ids.
* Nested pointers stop at `binco.MaxDepth` / `binco.MaxPointerDepth` with
  `ErrTooDeep` instead of exhausting the stack. Containers whose elements consume
  no bytes are capped by `binco.MaxZeroSizeElements`.
* `binco/zerocopy` converts scalars explicitly, so it is safe on any byte order and
  at unaligned offsets. Its strings and slices still alias the input buffer.
* The generator validates validator and migration expressions before it emits
  them, replaces panics with errors, and reports name collisions instead of
  silently merging two types.

## 12. Error reference

| Error | Returned when |
|-------|---------------|
| `binco.ErrOverflow` | a varint carries more bits than the type has, or a value does not fit the host word |
| `binco.ErrNegativeLength` | a decoded length is negative |
| `binco.ErrNonCanonical` | a varint is not in its shortest form |
| `binco.ErrWrongFormat` | a flag byte or shape is invalid (bool > 1, unknown pointer flag) |
| `binco.ErrTruncated` | the input ends inside a value |
| `binco.ErrBufferTooSmall` | Marshal would write past the output buffer |
| `binco.ErrTooLargeLength` | a length exceeds a fixed maximum (zero-size elements, array overflow) |
| `binco.ErrWrongLength` | a length does not match the type (array shorter or longer than its type) |
| `binco.ErrTooDeep` | a value nests deeper than `binco.MaxDepth` / `binco.MaxPointerDepth` |
| `binco.ErrWrongType` | a pointer identifier was defined by another type |
| `binco.ErrTrailingBytes` | `binco.DecodeExact` read a complete value followed by more bytes |
| `binco.ErrWrongArrayType` | `NewArraySerializer` was built with mismatched type parameters |
| `binco.ErrNilInterface` | reserved for a nil interface where a value was required |
| `binco.ErrDuplicateDTM` | the generator assigned the same type identifier twice |
| `codegen.ErrNameCollision` | two types would be generated under the same name |
| `binco.ErrUnsupportedIntSize` | kept for compatibility; no code path raises it now that `int` has a fixed wire width |
| `binco.ErrUnsupportedArch` | kept for compatibility; `zerocopy` no longer requires a little-endian host |
| `binco.UnexpectedDTMError` | a dispatcher read an unknown type identifier |
| `binco.WrongDTMError` | `typeid.Serializer` read a different identifier than its own |
| `io.EOF` (streams) | the stream ended cleanly between values |
| `io.ErrUnexpectedEOF` (streams) | the stream ended inside a value |
| `ErrCodeGenFailed` (`codegen`) | the generated source does not parse |

All library errors are prefixed with `binco: ` and support `errors.Is`;
the error types expose their details:

```go
var wrong binco.WrongDTMError
if errors.As(err, &wrong) {
	fmt.Println(wrong.Expected(), wrong.Actual())
}
```

## 13. Cheat sheet

```go
// Encoding and decoding
bs := binco.Encode(varnum.Int, 42)
v, err := binco.Decode(varnum.Int, bs)
bs = binco.Marshal(v)        // v implements binco.Marshaller
bs = binco.MarshalTyped(v)   // v implements binco.MarshallerTyped

// Without allocation
bs = make([]byte, serializer.Size(v))
n := serializer.Marshal(v, bs)

// Containers
composite.NewSliceSerializer(varnum.Int)
composite.NewMapSerializer(composite.String, varnum.Int)
composite.NewPtrSerializer(composite.String)
zerocopy.NewArraySerializer[[3]int, int](varnum.Int)

// Validation
composite.NewStringSerializer(stropts.WithLenValidator(vl))
composite.NewSliceSerializer(elem, slopts.WithLenValidator[int](vl), slopts.WithElemValidator[int](evl))
composite.NewMapSerializer(k, v, mapopts.WithLenValidator[K, V](vl), mapopts.WithKeyValidator[K, V](kvl))

// Type identifiers
typeid.NewSerializer(MyDTM, myCodec)
typeid.DTMSerializer.Unmarshal(bs)
serializer.UnmarshalData(bs) / serializer.SkipData(bs)

// Pointer mapping
ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
ptrmap.NewPtrSerializer(ptrMap, revPtrMap, base)
ptrmap.Wrap(ptrMap, revPtrMap, outer)

// Zero copy
zerocopy.String.Unmarshal(bs)      // aliases bs
zerocopy.ByteSlice.Unmarshal(bs)   // aliases bs
strings.Clone(s) / bytes.Clone(b) // make it independent

// Streaming
wirefmt.Encode(wcomposite.String, w, "hi")
wirefmt.Decode(wcomposite.String, r)
wirefmt.NewWriter(w) / wirefmt.NewReader(r)

// Generator
g, err := codegen.NewGenerator(genopts.WithPath(path))
g.AddStruct(reflect.TypeFor[T]())
bs, err := g.Generate()
```

## 14. FAQ

**Can I use binco for data that must be readable by a program in another
language?** Yes. The format is described in [The wire format](#5-the-wire-format)
and uses only LEB128 varints, ZigZag, little-endian fixed-width integers, length
prefixes and the pointer flags. You need to replicate the schema, not a runtime.

**Why are there no field tags?** Because the schema is code. Tags would add
bytes to every payload; the receiver already knows the order.

**Can I skip a field I do not understand?** If you know its codec, yes:
`Skip`. If you do not know its length (an unknown DTM, an unknown schema), no;
stop and report it.

**Do I need the generator?** No. The hand-written codec in
[Structs by hand](#83-structs-by-hand) is exactly what the generator emits.

**Can I serialize a struct with unexported fields?** Yes, in the same package
where it is declared (`reflect` is only used by the generator, which cannot see
unexported fields of other packages). The generated code works on any field
visible at the generation site.

**Are maps deterministic?** No. Decoding is; encoding follows Go's map iteration
order. Use sorted slices for canonical payloads.

**Is there a size limit?** Lengths are `int` values, so the limit is the
platform's `int` and available memory. Validate lengths at trust boundaries.

**Can I mix buffer and streaming codecs?** Yes for the format: bytes are
identical. You cannot hand a `[]byte` to a stream codec without a
`bytes.Reader`/`bytes.Buffer` in between.

**What happens if the payload has trailing bytes?** Nothing: `Unmarshal`
consumes one value. Check the returned `n` if you require the payload to be
fully consumed.

**Does `ptrmap` work with the streaming API?** Yes: `wirefmt/ptrmap` mirrors `ptrmap`; the
same rules apply.

**Is the generated code stable across versions?** The format is; the generated
source can change when the generator improves. Commit the generated file and
review diffs, exactly like any other generated code.

## 15. Working on binco

The architecture, repository layout, commands, testing strategy, conventions
and the full list of runnable examples live in [`AGENTS.md`](./AGENTS.md),
which is the entry point for contributors and coding agents. The user-facing
documentation is this file; the wire format in
[The wire format](#5-the-wire-format) is the shared reference.

## License

MIT. See [LICENSE](./LICENSE).
