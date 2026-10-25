# binco [![Go Reference](https://pkg.go.dev/badge/github.com/malivvan/binco.svg)](https://pkg.go.dev/github.com/malivvan/binco) [![Release](https://img.shields.io/github/v/release/malivvan/binco.svg?sort=semver)](https://github.com/malivvan/binco/releases/latest) [test](https://github.com/malivvan/binco/actions/workflows/ci.yml/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/malivvan/binco/badge.svg?branch=master)](https://coveralls.io/github/malivvan/binco?branch=master) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

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
every kind of data you are likely to serialize. New here? The
[Start here](#start-here) table maps a task to the section that answers it. If
you are working on the library itself, [`AGENTS.md`](./AGENTS.md) covers the
architecture, conventions and commands.

## Start here

| Goal | Where to look |
|------|---------------|
| Encode and decode your first value | [Quick start](#3-quick-start) |
| Choose an encoding for a field | [Choosing codecs](#5-choosing-codecs) |
| Write or generate a codec for a struct | [Structs by hand](#63-structs-by-hand), [Code generation](#7-code-generation) |
| Send values over a connection | [Streaming protocols](#611-streaming-protocols) |
| Decode untrusted input safely | [Validating untrusted input](#612-validating-untrusted-input), [Correctness guarantees and security](#11-correctness-guarantees-and-security) |
| Look up an error, a symbol or the bytes | [Error reference](#10-error-reference), [Package guide](#8-package-guide), [The wire format](#13-the-wire-format) |
| Change the library itself | [AGENTS.md](./AGENTS.md) |

## Contents

**Getting started**

* [Why binco](#1-why-binco)
* [Install and requirements](#2-install-and-requirements)
* [Quick start](#3-quick-start)
  * [A number](#31-a-number)
  * [A string](#32-a-string)
  * [A struct by hand](#33-a-struct-by-hand)
  * [The same struct from the generator](#34-the-same-struct-from-the-generator)
  * [Over a stream](#35-over-a-stream)
* [Core concepts](#4-core-concepts)
  * [Codecs](#41-codecs)
  * [Buffers and streams](#42-buffers-and-streams)
  * [The schema is the contract](#43-the-schema-is-the-contract)
  * [Errors and panics](#44-errors-and-panics)
  * [Sizes skipping and the exactly-N-bytes invariant](#45-sizes-skipping-and-the-exactly-n-bytes-invariant)
* [Choosing codecs](#5-choosing-codecs)

**Task guides**

* [Use cases](#6-use-cases)
  * [Numbers and enums](#61-numbers-and-enums)
  * [Strings and byte blobs](#62-strings-and-byte-blobs)
  * [Structs by hand](#63-structs-by-hand)
  * [Structs from the generator](#64-structs-from-the-generator)
  * [Collections](#65-collections)
  * [Optional fields and nil](#66-optional-fields-and-nil)
  * [Shared pointers and identity](#67-shared-pointers-and-identity)
  * [Cycles linked lists trees graphs](#68-cycles-linked-lists-trees-graphs)
  * [Polymorphism with interfaces](#69-polymorphism-with-interfaces)
  * [Versioning and migration](#610-versioning-and-migration)
  * [Streaming protocols](#611-streaming-protocols)
  * [Validating untrusted input](#612-validating-untrusted-input)
  * [Partial decoding and out-of-order reads](#613-partial-decoding-and-out-of-order-reads)
  * [Zero-copy hot paths](#614-zero-copy-hot-paths)
  * [Custom codecs](#615-custom-codecs)
  * [Composition recursion and generics](#616-composition-recursion-and-generics)
  * [Zero-allocation encoding](#617-zero-allocation-encoding)
  * [Concurrency](#618-concurrency)
  * [Testing your codecs](#619-testing-your-codecs)
  * [Interoperability](#620-interoperability)
* [Code generation](#7-code-generation)
  * [Getting started](#71-getting-started)
  * [Generator configuration](#72-generator-configuration)
  * [Generator methods](#73-generator-methods)
  * [Generated code and naming](#74-generated-code-and-naming)
  * [Serialization options](#75-serialization-options)
  * [Interfaces](#76-interfaces)
  * [Versioned types](#77-versioned-types)
  * [Multi-package and cross-package generation](#78-multi-package-and-cross-package-generation)
  * [Streaming mode](#79-streaming-mode)
  * [Workflow and troubleshooting](#710-workflow-and-troubleshooting)

**Reference**

* [Package guide](#8-package-guide)
  * [binco](#81-binco)
  * [varnum](#82-varnum)
  * [scalar](#83-scalar)
  * [composite](#84-composite)
  * [zerocopy](#85-zerocopy)
  * [ptrmap](#86-ptrmap)
  * [typeid](#87-typeid)
  * [config](#88-config)
  * [wirefmt](#89-wirefmt)
* [Cheat sheet](#9-cheat-sheet)
* [Error reference](#10-error-reference)
* [Correctness guarantees and security](#11-correctness-guarantees-and-security)
* [Performance guide](#12-performance-guide)
* [The wire format](#13-the-wire-format)
  * [Anatomy of a payload](#131-anatomy-of-a-payload)
  * [Booleans and bytes](#132-booleans-and-bytes)
  * [Varint](#133-varint)
  * [Fixed width](#134-fixed-width)
  * [Strings and byte slices](#135-strings-and-byte-slices)
  * [Slices and arrays](#136-slices-and-arrays)
  * [Maps](#137-maps)
  * [Pointers](#138-pointers)
  * [Time](#139-time)
  * [Type identifiers](#1310-type-identifiers)
  * [A payload dissected](#1311-a-payload-dissected)
  * [Compatibility rules](#1312-compatibility-rules)

**Operating it**

* [FAQ](#14-faq)
* [Working on binco](#15-working-on-binco)

## 1. Why binco

### Small payloads by construction

There is no per-field metadata, so the cost of a payload is the cost of its
values. Numbers use a variable-length encoding when that saves space, a
fixed-width encoding when speed matters more, and you pick per field. The same
struct above occupies five bytes, of which four carry actual information.

### Fast paths, chosen at compile time

Every encoding decision is static: which codec, which width, which length
prefix. The generator inlines the whole structure of a type, and the `zerocopy`
package avoids copies for strings and byte slices while keeping scalar work
branch-free. Nothing in the hot path uses reflection, interfaces you did not ask
for, or map lookups.

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
  [Compatibility rules](#1312-compatibility-rules)).
* No framing: a payload has no length prefix or delimiter. If you need framing,
  add it yourself or use a type identifier, as shown in
  [Streaming protocols](#611-streaming-protocols).
* No reflection at run time: reflection exists only in the code generator.
* No canonical map order by default: a Go map iterates in a random order, so
  encoding one is not deterministic. `composite.NewSortedMapSerializer` writes
  the entries sorted by their encoded key bytes when the bytes must be
  reproducible; see [Collections](#65-collections).


## 2. Install and requirements

```bash
go get github.com/malivvan/binco@latest
```

v1.0.0 is the first stable release. The module follows
[semantic versioning](https://semver.org/spec/v2.0.0.html), and the wire format
described in [The wire format](#13-the-wire-format) is frozen for the v1 series.
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


## 3. Quick start

### 3.1 A number

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

### 3.2 A string

```go
import "github.com/malivvan/binco/composite"

bs := binco.Encode(composite.String, "hello world")   // [11 104 101 ... 100]
text, err := binco.Decode(composite.String, bs)        // "hello world"
```

`binco.Encode` allocates the exact buffer and returns it; `binco.Decode` returns
the value and the error. Use them when convenience matters more than buffer
reuse.

### 3.3 A struct by hand

A codec for a struct is the concatenation of its fields' codecs, in declaration
order. For

```go
type Person struct {
	Name string
	Age  int
	Tags []string
}
```

the core of the codec is:

```go
var tagsSerializer = composite.NewSliceSerializer(composite.String)

func (PersonCodec) Marshal(v Person, bs []byte) (n int) {
	n = composite.String.Marshal(v.Name, bs)
	n += varnum.Int.Marshal(v.Age, bs[n:])
	return n + tagsSerializer.Marshal(v.Tags, bs[n:])
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
	v.Tags, n1, err = tagsSerializer.Unmarshal(bs[n:])
	n += n1
	return
}
```

`Size` sums the field sizes, and `Skip` advances exactly as `Unmarshal` reads.
The pattern is always the same: encode field by field into `bs[n:]`, decode
field by field from `bs[n:]`, and accumulate `n`. The complete program, with all
four methods and a runnable `main`, is in
[Structs by hand](#63-structs-by-hand).

### 3.4 The same struct from the generator

Writing that by hand does not scale. Add a directive to the file that declares
the type:

```go
//go:generate go run ./gen
```

and a small generator program:

```go
// gen/main.go
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
```

`go generate` now produces `PersonBinco`, a `binco.Serializer[Person]` with the
same behaviour as the hand-written codec. The complete program - types,
generator, generated file and the code that uses it - is in
[Getting started](#71-getting-started); the options are in
[Generator configuration](#72-generator-configuration).

### 3.5 Over a stream

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

**Where to go next**

* [Core concepts](#4-core-concepts) - the four invariants every codec keeps and
  the error/panic contract.
* [Choosing codecs](#5-choosing-codecs) - varint, fixed width or zero-copy.
* [Use cases](#6-use-cases) - a recipe for each kind of data.
* [Code generation](#7-code-generation) - let the generator write the codecs.
* [Cheat sheet](#9-cheat-sheet) - the whole API on one screen.


## 4. Core concepts

### 4.1 Codecs

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
  `binco.ErrBufferTooSmall` when `bs` is too small - a caller bug, not a data
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

### 4.2 Buffers and streams

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

### 4.3 The schema is the contract

A payload is the concatenation of the encoded fields of a type, in declaration
order. Nothing marks where a field ends: the receiver's code knows. Three
consequences:

1. **Field order is part of the format.** Reordering struct fields changes the
   wire format. Add new versions instead (see
   [Versioning and migration](#610-versioning-and-migration)).
2. **Both sides must agree on every field's codec.** A `varnum.Int` field read
   by a `scalar.Int` parser produces garbage, not an error.
3. **You can drop data by not reading it.** Unknown trailing fields are not an
   error; a reader that knows fewer fields simply stops early.

The generator writes a `SchemaFingerprint` constant next to every struct: a
SHA-256 of the field names and types in declaration order. Comparing it with the
value that produced a payload turns a silent misread (two `int` fields swapped)
into a clear failure. See
[Compatibility rules](#1312-compatibility-rules).

### 4.4 Errors and panics

`Marshal` panics with `binco.ErrBufferTooSmall` when the output buffer is too
small; that is the only panic in normal operation and it signals a bug in the
caller (`binco.Encode` sizes the buffer for you and never triggers it). Every
codec panics with that same sentinel rather than with a raw index error, so one
`recover` can recognise the case. `Unmarshal` and `Skip` never panic on malformed
input: they return one of the errors in the
[error reference](#10-error-reference), all prefixed with `binco: ` and all
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

### 4.5 Sizes skipping and the exactly-N-bytes invariant

Three guarantees make zero-allocation encoding safe and let you index into a
payload without decoding it:

* `Size(v)` equals the number of bytes `Marshal(v, bs)` writes.
* `Unmarshal` consumes exactly `Size` bytes of the value it returns.
* `Skip` advances exactly as far as `Unmarshal` would.

The test suite asserts these invariants for every codec, including the
generated ones.


## 5. Choosing codecs

Three families encode numbers, and they are interchangeable byte-wise only
within a family:

| Package | Encoding | Best for |
|---------|----------|----------|
| `varnum` | LEB128, ZigZag for signed | small and medium numbers, ids, counters |
| `scalar` | fixed width, little-endian | large numbers, random-access data, speed |
| `zerocopy` | fixed width, same bytes as `scalar`; zero-copy strings and blobs | hot paths, aliasing views |

`zerocopy` is byte-compatible with `scalar` for numbers and with `composite` for
strings and byte slices, so switching codecs never changes the format, only the
code.

**Points to weigh**

* Space: varint wins below the thresholds in [Fixed width](#134-fixed-width).
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

## 6. Use cases

Each recipe shows a working shape: the helper types and codecs it uses are
declared in the same recipe or in a neighbouring one, and the snippets compile
as written once those declarations exist.

### 6.1 Numbers and enums

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

### 6.2 Strings and byte blobs

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

Read [Zero-copy hot paths](#614-zero-copy-hot-paths) before using it.

### 6.3 Structs by hand

A struct is the concatenation of its fields in declaration order, so a struct
codec is the concatenation of its field codecs. `Size` is exact, `Skip` advances
in step with `Unmarshal`, and the same four methods work in the buffer and the
streaming API.

The complete program below defines a type, implements `binco.Serializer[Foo]`
by hand, gives `Foo` the `MarshalBinco`/`SizeBinco` methods (so `binco.Marshal`
works), and prints the bytes at every step.

**`types/struct/foo.go`**

```go
package main

type Foo struct {
	Str string
	Sl  []int
}
```

**`types/struct/serializer.go`**

```go
// FooSerializer is a hand-written codec; the generated code has the same shape.
package main

import (
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

var (
	FooSerializer  = fooSerializer{}
	intsSerializer = composite.NewSliceSerializer(varnum.Int)
)

type fooSerializer struct{}

func (s fooSerializer) Marshal(v Foo, bs []byte) (n int) {

	n = composite.String.Marshal(v.Str, bs)

	return n + intsSerializer.Marshal(v.Sl, bs[n:])
}

func (s fooSerializer) Unmarshal(bs []byte) (v Foo, n int, err error) {

	v.Str, n, err = composite.String.Unmarshal(bs)
	if err != nil {
		return
	}

	var n1 int
	v.Sl, n1, err = intsSerializer.Unmarshal(bs[n:])

	n += n1
	return
}

func (s fooSerializer) Size(v Foo) (size int) {
	return composite.String.Size(v.Str) + intsSerializer.Size(v.Sl)
}

func (s fooSerializer) Skip(bs []byte) (n int, err error) {
	n, err = composite.String.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = intsSerializer.Skip(bs[n:])
	return n + n1, err
}

func (f Foo) MarshalBinco(bs []byte) (n int) {
	return FooSerializer.Marshal(f, bs)
}

func (f Foo) SizeBinco() (size int) {
	return FooSerializer.Size(f)
}
```

**`types/struct/main.go`**

```go
// Command struct uses the hand-written FooSerializer serializer to show the shape of an encoded struct.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {
	v := Foo{
		Str: "hello, binco",
		Sl:  []int{1, 2, 3},
	}

	bs := make([]byte, FooSerializer.Size(v))
	n := FooSerializer.Marshal(v, bs)
	fmt.Printf("value:   %+v\n", v)
	fmt.Printf("size:    %d bytes\n", len(bs))
	fmt.Printf("encoded: %v (wrote %d)\n", bs, n)

	back, n, err := FooSerializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %+v (read %d)\n", back, n)

	n, err = FooSerializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)

	bs = binco.Marshal(v)
	fmt.Printf("binco.Marshal: %v\n", bs)

	if _, _, err = FooSerializer.Unmarshal(bs[:len(bs)-1]); err != nil {
		fmt.Printf("truncated payload: %v\n", err)
	}
}
```

Output:

```text
value:   {Str:hello, binco Sl:[1 2 3]}
size:    17 bytes
encoded: [12 104 101 108 108 111 44 32 98 105 110 99 111 3 2 4 6] (wrote 17)
decoded: {Str:hello, binco Sl:[1 2 3]} (read 17)
skipped: 17 bytes
binco.Marshal: [12 104 101 108 108 111 44 32 98 105 110 99 111 3 2 4 6]
truncated payload: binco: truncated input
```

**Trailing fields.** Appending a field changes the payload, and old payloads do
not contain it. Either version the type (see
[Versioning and migration](#610-versioning-and-migration)), or let the reader
detect the short tail before reading the new field:

```go
if n == len(bs) {
	return v, n, nil // a payload written before the field existed
}
v.Extra, n1, err = extraSerializer.Unmarshal(bs[n:])
n += n1
```

**Propagate errors.** `Unmarshal` stops at the first error, and `Skip` must do
the same, or the failure surfaces far from its cause.

### 6.4 Structs from the generator

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
automatically. Read [Code generation](#7-code-generation) for the options and
the workflow.

### 6.5 Collections

Slices and arrays are a length followed by the elements; maps are a length
followed by key/value pairs. An array is encoded exactly like a slice, but its
type fixes the length: a payload with a different element count is rejected with
`binco.ErrWrongLength`. Map entry order is not deterministic on encode; decoding
is.

The three programs below are complete `main` packages.

**`types/array/main.go`**

```go
// Command array encodes an array, decodes it, skips it and shows that a payload whose length differs from the array type is rejected.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/zerocopy"
)

func main() {
	var (
		value      = [3]int{1, 2, 3}
		serializer = zerocopy.NewArraySerializer[[3]int, int](varnum.Int)
	)

	size := serializer.Size(value)
	bs := make([]byte, size)
	n := serializer.Marshal(value, bs)

	fmt.Printf("value:   %v\n", value)
	fmt.Printf("size:    %d bytes\n", size)
	fmt.Printf("encoded: %v (wrote %d)\n", bs, n)

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %v (read %d)\n", back, n)

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)

	shorter := binco.Encode(composite.NewSliceSerializer(varnum.Int), []int{1, 2})
	if _, _, err = serializer.Unmarshal(shorter); err != nil {
		fmt.Printf("shorter payload: %v\n", err)
	}
}
```

Output:

```text
value:   [1 2 3]
size:    4 bytes
encoded: [3 2 4 6] (wrote 4)
decoded: [1 2 3] (read 4)
skipped: 4 bytes
shorter payload: binco: wrong length
```

**`types/slice/main.go`**

```go
// Command slice encodes a slice with the default length codec and with a fixed-width one, and shows how nil and empty slices behave.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	slopts "github.com/malivvan/binco/config/slice"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

func main() {
	var (
		value = []int{1, 2, 3}

		serializer = composite.NewSliceSerializer(varnum.Int)

		rawLenSerializer = composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](scalar.Int))
	)

	bs := binco.Encode(serializer, value)
	fmt.Printf("value:     %v\n", value)
	fmt.Printf("encoded:   %v (%d bytes)\n", bs, len(bs))

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded:   %v (%d bytes)\n", back, n)

	bs = binco.Encode(rawLenSerializer, value)
	fmt.Printf("raw len:   %v (%d bytes)\n", bs, len(bs))

	for _, v := range [][]int{nil, {}} {
		back, err = binco.Decode(serializer, binco.Encode(serializer, v))
		if err != nil {
			panic(err)
		}
		fmt.Printf("nil=%v → len=%d, nil=%v\n", v == nil, len(back), back == nil)
	}

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped:   %d bytes\n", n)
}
```

Output:

```text
value:     [1 2 3]
encoded:   [3 2 4 6] (4 bytes)
decoded:   [1 2 3] (4 bytes)
raw len:   [3 0 0 0 0 0 0 0 2 4 6] (11 bytes)
nil=true → len=0, nil=false
nil=false → len=0, nil=false
skipped:   4 bytes
```

**`types/map/main.go`**

```go
// Command map encodes a map and decodes it back. Entry order is Go's map order, so only decoding is deterministic.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

func main() {
	var (
		value = map[int]string{1: "one", 2: "two", 3: "three"}

		serializer = composite.NewMapSerializer(varnum.Int, composite.String)
	)

	bs := make([]byte, serializer.Size(value))
	n := serializer.Marshal(value, bs)
	fmt.Printf("value:    %v\n", sorted(value))
	fmt.Printf("size:     %d bytes\n", len(bs))
	fmt.Printf("encoded:  %v (wrote %d)\n", bs, n)

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded:  %v (read %d)\n", sorted(back), n)

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped:  %d bytes\n", n)

	empty, err := binco.Decode(serializer, binco.Encode(serializer, map[int]string(nil)))
	if err != nil {
		panic(err)
	}
	fmt.Printf("empty:    len=%d, nil=%v\n", len(empty), empty == nil)
}

func sorted(m map[int]string) string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%d:%q", k, m[k]))
	}
	return "{" + strings.Join(out, ", ") + "}"
}
```

Output:

```text
value:    {1:"one", 2:"two", 3:"three"}
size:     18 bytes
encoded:  [3 2 3 111 110 101 4 3 116 119 111 6 5 116 104 114 101 101] (wrote 18)
decoded:  {1:"one", 2:"two", 3:"three"} (read 18)
skipped:  18 bytes
empty:    len=0, nil=false
```

**Length codecs.** The default length is `varnum.PositiveInt`; replace it when a
fixed-size record matters:

```go
seekable := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](scalar.Int))
```

**Canonical maps.** Use the sorted map codec when the bytes must be reproducible
(signatures, content addressing, golden files):

```go
canonical := composite.NewSortedMapSerializer(composite.String, varnum.Int)
```

**Sets.** A set is a map with an empty value, or a sorted slice of keys.

### 6.6 Optional fields and nil

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

### 6.7 Shared pointers and identity

A plain pointer codec writes every pointer by value: two fields pointing at
the same value store it twice, and a cycle cannot be encoded at all. Pointer
mapping (`ptrmap`) assigns an id to each pointer and writes the pointed-to data
only the first time; later occurrences are the mapping flag plus the id.

The complete program below stores one string in two fields and proves that the
identity survives the round trip.

**`pm/simple/two_ptrs.go`**

```go
package main

type TwoPtrs struct {
	Ptr1 *string
	Ptr2 *string
}
```

**`pm/simple/embedded_ptrs.go`**

```go
package main

type EmbeddedPtrs struct {
	TwoPtrs
	Ptr3 *string
}
```

**`pm/simple/serializers.go`**

```go
package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
)

func MakeTwoPtrsSerializer() binco.Serializer[TwoPtrs] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newTwoPtrsSerializer(ptrMap, revPtrMap))
}

func MakeEmbeddedPtrsSerializer() binco.Serializer[EmbeddedPtrs] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newEmbeddedPtrsSerializer(ptrMap, revPtrMap))
}

func newTwoPtrsSerializer(ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap) twoPtrsSerializer {
	return twoPtrsSerializer{strPtrSerializer: ptrmap.NewPtrSerializer(ptrMap, revPtrMap, composite.String)}
}

func newEmbeddedPtrsSerializer(ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap) embeddedPtrsSerializer {
	return embeddedPtrsSerializer{

		twoSerializer:  newTwoPtrsSerializer(ptrMap, revPtrMap),
		ptr3Serializer: ptrmap.NewPtrSerializer(ptrMap, revPtrMap, composite.String),
	}
}

type twoPtrsSerializer struct {
	strPtrSerializer binco.Serializer[*string]
}

func (s twoPtrsSerializer) Marshal(v TwoPtrs, bs []byte) (n int) {
	n = s.strPtrSerializer.Marshal(v.Ptr1, bs)
	return n + s.strPtrSerializer.Marshal(v.Ptr2, bs[n:])
}

func (s twoPtrsSerializer) Unmarshal(bs []byte) (v TwoPtrs, n int, err error) {
	v.Ptr1, n, err = s.strPtrSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Ptr2, n1, err = s.strPtrSerializer.Unmarshal(bs[n:])
	n += n1
	return
}

func (s twoPtrsSerializer) Size(v TwoPtrs) (size int) {
	return s.strPtrSerializer.Size(v.Ptr1) + s.strPtrSerializer.Size(v.Ptr2)
}

func (s twoPtrsSerializer) Skip(bs []byte) (n int, err error) {
	n, err = s.strPtrSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.strPtrSerializer.Skip(bs[n:])
	return n + n1, err
}

type embeddedPtrsSerializer struct {
	twoSerializer  twoPtrsSerializer
	ptr3Serializer binco.Serializer[*string]
}

func (s embeddedPtrsSerializer) Marshal(v EmbeddedPtrs, bs []byte) (n int) {
	n = s.twoSerializer.Marshal(v.TwoPtrs, bs)
	return n + s.ptr3Serializer.Marshal(v.Ptr3, bs[n:])
}

func (s embeddedPtrsSerializer) Unmarshal(bs []byte) (v EmbeddedPtrs, n int, err error) {
	v.TwoPtrs, n, err = s.twoSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Ptr3, n1, err = s.ptr3Serializer.Unmarshal(bs[n:])
	n += n1
	return
}

func (s embeddedPtrsSerializer) Size(v EmbeddedPtrs) (size int) {
	return s.twoSerializer.Size(v.TwoPtrs) + s.ptr3Serializer.Size(v.Ptr3)
}

func (s embeddedPtrsSerializer) Skip(bs []byte) (n int, err error) {
	n, err = s.twoSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.ptr3Serializer.Skip(bs[n:])
	return n + n1, err
}
```

**`pm/simple/main.go`**

```go
// Command simple shows how pointer mapping keeps one pointer shared by several fields.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	two := NewTwoPtrs("the same pointer in two fields")
	bs := binco.Encode(MakeTwoPtrsSerializer(), two)
	fmt.Printf("TwoPtrs: %v\n", bs)
	fmt.Printf("  [2 0 ...] flag 2 (mapping), ID 0, then the string\n")
	fmt.Printf("  [2 0]     the same ID again, no string\n")

	back := decodeTwoPtrs(bs)
	fmt.Printf("  decoded: %q, %q\n", *back.Ptr1, *back.Ptr2)
	fmt.Printf("  identity preserved: %v\n\n", back.Ptr1 == back.Ptr2)

	embedded := NewEmbeddedPtrs("the same pointer in a composite struct")
	bs = binco.Encode(MakeEmbeddedPtrsSerializer(), embedded)
	fmt.Printf("EmbeddedPtrs: %v (%d bytes)\n", bs, len(bs))

	backE := decodeEmbeddedPtrs(bs)
	fmt.Printf("  decoded: %q, %q, %q\n", *backE.Ptr1, *backE.Ptr2, *backE.Ptr3)
	fmt.Printf("  identity preserved: %v (all three are the same pointer)\n",
		backE.Ptr1 == backE.Ptr2 && backE.Ptr1 == backE.Ptr3)

	var nilPtrs TwoPtrs
	bs = binco.Encode(MakeTwoPtrsSerializer(), nilPtrs)
	fmt.Printf("\nnil pointers: %v\n", bs)
	back = decodeTwoPtrs(bs)
	fmt.Printf("  decoded nil: %v, %v\n", back.Ptr1 == nil, back.Ptr2 == nil)
}

func NewTwoPtrs(s string) TwoPtrs {
	ptr := &s
	return TwoPtrs{Ptr1: ptr, Ptr2: ptr}
}

func NewEmbeddedPtrs(s string) EmbeddedPtrs {
	ptr := &s
	return EmbeddedPtrs{
		Ptr1: ptr,
		Ptr2: ptr,
		Ptr3: ptr,
	}
}

func decodeTwoPtrs(bs []byte) TwoPtrs {
	v, err := binco.Decode(MakeTwoPtrsSerializer(), bs)
	if err != nil {
		panic(err)
	}
	return v
}

func decodeEmbeddedPtrs(bs []byte) EmbeddedPtrs {
	v, err := binco.Decode(MakeEmbeddedPtrsSerializer(), bs)
	if err != nil {
		panic(err)
	}
	return v
}
```

Output:

```text
TwoPtrs: [2 0 30 116 104 101 32 115 97 109 101 32 112 111 105 110 116 101 114 32 105 110 32 116 119 111 32 102 105 101 108 100 115 2 0]
  [2 0 ...] flag 2 (mapping), ID 0, then the string
  [2 0]     the same ID again, no string
  decoded: "the same pointer in two fields", "the same pointer in two fields"
  identity preserved: true

EmbeddedPtrs: [2 0 38 116 104 101 32 115 97 109 101 32 112 111 105 110 116 101 114 32 105 110 32 97 32 99 111 109 112 111 115 105 116 101 32 115 116 114 117 99 116 2 0 2 0] (45 bytes)
  decoded: "the same pointer in a composite struct", "the same pointer in a composite struct", "the same pointer in a composite struct"
  identity preserved: true (all three are the same pointer)

nil pointers: [1 1]
  decoded nil: true, true
```

The wiring rules:

1. Every `ptrmap` codec of one value shares the same `PtrMap` and
   `ReversePtrMap`, or the ids will not match.
2. Wrap the outermost codec with `ptrmap.Wrap`, which clears both maps around
   every operation (including `Size`, which assigns ids). A wrapped codec is
   stateful: give each goroutine its own maps.
3. Ids start at zero in encounter order and remember the type they were defined
   as; a mismatched reference is `binco.ErrWrongType`, never a reinterpreted
   pointer.

**When not to use it.** For a tree without shared references, the plain
`composite.NewPtrSerializer` is smaller and faster: no ids, no maps, no
bookkeeping per pointer.

### 6.8 Cycles linked lists trees graphs

A cycle is a pointer graph where following pointers returns to a node. The
plain pointer codec cannot represent it; `ptrmap` can, because the second visit
is written as an id. Each element is written once, and the back references become
ids.

**A doubly linked list.** Only the head, the tail and the length are data; the
elements are reached through pointers. The program also walks the list backwards
after decoding, which only works if `next.Prev == current` everywhere.

**`pm/linked_list/elem.go`**

```go
package main

type Elem[T any] struct {
	Val  T
	Prev *Elem[T]
	Next *Elem[T]
}
```

**`pm/linked_list/linked_list.go`**

```go
package main

import (
	"strconv"
	"strings"
)

type LinkedList[T any] struct {
	Head *Elem[T]
	Tail *Elem[T]
	Len  int
}

func (l *LinkedList[T]) AddFront(v T) {
	elem := &Elem[T]{Val: v}
	if l.Head == nil {
		l.Tail = elem
	} else {
		elem.Next = l.Head
		l.Head.Prev = elem
	}
	l.Head = elem
	l.Len++
}

func (l *LinkedList[T]) AddBack(v T) {
	elem := &Elem[T]{Val: v}
	if l.Tail == nil {
		l.Head = elem
	} else {
		elem.Prev = l.Tail
		l.Tail.Next = elem
	}
	l.Tail = elem
	l.Len++
}

func (l *LinkedList[T]) Remove(e *Elem[T]) {
	switch {
	case e == l.Head && e == l.Tail:
		l.Head, l.Tail = nil, nil
	case e == l.Head:
		l.Head = e.Next
		l.Head.Prev = nil
	case e == l.Tail:
		l.Tail = e.Prev
		l.Tail.Next = nil
	default:
		e.Prev.Next = e.Next
		e.Next.Prev = e.Prev
	}
	l.Len--
}

func (l LinkedList[T]) Forward() (values []T) {
	for e := l.Head; e != nil; e = e.Next {
		values = append(values, e.Val)
	}
	return
}

func (l LinkedList[T]) Backward() (values []T) {
	for e := l.Tail; e != nil; e = e.Prev {
		values = append(values, e.Val)
	}
	return
}

func (l LinkedList[T]) String() string {
	parts := make([]string, 0, l.Len)
	for _, v := range l.Forward() {
		parts = append(parts, strconv.FormatInt(int64(toInt(v)), 10))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func toInt[T any](v T) int {
	if n, ok := any(v).(int); ok {
		return n
	}
	return 0
}
```

**`pm/linked_list/serializers.go`**

```go
package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func MakeLinkedListSerializer[T any](valSerializer binco.Serializer[T]) binco.Serializer[LinkedList[T]] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newLinkedListSerializer(ptrMap, revPtrMap, valSerializer))
}

func newLinkedListSerializer[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	valSerializer binco.Serializer[T],
) linkedListSerializer[T] {
	var (
		elem              elemSerializer[T]
		elemPtrSerializer binco.Serializer[*Elem[T]]
	)

	elem = elemSerializer[T]{valSerializer: valSerializer, elemPtrSerializer: &elemPtrSerializer}
	elemPtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, elem)

	return linkedListSerializer[T]{elemPtrSerializer: elemPtrSerializer}
}

type elemSerializer[T any] struct {
	valSerializer     binco.Serializer[T]
	elemPtrSerializer *binco.Serializer[*Elem[T]]
}

func (s elemSerializer[T]) Marshal(v Elem[T], bs []byte) (n int) {
	n = s.valSerializer.Marshal(v.Val, bs)
	n += (*s.elemPtrSerializer).Marshal(v.Prev, bs[n:])
	return n + (*s.elemPtrSerializer).Marshal(v.Next, bs[n:])
}

func (s elemSerializer[T]) Unmarshal(bs []byte) (v Elem[T], n int, err error) {
	v.Val, n, err = s.valSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Prev, n1, err = (*s.elemPtrSerializer).Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Next, n1, err = (*s.elemPtrSerializer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s elemSerializer[T]) Size(v Elem[T]) (size int) {
	return s.valSerializer.Size(v.Val) +
		(*s.elemPtrSerializer).Size(v.Prev) +
		(*s.elemPtrSerializer).Size(v.Next)
}

func (s elemSerializer[T]) Skip(bs []byte) (n int, err error) {
	n, err = s.valSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.elemPtrSerializer).Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = (*s.elemPtrSerializer).Skip(bs[n:])
	return n + n1, err
}

type linkedListSerializer[T any] struct {
	elemPtrSerializer binco.Serializer[*Elem[T]]
}

func (s linkedListSerializer[T]) Marshal(v LinkedList[T], bs []byte) (n int) {
	n = s.elemPtrSerializer.Marshal(v.Head, bs)
	n += s.elemPtrSerializer.Marshal(v.Tail, bs[n:])
	return n + varnum.PositiveInt.Marshal(v.Len, bs[n:])
}

func (s linkedListSerializer[T]) Unmarshal(bs []byte) (v LinkedList[T], n int, err error) {
	v.Head, n, err = s.elemPtrSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Tail, n1, err = s.elemPtrSerializer.Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Len, n1, err = varnum.PositiveInt.Unmarshal(bs[n:])
	n += n1
	return
}

func (s linkedListSerializer[T]) Size(v LinkedList[T]) (size int) {
	return s.elemPtrSerializer.Size(v.Head) +
		s.elemPtrSerializer.Size(v.Tail) +
		varnum.PositiveInt.Size(v.Len)
}

func (s linkedListSerializer[T]) Skip(bs []byte) (n int, err error) {
	n, err = s.elemPtrSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.elemPtrSerializer.Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = varnum.PositiveInt.Skip(bs[n:])
	return n + n1, err
}
```

**`pm/linked_list/main.go`**

```go
// Command linked_list encodes a doubly linked list; the pointers are restored, so the decoded list is fully functional.
package main

import (
	"fmt"

	"github.com/malivvan/binco/varnum"
)

func main() {
	list := ShortLinkedList()
	serializer := MakeLinkedListSerializer[int](varnum.Int)

	bs := make([]byte, serializer.Size(list))
	n := serializer.Marshal(list, bs)
	fmt.Printf("list:     %v (%d elements)\n", list, list.Len)
	fmt.Printf("size:     %d bytes\n", len(bs))
	fmt.Printf("encoded:  %v (wrote %d)\n", bs, n)

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded:  %v (read %d bytes)\n", back, n)
	fmt.Printf("forward:  %v\n", back.Forward())
	fmt.Printf("backward: %v\n", back.Backward())

	ok := true
	for e := back.Head; e != nil && e.Next != nil; e = e.Next {
		if e.Next.Prev != e {
			ok = false
		}
	}
	fmt.Printf("prev/next links consistent: %v\n", ok)
	fmt.Printf("Head.Prev == nil: %v, Tail.Next == nil: %v\n",
		back.Head.Prev == nil, back.Tail.Next == nil)

	second := back.Head.Next
	back.Remove(second)
	fmt.Printf("\nafter removing %v: %v (len=%d)\n", second.Val, back, back.Len)
	fmt.Printf("new head.Next.Prev == head: %v\n", back.Head.Next.Prev == back.Head)

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped:  %d bytes\n", n)
}

func ShortLinkedList() LinkedList[int] {
	var l LinkedList[int]
	for _, v := range []int{8, 9, 10, 11} {
		l.AddBack(v)
	}
	return l
}
```

Output:

```text
list:     [8, 9, 10, 11] (4 elements)
size:     23 bytes
encoded:  [2 0 16 1 2 1 18 2 0 2 2 20 2 1 2 3 22 2 2 1 2 3 4] (wrote 23)
decoded:  [8, 9, 10, 11] (read 23 bytes)
forward:  [8 9 10 11]
backward: [11 10 9 8]
prev/next links consistent: true
Head.Prev == nil: true, Tail.Next == nil: true

after removing 9: [8, 10, 11] (len=3)
new head.Next.Prev == head: true
skipped:  23 bytes
```

**A graph.** A vertex points at its edges and every edge points back at a
vertex, so the serializer graph itself is cyclic. The serializers are wired
through variables that are assigned at the end of the constructor.

**`pm/graph/vertex.go`**

```go
package main

type Vertex[T comparable, V any] struct {
	Val   V
	Edges map[T]*Edge[T, V]
}
```

**`pm/graph/edge.go`**

```go
package main

type Edge[T comparable, V any] struct {
	Weight int
	Vertex *Vertex[T, V]
}
```

**`pm/graph/graph.go`**

```go
package main

import (
	"fmt"
	"slices"
	"strings"
)

func NewGraph[T comparable, V any]() Graph[T, V] {
	return Graph[T, V]{Vertices: make(map[T]*Vertex[T, V])}
}

type Graph[T comparable, V any] struct {
	Vertices map[T]*Vertex[T, V]
}

func (g Graph[T, V]) AddVertex(id T, val V) {
	g.Vertices[id] = &Vertex[T, V]{Val: val}
}

func (g Graph[T, V]) AddEdge(from, to T, weight int) error {
	fromVertex, ok := g.Vertices[from]
	if !ok {
		return fmt.Errorf("vertex %v not found", from)
	}
	toVertex, ok := g.Vertices[to]
	if !ok {
		return fmt.Errorf("vertex %v not found", to)
	}
	if fromVertex.Edges == nil {
		fromVertex.Edges = make(map[T]*Edge[T, V])
	}
	fromVertex.Edges[to] = &Edge[T, V]{Vertex: toVertex, Weight: weight}
	return nil
}

func (g Graph[T, V]) Render() string {
	var b strings.Builder
	for _, id := range sortedKeys(g.Vertices) {
		vertex := g.Vertices[id]
		fmt.Fprintf(&b, "  vertex %v: %v", id, vertex.Val)
		if len(vertex.Edges) > 0 {
			b.WriteString(" → ")
			edges := make([]string, 0, len(vertex.Edges))
			for _, to := range sortedKeys(vertex.Edges) {
				edges = append(edges, fmt.Sprintf("%v (weight %d)", to, vertex.Edges[to].Weight))
			}
			b.WriteString(strings.Join(edges, ", "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func sortedKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b K) int {
		return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
	})
	return keys
}
```

**`pm/graph/serializers.go`**

```go
package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func MakeGraphSerializer[T comparable, V any](keySerializer binco.Serializer[T],
	valSerializer binco.Serializer[V],
) binco.Serializer[Graph[T, V]] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newGraphSerializer(ptrMap, revPtrMap, keySerializer, valSerializer))
}

func newGraphSerializer[T comparable, V any](ptrMap *binco.PtrMap,
	revPtrMap *binco.ReversePtrMap, keySerializer binco.Serializer[T], valSerializer binco.Serializer[V],
) graphSerializer[T, V] {
	var (
		edgesSerializer    binco.Serializer[map[T]*Edge[T, V]]
		verticesSerializer binco.Serializer[map[T]*Vertex[T, V]]

		edgePtrSerializer   binco.Serializer[*Edge[T, V]]
		vertexPtrSerializer binco.Serializer[*Vertex[T, V]]

		edge   = edgeSerializer[T, V]{vertexPtrSerializer: &vertexPtrSerializer}
		vertex = vertexSerializer[T, V]{valSerializer: valSerializer, edgesSerializer: &edgesSerializer}
	)
	edgePtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, edge)
	vertexPtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, vertex)

	edgesSerializer = composite.NewMapSerializer(keySerializer, edgePtrSerializer)
	verticesSerializer = composite.NewMapSerializer(keySerializer, vertexPtrSerializer)

	return graphSerializer[T, V]{verticesSerializer: verticesSerializer}
}

type edgeSerializer[T comparable, V any] struct {
	vertexPtrSerializer *binco.Serializer[*Vertex[T, V]]
}

func (s edgeSerializer[T, V]) Marshal(v Edge[T, V], bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Weight, bs)
	return n + (*s.vertexPtrSerializer).Marshal(v.Vertex, bs[n:])
}

func (s edgeSerializer[T, V]) Unmarshal(bs []byte) (v Edge[T, V], n int, err error) {
	v.Weight, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Vertex, n1, err = (*s.vertexPtrSerializer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s edgeSerializer[T, V]) Size(v Edge[T, V]) (size int) {
	return varnum.Int.Size(v.Weight) + (*s.vertexPtrSerializer).Size(v.Vertex)
}

func (s edgeSerializer[T, V]) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.vertexPtrSerializer).Skip(bs[n:])
	return n + n1, err
}

type vertexSerializer[T comparable, V any] struct {
	valSerializer   binco.Serializer[V]
	edgesSerializer *binco.Serializer[map[T]*Edge[T, V]]
}

func (s vertexSerializer[T, V]) Marshal(v Vertex[T, V], bs []byte) (n int) {
	n = s.valSerializer.Marshal(v.Val, bs)
	return n + (*s.edgesSerializer).Marshal(v.Edges, bs[n:])
}

func (s vertexSerializer[T, V]) Unmarshal(bs []byte) (v Vertex[T, V], n int, err error) {
	v.Val, n, err = s.valSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Edges, n1, err = (*s.edgesSerializer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s vertexSerializer[T, V]) Size(v Vertex[T, V]) (size int) {
	return s.valSerializer.Size(v.Val) + (*s.edgesSerializer).Size(v.Edges)
}

func (s vertexSerializer[T, V]) Skip(bs []byte) (n int, err error) {
	n, err = s.valSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.edgesSerializer).Skip(bs[n:])
	return n + n1, err
}

type graphSerializer[T comparable, V any] struct {
	verticesSerializer binco.Serializer[map[T]*Vertex[T, V]]
}

func (s graphSerializer[T, V]) Marshal(v Graph[T, V], bs []byte) (n int) {
	return s.verticesSerializer.Marshal(v.Vertices, bs)
}

func (s graphSerializer[T, V]) Unmarshal(bs []byte) (v Graph[T, V], n int, err error) {
	v = Graph[T, V]{}
	v.Vertices, n, err = s.verticesSerializer.Unmarshal(bs)
	return
}

func (s graphSerializer[T, V]) Size(v Graph[T, V]) (size int) {
	return s.verticesSerializer.Size(v.Vertices)
}

func (s graphSerializer[T, V]) Skip(bs []byte) (n int, err error) {
	return s.verticesSerializer.Skip(bs)
}
```

**`pm/graph/main.go`**

```go
// Command graph encodes a cyclic graph. Edges point back at vertices, and the pointer identities survive the round trip.
package main

import (
	"fmt"

	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

func main() {
	graph := CyclicGraph()
	serializer := MakeGraphSerializer[int, string](varnum.Int, composite.String)

	bs := make([]byte, serializer.Size(graph))
	n := serializer.Marshal(graph, bs)
	fmt.Printf("graph:\n%s", graph.Render())
	fmt.Printf("size:    %d bytes\n", len(bs))
	fmt.Printf("encoded: %v (wrote %d)\n\n", bs, n)

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded %d bytes:\n%s\n", n, back.Render())

	v1, v2, v3 := back.Vertices[1], back.Vertices[2], back.Vertices[3]
	fmt.Printf("edge 1→2 → vertex 2: %v\n", v1.Edges[2].Vertex == v2)
	fmt.Printf("edge 2→3 → vertex 3: %v\n", v2.Edges[3].Vertex == v3)
	fmt.Printf("edge 3→1 → vertex 1: %v\n", v3.Edges[1].Vertex == v1)

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)
}

func CyclicGraph() Graph[int, string] {
	g := NewGraph[int, string]()
	g.AddVertex(1, "one")
	g.AddVertex(2, "two")
	g.AddVertex(3, "three")

	edges := [][3]int{{1, 2, 10}, {2, 3, 20}, {3, 1, 30}}
	for _, e := range edges {
		if err := g.AddEdge(e[0], e[1], e[2]); err != nil {
			panic(err)
		}
	}
	return g
}
```

Output:

```text
graph:
  vertex 1: one → 2 (weight 10)
  vertex 2: two → 3 (weight 20)
  vertex 3: three → 1 (weight 30)
size:    45 bytes
encoded: [3 2 2 0 3 111 110 101 1 4 2 1 20 2 2 3 116 119 111 1 6 2 3 40 2 4 5 116 104 114 101 101 1 2 2 5 60 2 0 4 2 2 6 2 4] (wrote 45)

decoded 45 bytes:
  vertex 1: one → 2 (weight 10)
  vertex 2: two → 3 (weight 20)
  vertex 3: three → 1 (weight 30)

edge 1→2 → vertex 2: true
edge 2→3 → vertex 3: true
edge 3→1 → vertex 1: true
skipped: 45 bytes
```

**Trees.** A tree with parent pointers has the same shape: use
`ptrmap.NewPtrSerializer` for the parent and child pointers, and leave the root's
parent nil.

**Depth.** `ptrmap` writes data the first time it meets a pointer, so a value is
written and read once. Recursion is bounded by `binco.MaxDepth`; a
pathologically deep graph reports `binco.ErrTooDeep` instead of exhausting the
stack.

### 6.9 Polymorphism with interfaces

An interface field can hold several concrete types, so the payload must say
which one. Each implementation gets a DTM and the interface codec dispatches on
it.

**Marshalling without a type switch.** If every implementation satisfies
`binco.MarshallerTyped`, the interface codec delegates to
`MarshalTypedBinco`; adding an implementation then needs no change to the
interface codec. The first program sends two unrelated types through typed
serializers.

**`typed/different/foo.go`**

```go
package main

type Foo struct {
	Num int
}
```

**`typed/different/bar.go`**

```go
package main

type Bar struct {
	Str string
}
```

**`typed/different/serializers.go`**

```go
package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const (
	FooDTM binco.DTM = iota + 1
	BarDTM
)

var (
	FooSerializer = fooSerializer{}
	BarSerializer = barSerializer{}
)

var (
	FooTypedSerializer = typeid.NewSerializer[Foo](FooDTM, FooSerializer)
	BarTypedSerializer = typeid.NewSerializer[Bar](BarDTM, BarSerializer)
)

type fooSerializer struct{}

func (fooSerializer) Marshal(v Foo, bs []byte) (n int) {
	return varnum.Int.Marshal(v.Num, bs)
}

func (fooSerializer) Unmarshal(bs []byte) (v Foo, n int, err error) {
	v.Num, n, err = varnum.Int.Unmarshal(bs)
	return
}

func (fooSerializer) Size(v Foo) (size int) {
	return varnum.Int.Size(v.Num)
}

func (fooSerializer) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

type barSerializer struct{}

func (barSerializer) Marshal(v Bar, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (barSerializer) Unmarshal(bs []byte) (v Bar, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (barSerializer) Size(v Bar) (size int) {
	return composite.String.Size(v.Str)
}

func (barSerializer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}
```

**`typed/different/main.go`**

```go
// Command different puts unrelated types behind type identifiers and dispatches on them.
package main

import (
	"errors"
	"fmt"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/typeid"
)

func main() {

	fooBS := binco.Encode(FooTypedSerializer, Foo{Num: 5})
	barBS := binco.Encode(BarTypedSerializer, Bar{Str: "hello world"})
	fmt.Printf("Foo: %v (DTM %d)\n", fooBS, FooDTM)
	fmt.Printf("Bar: %v (DTM %d)\n\n", barBS, BarDTM)

	for _, bs := range [][]byte{fooBS, barBS} {
		if err := receive(bs); err != nil {
			panic(err)
		}
	}

	_, _, err := BarTypedSerializer.Unmarshal(fooBS)
	if wrong, ok := errors.AsType[binco.WrongDTMError](err); ok {
		fmt.Printf("Bar reader got a Foo payload: expected %d, actual %d\n",
			wrong.Expected(), wrong.Actual())
	}

	dtm, n, err := typeid.DTMSerializer.Unmarshal(fooBS)
	if err != nil {
		panic(err)
	}
	fmt.Printf("DTM prefix: %d (%d byte), data: %v\n", dtm, n, fooBS[n:])
}

func receive(bs []byte) (err error) {

	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}

	switch dtm {
	case FooDTM:
		foo, _, err := FooTypedSerializer.UnmarshalData(bs[n:])
		if err != nil {
			return err
		}
		fmt.Printf("received Foo %+v\n", foo)
	case BarDTM:
		bar, _, err := BarTypedSerializer.UnmarshalData(bs[n:])
		if err != nil {
			return err
		}
		fmt.Printf("received Bar %+v\n", bar)
	default:
		return fmt.Errorf("unexpected message: %w", binco.NewUnexpectedDTMError(dtm))
	}
	return nil
}
```

Output:

```text
Foo: [1 10] (DTM 1)
Bar: [2 11 104 101 108 108 111 32 119 111 114 108 100] (DTM 2)

received Foo {Num:5}
received Bar {Str:hello world}
Bar reader got a Foo payload: expected 2, actual 1
DTM prefix: 1 (1 byte), data: [10]
```

**An interface with a dispatch table.** Unmarshalling needs the concrete type,
which is unknown until the DTM is read, so the interface codec switches on it
(or looks it up in a `map[binco.DTM]`). The second program implements an
`Instruction` interface with two concrete instructions and reports unknown ids
with `binco.NewUnexpectedDTMError`.

**`typed/oneof/instruction.go`**

```go
// The implementations carry their own marshalling, so the interface codec never names them.
package main

import "fmt"

type Instruction interface {
	Do()
}

type Copy struct {
	Start int
	End   int
}

func (c Copy) Do() {
	fmt.Printf("copy from %d to %d\n", c.Start, c.End)
}

func (c Copy) MarshalTypedBinco(bs []byte) (n int) {
	return CopyTypedSerializer.Marshal(c, bs)
}

func (c Copy) SizeTypedBinco() (size int) {
	return CopyTypedSerializer.Size(c)
}

type Insert struct {
	Str string
}

func (i Insert) Do() {
	fmt.Printf("insert %q\n", i.Str)
}

func (i Insert) MarshalTypedBinco(bs []byte) (n int) {
	return InsertTypedSerializer.Marshal(i, bs)
}

func (i Insert) SizeTypedBinco() (size int) {
	return InsertTypedSerializer.Size(i)
}
```

**`typed/oneof/serializers.go`**

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const (
	CopyDTM binco.DTM = iota + 1
	InsertDTM
)

var (
	CopySerializer   = copySerializer{}
	InsertSerializer = insertSerializer{}

	CopyTypedSerializer   = typeid.NewSerializer[Copy](CopyDTM, CopySerializer)
	InsertTypedSerializer = typeid.NewSerializer[Insert](InsertDTM, InsertSerializer)

	InstructionSerializer = instructionSerializer{}
)

type copySerializer struct{}

func (copySerializer) Marshal(v Copy, bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Start, bs)
	return n + varnum.Int.Marshal(v.End, bs[n:])
}

func (copySerializer) Unmarshal(bs []byte) (v Copy, n int, err error) {
	v.Start, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.End, n1, err = varnum.Int.Unmarshal(bs[n:])
	n += n1
	return
}

func (copySerializer) Size(v Copy) (size int) {
	return varnum.Int.Size(v.Start) + varnum.Int.Size(v.End)
}

func (copySerializer) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = varnum.Int.Skip(bs[n:])
	return n + n1, err
}

type insertSerializer struct{}

func (insertSerializer) Marshal(v Insert, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (insertSerializer) Unmarshal(bs []byte) (v Insert, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (insertSerializer) Size(v Insert) (size int) {
	return composite.String.Size(v.Str)
}

func (insertSerializer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}

type instructionSerializer struct{}

func (instructionSerializer) Marshal(v Instruction, bs []byte) (n int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%v doesn't implement binco.MarshallerTyped", reflect.TypeOf(v)))
	}
	return m.MarshalTypedBinco(bs)
}

func (instructionSerializer) Size(v Instruction) (size int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%v doesn't implement binco.MarshallerTyped", reflect.TypeOf(v)))
	}
	return m.SizeTypedBinco()
}

func (instructionSerializer) Unmarshal(bs []byte) (v Instruction, n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case CopyDTM:
		v, n1, err = CopyTypedSerializer.UnmarshalData(bs[n:])
	case InsertDTM:
		v, n1, err = InsertTypedSerializer.UnmarshalData(bs[n:])
	default:
		err = fmt.Errorf("unexpected instruction: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}

func (instructionSerializer) Skip(bs []byte) (n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case CopyDTM:
		n1, err = CopyTypedSerializer.SkipData(bs[n:])
	case InsertDTM:
		n1, err = InsertTypedSerializer.SkipData(bs[n:])
	default:
		err = fmt.Errorf("unexpected instruction: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}
```

**`typed/oneof/main.go`**

```go
// Command oneof encodes a Go interface: a type identifier followed by the concrete value.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	values := []Instruction{
		Copy{Start: 10, End: 20},
		Insert{Str: "hello world"},
	}
	for _, v := range values {
		bs := binco.Encode(InstructionSerializer, v)
		fmt.Printf("marshaled %T %+v → %v (%d bytes)\n", v, v, bs, len(bs))

		back, n, err := InstructionSerializer.Unmarshal(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  unmarshaled %T (read %d bytes): ", back, n)
		back.Do()
	}

	var copyIns Instruction = Copy{Start: 1, End: 2}
	bs := binco.Encode(InstructionSerializer, copyIns)
	n, err := InstructionSerializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nskipped %d of %d bytes\n", n, len(bs))

	unknown := append([]byte{99}, bs[1:]...)
	if _, _, err = InstructionSerializer.Unmarshal(unknown); err != nil {
		fmt.Printf("unknown instruction: %v\n", err)
	}
}
```

Output:

```text
marshaled main.Copy {Start:10 End:20} → [1 20 40] (3 bytes)
  unmarshaled main.Copy (read 3 bytes): copy from 10 to 20
marshaled main.Insert {Str:hello world} → [2 11 104 101 108 108 111 32 119 111 114 108 100] (13 bytes)
  unmarshaled main.Insert (read 13 bytes): insert "hello world"

skipped 3 of 3 bytes
unknown instruction: unexpected instruction: binco: unexpected DTM 99
```

**Interface inside a struct.** Nothing special: the field's codec is the
interface codec, and the DTM is written where the field sits.

**"Oneof" with a marker field.** If you prefer explicit tags over DTMs, encode
a small enum before the value:

```go
type Shape struct {
	Kind  ShapeKind // 0 circle, 1 square
	Value any       // interpreted according to Kind
}
```

DTMs cost the same byte and are less error-prone, so prefer them.

### 6.10 Versioning and migration

Type versions are DTMs plus a dispatcher. The recommended strategy is one
DTM per version: each version has its own codec, the current codec reads all of
them, and old payloads are converted by a migration function. Old payloads stay
readable forever, and new payloads only ever contain the current version.

The complete program below has `FooV1` and `FooV2`, writes the current version,
decodes both, and shows `binco.Marshal`/`binco.MarshalTyped` on the same value.

**`typed/versioning/foo.go`**

```go
package main

type Foo FooV2

func (f Foo) MarshalBinco(bs []byte) (n int) {
	return FooSerializer.Marshal(f, bs)
}

func (f Foo) SizeBinco() (size int) {
	return FooSerializer.Size(f)
}

func (f Foo) String() string {
	return FooV2(f).Str
}
```

**`typed/versioning/foo_v1.go`**

```go
package main

type FooV1 struct {
	Num int
}

func (v FooV1) MarshalTypedBinco(bs []byte) (n int) {
	return FooV1TypedSerializer.Marshal(v, bs)
}

func (v FooV1) SizeTypedBinco() (size int) {
	return FooV1TypedSerializer.Size(v)
}
```

**`typed/versioning/foo_v2.go`**

```go
package main

type FooV2 struct {
	Str string
}

func (v FooV2) MarshalTypedBinco(bs []byte) (n int) {
	return FooV2TypedSerializer.Marshal(v, bs)
}

func (v FooV2) SizeTypedBinco() (size int) {
	return FooV2TypedSerializer.Size(v)
}
```

**`typed/versioning/serializers.go`**

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const (
	FooV1DTM binco.DTM = iota + 1
	FooV2DTM
)

var (
	FooV1Serializer = fooV1Serializer{}
	FooV2Serializer = fooV2Serializer{}

	FooV1TypedSerializer = typeid.NewSerializer[FooV1](FooV1DTM, FooV1Serializer)
	FooV2TypedSerializer = typeid.NewSerializer[FooV2](FooV2DTM, FooV2Serializer)

	FooSerializer = fooSerializer{}
)

type fooV1Serializer struct{}

func (fooV1Serializer) Marshal(v FooV1, bs []byte) (n int) {
	return varnum.Int.Marshal(v.Num, bs)
}

func (fooV1Serializer) Unmarshal(bs []byte) (v FooV1, n int, err error) {
	v.Num, n, err = varnum.Int.Unmarshal(bs)
	return
}

func (fooV1Serializer) Size(v FooV1) (size int) {
	return varnum.Int.Size(v.Num)
}

func (fooV1Serializer) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

type fooV2Serializer struct{}

func (fooV2Serializer) Marshal(v FooV2, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (fooV2Serializer) Unmarshal(bs []byte) (v FooV2, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (fooV2Serializer) Size(v FooV2) (size int) {
	return composite.String.Size(v.Str)
}

func (fooV2Serializer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}

type fooSerializer struct{}

func (fooSerializer) Marshal(v Foo, bs []byte) (n int) {
	return FooV2TypedSerializer.Marshal(FooV2(v), bs)
}

func (fooSerializer) Size(v Foo) (size int) {
	return FooV2TypedSerializer.Size(FooV2(v))
}

func (fooSerializer) Unmarshal(bs []byte) (v Foo, n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case FooV1DTM:
		var old FooV1
		old, n1, err = FooV1TypedSerializer.UnmarshalData(bs[n:])
		if err != nil {
			return
		}
		v = migrateFooV1(old)
	case FooV2DTM:
		var cur FooV2
		cur, n1, err = FooV2TypedSerializer.UnmarshalData(bs[n:])
		v = Foo(cur)
	default:
		err = fmt.Errorf("unexpected version: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}

func (fooSerializer) Skip(bs []byte) (n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case FooV1DTM:
		n1, err = FooV1TypedSerializer.SkipData(bs[n:])
	case FooV2DTM:
		n1, err = FooV2TypedSerializer.SkipData(bs[n:])
	default:
		err = fmt.Errorf("unexpected version: %w", binco.NewUnexpectedDTMError(dtm))
	}
	return n + n1, err
}

func migrateFooV1(v FooV1) Foo {
	return Foo(FooV2{Str: strconv.Itoa(v.Num)})
}
```

**`typed/versioning/main.go`**

```go
// Command versioning reads payloads written by an older version and converts them to the current type.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	v1 := FooV1{Num: 10}
	v1BS := binco.Encode(FooV1TypedSerializer, v1)
	fmt.Printf("V1 payload: %v (%d bytes)\n", v1BS, len(v1BS))

	v2 := FooV2{Str: "current"}
	v2BS := binco.Encode(FooV2TypedSerializer, v2)
	fmt.Printf("V2 payload: %v (%d bytes)\n\n", v2BS, len(v2BS))

	for _, bs := range [][]byte{v1BS, v2BS} {
		foo, n, err := FooSerializer.Unmarshal(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("decoded %d bytes: %q\n", n, foo)
	}

	foo := Foo(FooV2{Str: "rewritten"})
	bs := binco.Encode(FooSerializer, foo)
	fmt.Printf("\nmarshaled %q → %v\n", foo, bs)
	back, err := binco.Decode(FooSerializer, bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded back: %q\n", back)

	fmt.Printf("binco.Marshal:      %v\n", binco.Marshal(foo))
	fmt.Printf("binco.MarshalTyped: %v\n", binco.MarshalTyped(v1))

	for _, bs := range [][]byte{v1BS, v2BS} {
		n, err := FooSerializer.Skip(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("skipped %d of %d bytes\n", n, len(bs))
	}

	unknown := append([]byte{99}, v2BS[1:]...)
	if _, _, err = FooSerializer.Unmarshal(unknown); err != nil {
		fmt.Printf("\nunknown version: %v\n", err)
	}
}
```

Output:

```text
V1 payload: [1 20] (2 bytes)
V2 payload: [2 7 99 117 114 114 101 110 116] (9 bytes)

decoded 2 bytes: "10"
decoded 9 bytes: "current"

marshaled "rewritten" → [2 9 114 101 119 114 105 116 116 101 110]
decoded back: "rewritten"
binco.Marshal:      [2 9 114 101 119 114 105 116 116 101 110]
binco.MarshalTyped: [1 20]
skipped 2 of 2 bytes
skipped 9 of 9 bytes

unknown version: unexpected version: binco: unexpected DTM 99
```

**Other strategies.** A version field inside the payload (a `uint8` first field)
is cheaper than a DTM when the type is already inside a typed envelope, but it
cannot express a change of the payload's overall shape. Side-by-side types with
their own DTMs work when the versions are semantically different rather than
successive revisions.

**Rules of thumb.**

* Never change the meaning of an existing DTM.
* Keep the old version's codec, not just its data, for as long as payloads can
  exist (logs, queues, archives).
* Migrations run while decoding; keep them allocation-free where possible.
* Validate after migrating if the migration cannot fail.

### 6.11 Streaming protocols

A stream has no structure of its own, so the codec supplies it. Three patterns
cover most protocols: a fixed sequence both sides know, type-tagged messages
(DTS), and length-prefixed frames.

The complete program below is a DTS: it writes a login, a chat message, an
upload (which the reader skips without allocating) and another chat message on
one connection, then demonstrates that an unknown message type stops the
reader.

**`typed/stream/messages.go`**

```go
// The messages of the stream share the connection; each one is written as identifier followed by data.
package main

import (
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wtypeid "github.com/malivvan/binco/wirefmt/typeid"
)

const (
	LoginDTM binco.DTM = iota + 1
	ChatDTM
	UploadDTM
)

type Login struct {
	User string
}

type Chat struct {
	From string
	Text string
	At   time.Time
}

type Upload struct {
	Name string
	Data []byte
}

type typedMessage interface {
	writeTo(w wirefmt.Writer) (n int, err error)
}

func (m Login) writeTo(w wirefmt.Writer) (n int, err error) {
	return LoginTyped.Marshal(m, w)
}

func (m Chat) writeTo(w wirefmt.Writer) (n int, err error) {
	return ChatTyped.Marshal(m, w)
}

func (m Upload) writeTo(w wirefmt.Writer) (n int, err error) {
	return UploadTyped.Marshal(m, w)
}

var (
	LoginTyped  = wtypeid.NewSerializer(LoginDTM, loginSerializer{})
	ChatTyped   = wtypeid.NewSerializer(ChatDTM, chatSerializer{})
	UploadTyped = wtypeid.NewSerializer(UploadDTM, uploadSerializer{})
)

type loginSerializer struct{}

func (loginSerializer) Marshal(v Login, w wirefmt.Writer) (n int, err error) {
	return wcomposite.String.Marshal(v.User, w)
}

func (loginSerializer) Unmarshal(r wirefmt.Reader) (v Login, n int, err error) {
	v.User, n, err = wcomposite.String.Unmarshal(r)
	return
}

func (loginSerializer) Size(v Login) (size int) {
	return wcomposite.String.Size(v.User)
}

func (loginSerializer) Skip(r wirefmt.Reader) (n int, err error) {
	return wcomposite.String.Skip(r)
}

type chatSerializer struct{}

func (chatSerializer) Marshal(v Chat, w wirefmt.Writer) (n int, err error) {
	n, err = wcomposite.String.Marshal(v.From, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.String.Marshal(v.Text, w)
	n += n1
	if err != nil {
		return
	}
	n1, err = wscalar.TimeUnixUTC.Marshal(v.At, w)
	return n + n1, err
}

func (chatSerializer) Unmarshal(r wirefmt.Reader) (v Chat, n int, err error) {
	v.From, n, err = wcomposite.String.Unmarshal(r)
	if err != nil {
		return
	}
	var n1 int
	v.Text, n1, err = wcomposite.String.Unmarshal(r)
	n += n1
	if err != nil {
		return
	}
	v.At, n1, err = wscalar.TimeUnixUTC.Unmarshal(r)
	n += n1
	return
}

func (chatSerializer) Size(v Chat) (size int) {
	return wcomposite.String.Size(v.From) + wcomposite.String.Size(v.Text) +
		wscalar.TimeUnixUTC.Size(v.At)
}

func (chatSerializer) Skip(r wirefmt.Reader) (n int, err error) {
	n, err = wcomposite.String.Skip(r)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.String.Skip(r)
	n += n1
	if err != nil {
		return
	}
	n1, err = wscalar.TimeUnixUTC.Skip(r)
	return n + n1, err
}

type uploadSerializer struct{}

func (uploadSerializer) Marshal(v Upload, w wirefmt.Writer) (n int, err error) {
	n, err = wcomposite.String.Marshal(v.Name, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.ByteSlice.Marshal(v.Data, w)
	return n + n1, err
}

func (uploadSerializer) Unmarshal(r wirefmt.Reader) (v Upload, n int, err error) {
	v.Name, n, err = wcomposite.String.Unmarshal(r)
	if err != nil {
		return
	}
	var n1 int
	v.Data, n1, err = wcomposite.ByteSlice.Unmarshal(r)
	n += n1
	return
}

func (uploadSerializer) Size(v Upload) (size int) {
	return wcomposite.String.Size(v.Name) + wcomposite.ByteSlice.Size(v.Data)
}

func (uploadSerializer) Skip(r wirefmt.Reader) (n int, err error) {
	n, err = wcomposite.String.Skip(r)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wcomposite.ByteSlice.Skip(r)
	return n + n1, err
}
```

**`typed/stream/main.go`**

```go
// Command stream receives messages of different types over one connection, skipping the ones it does not need and stopping at an unknown type.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	wtypeid "github.com/malivvan/binco/wirefmt/typeid"
)

func main() {
	fmt.Println("1. receiving a stream of typed messages")
	stats, err := transfer(sampleMessages())
	if err != nil {
		panic(err)
	}
	fmt.Printf("   received %d messages, skipped %d, consumed %d bytes\n",
		stats.received, stats.skipped, stats.bytes)

	fmt.Println("\n2. an unknown message type stops the reader")
	stats, err = transfer([]typedMessage{unknownMessage{}})
	if unexpected, ok := errors.AsType[binco.UnexpectedDTMError](err); ok {
		fmt.Printf("   stopped with %v (DTM %d)\n", err, unexpected.DTM())
	} else {
		fmt.Printf("   stopped with %v\n", err)
	}
	fmt.Printf("   %d messages were received before the unknown one\n", stats.received)
}

func transfer(messages []typedMessage) (stats, error) {
	client, server := net.Pipe()

	sendErr := make(chan error, 1)
	go func() {
		defer func() { _ = client.Close() }()
		w := bufio.NewWriter(client)
		for _, m := range messages {
			if _, err := m.writeTo(w); err != nil {
				sendErr <- err
				return
			}
			if err := w.Flush(); err != nil {
				sendErr <- err
				return
			}
		}
		sendErr <- nil
	}()

	stats, err := receive(bufio.NewReader(server))

	_ = server.Close()
	if sendErr := <-sendErr; sendErr != nil && err == nil {
		err = sendErr
	}
	return stats, err
}

type stats struct {
	received int
	skipped  int
	bytes    int
}

func receive(r wirefmt.Reader) (stats, error) {
	var st stats
	for {

		dtm, n, err := wtypeid.DTMSerializer.Unmarshal(r)
		if errors.Is(err, io.EOF) {
			return st, nil
		}
		if err != nil {
			return st, err
		}
		st.bytes += n

		switch dtm {
		case LoginDTM:
			login, n, err := LoginTyped.UnmarshalData(r)
			if err != nil {
				return st, err
			}
			st.received++
			st.bytes += n
			fmt.Printf("   login:  %q joined\n", login.User)

		case ChatDTM:
			chat, n, err := ChatTyped.UnmarshalData(r)
			if err != nil {
				return st, err
			}
			st.received++
			st.bytes += n
			fmt.Printf("   chat:   %s says %q at %s\n", chat.From, chat.Text,
				chat.At.Format(time.RFC3339))

		case UploadDTM:

			n, err := UploadTyped.SkipData(r)
			if err != nil {
				return st, err
			}
			st.skipped++
			st.bytes += n
			fmt.Printf("   upload: skipped %d bytes\n", n)

		default:

			return st, fmt.Errorf("unknown message type: %w", binco.NewUnexpectedDTMError(dtm))
		}
	}
}

func sampleMessages() []typedMessage {
	at := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	return []typedMessage{
		Login{User: "alice"},
		Chat{From: "alice", Text: "hello everyone", At: at},
		Upload{Name: "report.pdf", Data: []byte{1, 2, 3, 4}},
		Chat{From: "bob", Text: "thanks!", At: at.Add(time.Minute)},
	}
}

type unknownMessage struct{}

func (unknownMessage) writeTo(w wirefmt.Writer) (n int, err error) {

	if n, err = wtypeid.DTMSerializer.Marshal(binco.DTM(99), w); err != nil {
		return
	}
	var n1 int
	n1, err = w.Write([]byte{0xDE, 0xAD})
	return n + n1, err
}
```

Output:

```text
1. receiving a stream of typed messages
   login:  "alice" joined
   chat:   alice says "hello everyone" at 2024-05-06T07:08:09Z
   upload: skipped 16 bytes
   chat:   bob says "thanks!" at 2024-05-06T07:09:09Z
   received 3 messages, skipped 1, consumed 75 bytes

2. an unknown message type stops the reader
   stopped with unknown message type: binco: unexpected DTM 99 (DTM 99)
   0 messages were received before the unknown one
```

Three rules make a DTS robust:

* Treat `io.EOF` **between** messages as a clean shutdown;
  `io.ErrUnexpectedEOF` **inside** a message is a truncated wire.
* Flush after each message if the peer should react to it immediately.
* Never ignore an unknown DTM: its length is unknowable, so skipping it would
  desynchronise the wire. Report `binco.NewUnexpectedDTMError(dtm)` and stop.

**Length-prefixed frames.** When you need to buffer a message before decoding
it (for example to hand it to another goroutine), prefix it with its size:

```go
payload := binco.Encode(OrderSerializer, order)
wirefmt.Encode(wvarnum.PositiveInt, conn, len(payload))
conn.Write(payload)
```

The receiver reads the size, allocates once, and calls `io.ReadFull`.

### 6.12 Validating untrusted input

A length read from the network is an allocation request. Validators check it
**before** the allocation happens, and they run on the reader: `Marshal` never
validates, and `Skip` performs only structural checks, so it never calls a
validator.

The three programs below validate a string length, slice length and elements,
and map length, keys and values.

**`validation/string/main.go`**

```go
// Command string validates a string length while decoding, before the characters are read.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	stropts "github.com/malivvan/binco/config/string"
)

const maxLen = 10

var ErrTooLong = errors.New("string too long")

type lengthLimit struct{ max int }

func (l lengthLimit) Validate(n int) error {
	if n > l.max {
		return fmt.Errorf("%w: %d > %d", ErrTooLong, n, l.max)
	}
	return nil
}

func main() {

	serializer := composite.NewStringSerializer(stropts.WithLenValidator(lengthLimit{max: maxLen}))

	for _, value := range []string{"hello", strings.Repeat("a", maxLen+1)} {
		bs := binco.Encode(serializer, value)
		got, _, err := serializer.Unmarshal(bs)
		switch {
		case err != nil:
			fmt.Printf("%-14q → rejected: %v (errors.Is: %v)\n", value, err,
				errors.Is(err, ErrTooLong))
		default:
			fmt.Printf("%-14q → accepted: %q\n", value, got)
		}
	}

	big := strings.Repeat("a", maxLen+1)
	got, _, err := composite.String.Unmarshal(binco.Encode(composite.String, big))
	if err != nil {
		panic(err)
	}
	fmt.Printf("plain reader accepts %d characters\n", len(got))

	fn := binco.ValidatorFn[int](func(n int) error {
		if n%2 != 0 {
			return errors.New("odd length")
		}
		return nil
	})
	fmt.Printf("ValidatorFn(3) = %v\n", fn.Validate(3))
}
```

Output:

```text
"hello"        → accepted: "hello"
"aaaaaaaaaaa"  → rejected: string too long: 11 > 10 (errors.Is: true)
plain reader accepts 11 characters
ValidatorFn(3) = odd length
```

**`validation/slice/main.go`**

```go
// Command slice validates the length and the elements of a slice while decoding.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	slopts "github.com/malivvan/binco/config/slice"
)

var (
	ErrTooMany = errors.New("too many elements")

	ErrEmptyElem = errors.New("empty element")
)

const maxElems = 2

func main() {
	lenVl := binco.ValidatorFn[int](func(length int) error {
		if length > maxElems {
			return fmt.Errorf("%w: %d > %d", ErrTooMany, length, maxElems)
		}
		return nil
	})
	elemVl := binco.ValidatorFn[string](func(elem string) error {
		if elem == "" {
			return ErrEmptyElem
		}
		return nil
	})

	serializer := composite.NewSliceSerializer(
		composite.String,
		slopts.WithLenValidator[string](lenVl),
		slopts.WithElemValidator[string](elemVl),
	)

	values := [][]string{
		{"hello", "world"},
		{"hello", "world", "again"},
		{"hello", ""},
	}
	for _, value := range values {
		bs := binco.Encode(serializer, value)
		got, _, err := serializer.Unmarshal(bs)
		if err != nil {
			fmt.Printf("%-28s → rejected: %v\n", render(value), err)
			continue
		}
		fmt.Printf("%-28s → accepted: %s\n", render(value), render(got))
	}

	n, err := serializer.Skip(binco.Encode(serializer, values[1]))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Skip: %d bytes, no validators called\n", n)
}

func render(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
```

Output:

```text
["hello", "world"]           → accepted: ["hello", "world"]
["hello", "world", "again"]  → rejected: too many elements: 3 > 2
["hello", ""]                → rejected: empty element
Skip: 19 bytes, no validators called
```

**`validation/map/main.go`**

```go
// Command map validates the length, the keys and the values of a map while decoding.
package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	mapopts "github.com/malivvan/binco/config/map"
	"github.com/malivvan/binco/varnum"
)

var (
	ErrTooMany = errors.New("too many entries")

	ErrNegativeKey = errors.New("negative key")

	ErrEmptyValue = errors.New("empty value")
)

const maxEntries = 1

func main() {
	lenVl := binco.ValidatorFn[int](func(length int) error {
		if length > maxEntries {
			return fmt.Errorf("%w: %d > %d", ErrTooMany, length, maxEntries)
		}
		return nil
	})
	keyVl := binco.ValidatorFn[int](func(key int) error {
		if key < 0 {
			return fmt.Errorf("%w: %d", ErrNegativeKey, key)
		}
		return nil
	})
	valueVl := binco.ValidatorFn[string](func(value string) error {
		if value == "" {
			return ErrEmptyValue
		}
		return nil
	})

	serializer := composite.NewMapSerializer(
		varnum.Int,
		composite.String,
		mapopts.WithLenValidator[int, string](lenVl),
		mapopts.WithKeyValidator[int, string](keyVl),
		mapopts.WithValueValidator[int, string](valueVl),
	)

	values := []map[int]string{
		{1: "one"},
		{1: "one", 2: "two"},
		{-1: "minus one"},
		{1: ""},
	}
	for _, value := range values {
		bs := binco.Encode(serializer, value)
		got, _, err := serializer.Unmarshal(bs)
		if err != nil {
			fmt.Printf("%-22s → rejected: %v\n", render(value), err)
			continue
		}
		fmt.Printf("%-22s → accepted: %s\n", render(value), render(got))
	}
}

func render(m map[int]string) string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%d:%q", k, m[k]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
```

Output:

```text
{1:"one"}              → accepted: {1:"one"}
{1:"one", 2:"two"}     → rejected: too many entries: 2 > 1
{-1:"minus one"}       → rejected: negative key: -1
{1:""}                 → rejected: empty value
```

**Order of checks.** Length, then keys, then values, each as soon as the item is
decoded. A bad length never leads to a large allocation, and a bad key stops the
decode before its value is read.

**What validation cannot do.** It cannot bound the total size of nested
containers (each level is checked independently) and it cannot reject a payload
that is merely semantically wrong (an impossible date, an unknown enum). Add
domain checks after decoding for that.

### 6.13 Partial decoding and out-of-order reads

Because `Skip` is exact and `Unmarshal` never reads past a value, you can index
into a payload without decoding it. The program below encodes a record as `ID`,
`Name`, `Score`, computes the offsets with `Skip`, and then reads the fields in
reverse order.

**`out_of_order/main.go`**

```go
// Command out_of_order computes field offsets with Skip and then decodes the fields in reverse order.
package main

import (
	"fmt"

	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

var (
	idSerializer    = varnum.Int
	nameSerializer  = composite.String
	scoreSerializer = scalar.Int32
)

func main() {
	const (
		id    = 42
		name  = "out of order"
		score = int32(1234)
	)
	bs := encode(id, name, score)
	fmt.Printf("payload: %v (%d bytes)\n\n", bs, len(bs))

	idEnd, err := idSerializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	nameEnd, err := nameSerializer.Skip(bs[idEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("offsets: ID ends at %d, Name ends at %d\n\n", idEnd, idEnd+nameEnd)

	gotScore, _, err := scoreSerializer.Unmarshal(bs[idEnd+nameEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("score: %d\n", gotScore)

	gotName, _, err := nameSerializer.Unmarshal(bs[idEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("name:  %q\n", gotName)

	gotID, _, err := idSerializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("id:    %d\n", gotID)

	if _, err = idSerializer.Skip(bs); err != nil {
		panic(err)
	}
	if _, err = nameSerializer.Skip(bs[idEnd:]); err != nil {
		panic(err)
	}
	onlyScore, _, err := scoreSerializer.Unmarshal(bs[idEnd+nameEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nonly the score: %d\n", onlyScore)

	if _, err = nameSerializer.Skip(bs[idEnd : idEnd+5]); err != nil {
		fmt.Printf("truncated payload: %v\n", err)
	}
}

func encode(id int, name string, score int32) []byte {
	bs := make([]byte, idSerializer.Size(id)+nameSerializer.Size(name)+scoreSerializer.Size(score))
	n := idSerializer.Marshal(id, bs)
	n += nameSerializer.Marshal(name, bs[n:])
	scoreSerializer.Marshal(score, bs[n:])
	return bs
}
```

Output:

```text
payload: [84 12 111 117 116 32 111 102 32 111 114 100 101 114 210 4 0 0] (18 bytes)

offsets: ID ends at 1, Name ends at 14

score: 1234
name:  "out of order"
id:    42

only the score: 1234
truncated payload: binco: truncated input
```

The same works with the streaming API, but only forwards: a stream cannot seek,
so values are skipped to reach the next one. To read the fields of one streamed
value in reverse order, buffer that value first and use the buffer API.

### 6.14 Zero-copy hot paths

`zerocopy` decodes strings and byte slices without copying: the returned value
**aliases the input buffer**. That is the fastest option and the easiest to get
wrong. The program below shows the payload equivalence with `composite`, the
aliasing pitfall (a reused buffer changes strings that were already decoded), the
fixes, and the array codec.

**`native/main.go`**

```go
// Command native shows the zero-copy codecs: decoded strings and slices point into the input buffer, so they must be copied when they outlive it.
package main

import (
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/zerocopy"
)

func main() {

	value := "hello"
	fmt.Printf("zerocopy.String: %v\n", binco.Encode(zerocopy.String, value))
	fmt.Printf("composite.String:    %v\n\n", binco.Encode(composite.String, value))

	messages := []string{"first", "second"}
	buf := make([]byte, 64)

	var (
		aliased []string
		copied  []string
	)
	for _, msg := range messages {

		n := copy(buf, binco.Encode(zerocopy.String, msg))

		str, _, err := zerocopy.String.Unmarshal(buf[:n])
		if err != nil {
			panic(err)
		}

		fmt.Printf("processing %q\n", str)

		aliased = append(aliased, str)

		copied = append(copied, strings.Clone(str))
	}
	fmt.Printf("\naliased strings after the loop: %q\n", aliased)
	fmt.Printf("copied strings after the loop:  %q\n", copied)

	payload := binco.Encode(zerocopy.ByteSlice, []byte{1, 2, 3})
	data, _, err := zerocopy.ByteSlice.Unmarshal(payload)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\ndecoded bytes: %v\n", data)
	payload[1] = 9
	fmt.Printf("after modifying the payload: %v (the decoded slice changed)\n", data)

	safe, _, err := composite.ByteSlice.Unmarshal(payload)
	if err != nil {
		panic(err)
	}
	payload[1] = 7
	fmt.Printf("composite.ByteSlice copied %v and is unaffected by the payload change to %v\n",
		safe, payload)

	parts := make([]string, 0, 4)
	for _, v := range []uint64{0, 40, 1 << 40, ^uint64(0)} {
		bs := binco.Encode(zerocopy.Uint64, v)
		got, _, err := zerocopy.Uint64.Unmarshal(bs)
		if err != nil {
			panic(err)
		}
		parts = append(parts, fmt.Sprintf("%d (%d bytes)", got, len(bs)))
	}
	fmt.Printf("zerocopy.Uint64 round trips: %s\n", strings.Join(parts, ", "))

	arr := [3]int32{7, 8, 9}
	arrSerializer := zerocopy.NewArraySerializer[[3]int32, int32](scalar.Int32)
	arrBS := binco.Encode(arrSerializer, arr)
	gotArr, _, err := arrSerializer.Unmarshal(arrBS)
	if err != nil {
		panic(err)
	}
	fmt.Printf("native array: %v → %v (%d bytes)\n", arr, gotArr, len(arrBS))
}
```

Output:

```text
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

The rules:

1. **The value lives only as long as the buffer.** If the buffer is reused, the
   string changes under you.
2. **Never write through the returned slice.** It is the payload.
3. **Copy when it must outlive the buffer**: `strings.Clone(s)`,
   `bytes.Clone(b)`, or use `composite.String`/`composite.ByteSlice`, which copy
   while decoding.

Scalars are converted through `encoding/binary` with an explicit little-endian
order, so the package is correct on any host byte order and at an unaligned
offset. Only the array codec reinterprets memory, and it validates its type
parameters at construction (`binco.ErrWrongArrayType`).

### 6.15 Custom codecs

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

#### Self-marshalling types

A type can encode itself through `binco.Marshaller` (`MarshalBinco`,
`SizeBinco`) or `binco.MarshallerTyped` (`MarshalTypedBinco`, `SizeTypedBinco`).
`binco.Marshal` and `binco.MarshalTyped` then need no serializer argument.
Decoding has no shorthand, because the payload does not carry its type: name the
serializer, or read the DTM first and dispatch.

The method names carry the package name so a type can also satisfy
`encoding/json`, `encoding/binary` and friends without collisions. The complete
program below defines one type that implements both interfaces and one that
implements only `Marshaller`.

**`marshal_func/foo.go`**

```go
// Foo marshals itself through binco.Marshaller and binco.MarshallerTyped.
package main

type Foo struct {
	Num int
}

func (f Foo) MarshalBinco(bs []byte) (n int) {
	return FooSerializer.Marshal(f, bs)
}

func (f Foo) SizeBinco() (size int) {
	return FooSerializer.Size(f)
}

func (f Foo) MarshalTypedBinco(bs []byte) (n int) {
	return FooTypedSerializer.Marshal(f, bs)
}

func (f Foo) SizeTypedBinco() (size int) {
	return FooTypedSerializer.Size(f)
}
```

**`marshal_func/bar.go`**

```go
package main

type Bar struct {
	Str string
}

func (b Bar) MarshalBinco(bs []byte) (n int) {
	return BarSerializer.Marshal(b, bs)
}

func (b Bar) SizeBinco() (size int) {
	return BarSerializer.Size(b)
}
```

**`marshal_func/serializers.go`**

```go
package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const FooDTM binco.DTM = 1

var (
	FooSerializer = fooSerializer{}
	BarSerializer = barSerializer{}

	FooTypedSerializer = typeid.NewSerializer[Foo](FooDTM, FooSerializer)
)

type fooSerializer struct{}

func (fooSerializer) Marshal(v Foo, bs []byte) (n int) {
	return varnum.Int.Marshal(v.Num, bs)
}

func (fooSerializer) Unmarshal(bs []byte) (v Foo, n int, err error) {
	v.Num, n, err = varnum.Int.Unmarshal(bs)
	return
}

func (fooSerializer) Size(v Foo) (size int) {
	return varnum.Int.Size(v.Num)
}

func (fooSerializer) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

type barSerializer struct{}

func (barSerializer) Marshal(v Bar, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (barSerializer) Unmarshal(bs []byte) (v Bar, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (barSerializer) Size(v Bar) (size int) {
	return composite.String.Size(v.Str)
}

func (barSerializer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}
```

**`marshal_func/main.go`**

```go
// Command marshal_func compares the explicit Size/Marshal pair with the binco.Marshal and binco.MarshalTyped shorthands.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	foo := Foo{Num: 10}
	bs := make([]byte, FooSerializer.Size(foo))
	n := FooSerializer.Marshal(foo, bs)
	fmt.Printf("explicit:      %v (wrote %d of %d bytes)\n", bs, n, len(bs))

	fmt.Printf("binco.Marshal:  %v\n", binco.Marshal(foo))

	bar := Bar{Str: "10"}
	fmt.Println("Foo and Bar marshal through the same call:")
	for _, m := range []binco.Marshaller{foo, bar} {
		fmt.Printf("  %-10T → %v\n", m, binco.Marshal(m))
	}

	fmt.Printf("\nbinco.Marshal:      %v\n", binco.Marshal(foo))
	fmt.Printf("binco.MarshalTyped: %v\n", binco.MarshalTyped(foo))

	back, err := binco.Decode(FooSerializer, bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\ndecoded Foo: %+v\n", back)

	typedBS := binco.MarshalTyped(foo)
	back, err = binco.Decode(FooTypedSerializer, typedBS)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded Foo (typed, %d bytes): %+v\n", len(typedBS), back)
}
```

Output:

```text
explicit:      [20] (wrote 1 of 1 bytes)
binco.Marshal:  [20]
Foo and Bar marshal through the same call:
  main.Foo   → [20]
  main.Bar   → [2 49 48]

binco.Marshal:      [20]
binco.MarshalTyped: [1 20]

decoded Foo: {Num:10}
decoded Foo (typed, 2 bytes): {Num:10}
```


### 6.16 Composition recursion and generics

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
writing a codec: a defined type with an `int` underlying type is a two-line
wrapper around `varnum.Int` (see [Custom codecs](#615-custom-codecs)), and the
generator writes exactly that code for `AddDefinedType` (see
[Generator methods](#73-generator-methods)).

### 6.17 Zero-allocation encoding

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

### 6.18 Concurrency

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

### 6.19 Testing your codecs

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

### 6.20 Interoperability

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
[The wire format](#13-the-wire-format): LEB128 varint, ZigZag for signed values,
little-endian raw numbers, positive varints for lengths, and the pointer flags.
Any implementation that follows them interoperates.

## 7. Code generation

`binco/codegen` writes the serializer code you would otherwise write by hand. It
inspects types through reflection **once**, at generation time; the generated
code contains no reflection.

### 7.1 Getting started

The generator inspects types through `reflect` and writes the serializers you
would write by hand, for the buffer API or (with `genopts.WithStream`) for the
streaming API. The complete program below is a real generator setup: the types,
the generator program, the `//go:generate` directive, the generated file and the
code that uses it.

The layout is:

```
app/
  go.mod
  main.go              // uses the generated serializers; holds go:generate
  pkg/types.go         // the types
  pkg/binco.gen.go     // generated, committed
  gen/main.go          // the generator program
```

**`pkg/types.go`**

```go
// Package pkg declares the types the generator produces serializers for.
package pkg

type Int int

type Foo[T any] struct {
	S string
	T T
}
```

**`gen/main.go`**

```go
// Command gen writes the serializers of the pkg package. Run it through go generate.
package main

import (
	"log"
	"os"
	"reflect"

	"example.com/app/pkg"

	"github.com/malivvan/binco/codegen"
	genopts "github.com/malivvan/binco/codegen/config/gen"
)

func main() {

	g, err := codegen.NewGenerator(
		genopts.WithPath("example.com/app/pkg"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = g.AddDefinedType(reflect.TypeFor[pkg.Int]()); err != nil {
		log.Fatal(err)
	}

	if err = g.AddStruct(reflect.TypeFor[pkg.Foo[pkg.Int]]()); err != nil {
		log.Fatal(err)
	}

	bs, err := g.Generate()
	if err != nil {

		if writeErr := os.WriteFile("failed.gen.go", bs, 0o644); writeErr != nil {
			log.Fatal(writeErr)
		}
		log.Fatal(err)
	}
	if err = os.WriteFile("./pkg/binco.gen.go", bs, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

**`pkg/binco.gen.go (generated)`**

```go
// Code generated by binco/codegen. DO NOT EDIT.

package pkg

import (
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

var IntBinco = intBinco{}

type intBinco struct{}

func (s intBinco) Marshal(v Int, bs []byte) (n int) {
	return varnum.Int.Marshal(int(v), bs)
}

func (s intBinco) Unmarshal(bs []byte) (v Int, n int, err error) {
	tmp, n, err := varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	v = Int(tmp)
	return
}

func (s intBinco) Size(v Int) (size int) {
	return varnum.Int.Size(int(v))
}

func (s intBinco) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

// FooSchemaFingerprint identifies the wire layout of Foo[Int]: a hash
// of its field names and types in declaration order. The format is positional,
// so reordering fields silently changes what a stored payload means; compare
// this constant with the value the writer used to detect that.
const FooSchemaFingerprint = "d36b2f28fcf4f7a73be3a18964b84d190205fe57d41e1203788fc330875a874c"

var FooBinco = fooBinco{}

type fooBinco struct{}

func (s fooBinco) Marshal(v Foo[Int], bs []byte) (n int) {
	n = composite.String.Marshal(v.S, bs)
	return n + IntBinco.Marshal(v.T, bs[n:])
}

func (s fooBinco) Unmarshal(bs []byte) (v Foo[Int], n int, err error) {
	v.S, n, err = composite.String.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.T, n1, err = IntBinco.Unmarshal(bs[n:])
	n += n1
	return
}

func (s fooBinco) Size(v Foo[Int]) (size int) {
	size = composite.String.Size(v.S)
	return size + IntBinco.Size(v.T)
}

func (s fooBinco) Skip(bs []byte) (n int, err error) {
	n, err = composite.String.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = IntBinco.Skip(bs[n:])
	n += n1
	return
}
```

**`main.go`**

```go
//go:generate go run ./gen

// Command gen uses the serializers generated into pkg.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
	"example.com/app/pkg"
)

func main() {

	num := pkg.Int(300)
	fmt.Printf("Int:     %v (varint)\n", binco.Encode(pkg.IntBinco, num))

	foo := pkg.Foo[pkg.Int]{S: "hello world", T: 5}
	bs := make([]byte, pkg.FooBinco.Size(foo))
	n := pkg.FooBinco.Marshal(foo, bs)
	fmt.Printf("Foo:     %v (%d bytes)\n", bs, n)

	back, n, err := pkg.FooBinco.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %+v (read %d bytes)\n", back, n)

	n, err = pkg.FooBinco.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)

	var serializer binco.Serializer[pkg.Int] = pkg.IntBinco
	fmt.Printf("through binco.Serializer[int]: size=%d\n", serializer.Size(num))
}
```

Output:

```text
Int:     [216 4] (varint)
Foo:     [11 104 101 108 108 111 32 119 111 114 108 100 10] (13 bytes)
decoded: {S:hello world T:5} (read 13 bytes)
skipped: 13 bytes
through binco.Serializer[int]: size=2
```

`go generate` runs the directive at the top of `main.go`
(`//go:generate go run ./gen`). `genopts.WithPath` is the only required option:
it is the import path of the package the generated file belongs to.

Naming: the serializer variable is `<Type>Binco`, a DTM constant is `<Type>DTM`,
and a typed serializer is `<Type>TypedBinco`. Use `genopts.WithSerializer` to
change it (see [Serialization options](#75-serialization-options)).

### 7.2 Generator configuration

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

### 7.3 Generator methods

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

### 7.4 Generated code and naming

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

### 7.5 Serialization options

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

#### Complete example

The program below applies three options to one struct: a custom serializer name,
raw fixed-width encoding for a number, and a length validator for a string. The
generated file shows the effect directly.

**`pkg/types.go`**

```go
// Package pkg declares the type generated with a custom serializer name and per-field options.
package pkg

import (
	"errors"
	"fmt"
)

const MaxStrLen = 100

var ErrTooLong = errors.New("string too long")

type Foo struct {
	Num int
	Str string
}

func ValidateLength(length int) error {
	if length > MaxStrLen {
		return fmt.Errorf("%w: %d > %d", ErrTooLong, length, MaxStrLen)
	}
	return nil
}
```

**`gen/main.go`**

```go
// Command gen writes a serializer with per-field options. Run it through go generate.
package main

import (
	"log"
	"os"
	"reflect"

	"example.com/app/pkg"

	"github.com/malivvan/binco/codegen"
	fldopts "github.com/malivvan/binco/codegen/config/field"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

func main() {
	g, err := codegen.NewGenerator(
		genopts.WithPath("example.com/app/pkg"),

		genopts.WithSerializer(reflect.TypeFor[pkg.Foo](), "CustomFoo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = g.AddStruct(reflect.TypeFor[pkg.Foo](),

		stopts.WithField(
			fldopts.WithType(tpopts.WithNumEncoding(tpopts.NumEncodingRaw)),
		),

		stopts.WithField(
			fldopts.WithType(tpopts.WithLenValidator("ValidateLength")),
		),
	); err != nil {
		log.Fatal(err)
	}

	bs, err := g.Generate()
	if err != nil {
		if writeErr := os.WriteFile("failed.gen.go", bs, 0o644); writeErr != nil {
			log.Fatal(writeErr)
		}
		log.Fatal(err)
	}
	if err = os.WriteFile("./pkg/binco.gen.go", bs, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

**`pkg/binco.gen.go (generated)`**

```go
// Code generated by binco/codegen. DO NOT EDIT.

package pkg

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/scalar"
)

var (
	stringOpNvQW5om8MLxbCPLTsHBgΞΞ = composite.NewStringSerializer(stropts.WithLenValidator(binco.ValidatorFn[int](ValidateLength)))
)

// CustomFooSchemaFingerprint identifies the wire layout of Foo: a hash
// of its field names and types in declaration order. The format is positional,
// so reordering fields silently changes what a stored payload means; compare
// this constant with the value the writer used to detect that.
const CustomFooSchemaFingerprint = "03437121bdc167620d9f71555cda4157489041a513ae756be640d3851eaa580d"

var CustomFooBinco = customFooBinco{}

type customFooBinco struct{}

func (s customFooBinco) Marshal(v Foo, bs []byte) (n int) {
	n = scalar.Int.Marshal(v.Num, bs)
	return n + stringOpNvQW5om8MLxbCPLTsHBgΞΞ.Marshal(v.Str, bs[n:])
}

func (s customFooBinco) Unmarshal(bs []byte) (v Foo, n int, err error) {
	v.Num, n, err = scalar.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Str, n1, err = stringOpNvQW5om8MLxbCPLTsHBgΞΞ.Unmarshal(bs[n:])
	n += n1
	return
}

func (s customFooBinco) Size(v Foo) (size int) {
	size = scalar.Int.Size(v.Num)
	return size + stringOpNvQW5om8MLxbCPLTsHBgΞΞ.Size(v.Str)
}

func (s customFooBinco) Skip(bs []byte) (n int, err error) {
	n, err = scalar.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = stringOpNvQW5om8MLxbCPLTsHBgΞΞ.Skip(bs[n:])
	n += n1
	return
}
```

**`main.go`**

```go
//go:generate go run ./gen

// Command gen_options uses a serializer generated with a custom name, a fixed-width number codec and a length validator.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"example.com/app/pkg"
)

func main() {

	foo := pkg.Foo{Num: 1, Str: "hello"}
	bs := binco.Encode(pkg.CustomFooBinco, foo)
	fmt.Printf("value:   %+v\n", foo)
	fmt.Printf("size:    %d bytes (Num alone takes 8: raw, not varint)\n", len(bs))
	fmt.Printf("encoded: %v\n", bs)

	back, err := binco.Decode(pkg.CustomFooBinco, bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %+v\n\n", back)

	long := pkg.Foo{Num: 1, Str: strings.Repeat("a", pkg.MaxStrLen+1)}
	longBS := binco.Encode(pkg.CustomFooBinco, long)
	_, err = binco.Decode(pkg.CustomFooBinco, longBS)
	fmt.Printf("string of %d characters: %v\n", len(long.Str), err)
	fmt.Printf("errors.Is(err, pkg.ErrTooLong): %v\n", errors.Is(err, pkg.ErrTooLong))

	ok := pkg.Foo{Num: 2, Str: strings.Repeat("a", pkg.MaxStrLen)}
	back, err = binco.Decode(pkg.CustomFooBinco, binco.Encode(pkg.CustomFooBinco, ok))
	if err != nil {
		panic(err)
	}
	fmt.Printf("string of %d characters: accepted, %q…\n", len(back.Str), back.Str[:8])
}
```

Output:

```text
value:   {Num:1 Str:hello}
size:    14 bytes (Num alone takes 8: raw, not varint)
encoded: [1 0 0 0 0 0 0 0 5 104 101 108 108 111]
decoded: {Num:1 Str:hello}

string of 101 characters: string too long: 101 > 100
errors.Is(err, pkg.ErrTooLong): true
string of 100 characters: accepted, "aaaaaaaa"…
```


### 7.6 Interfaces

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

### 7.7 Versioned types

The generated versioned codec is exactly the hand-written one from
[Versioning and migration](#610-versioning-and-migration): `Marshal` writes the
current version, `Unmarshal` accepts every version and migrates, `Skip` advances
over every version.

Pin the identifiers with `genopts.WithDTM` in anything that is stored or sent.
Without it the generator numbers the types in registration order, which is stable
within one run but renumbers the payloads as soon as a type is added or removed.
Generated constants carry literal values, a duplicate id or a negative id fails
generation with `binco.ErrDuplicateDTM`, and mixing pinned and automatic ids is
rejected. `g.DTMRegistry()` returns the mapping so you can commit it as a golden
file and compare it in CI.

### 7.8 Multi-package and cross-package generation

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

### 7.9 Streaming mode

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

### 7.10 Workflow and troubleshooting

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


## 8. Package guide

### 8.1 binco

The root package holds everything the packages share: the codec contract
(defined in full in [Codecs](#41-codecs)), the error values, the safety helpers
and the marshalling shorthand below.

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
                                      // Get(unsafe.Pointer, reflect.Type), Len, Next, Clear
revPtrMap := binco.NewReversePtrMap() // Put(int, unsafe.Pointer, reflect.Type),
                                      // Get(int) (unsafe.Pointer, reflect.Type, bool),
                                      // Len, Enter, Leave, Clear
```

**Limits and widths**

```go
binco.Num8RawSize, binco.Num16RawSize, binco.Num32RawSize, binco.Num64RawSize
binco.Uint8MaxVarintLen ... binco.Uint64MaxVarintLen
binco.Uint8MaxLastByte  ... binco.Uint64MaxLastByte
```

**Type constraints** for writing generic codecs: `Signed`, `Unsigned`,
`Integer8`, `Integer16`, `Integer32`, `Integer64`, `Num32`, `Num64`.

**Errors**: see the [error reference](#10-error-reference), plus
`UnexpectedDTMError` (`DTM()`) and `WrongDTMError` (`Expected()`, `Actual()`).

### 8.2 varnum

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

### 8.3 scalar

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

### 8.4 composite

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

### 8.5 zerocopy

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

### 8.6 ptrmap

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
   maps (see [Concurrency](#618-concurrency)).

An identifier is stored together with the type it was decoded as, and the encoder
keys its table by `(address, type)`. Two pointers that share an address (a struct
and its first field, or any two zero-size values) therefore get different ids, and
a reference read as a different type fails with `binco.ErrWrongType` instead of
being cast with `unsafe.Pointer`. Recursive references stop at
`binco.MaxDepth` with `binco.ErrTooDeep`.

### 8.7 typeid

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

### 8.8 config

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

### 8.9 wirefmt

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


## 9. Cheat sheet

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


## 10. Error reference

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


## 12. Performance guide

**Plan for the buffer.** `Size` is exact and cheap; use it to allocate once and
reuse. The `Encoder` pattern in
[Zero-allocation encoding](#617-zero-allocation-encoding) is the recommended
shape for hot paths.

**Encode fixed-width slices in bulk.** A slice whose element codec reports a
fixed width (`scalar` and `zerocopy` scalars) is copied in one operation instead of
one interface call per element, provided the host is little-endian. The fast path
is automatic; element validators keep the per-element path.

**Pick the codec family per field.**

* `varnum` for small and medium integers, ids, counters, lengths.
* `scalar` for values above the thresholds in [Fixed width](#134-fixed-width), for constant-size
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


## 13. The wire format

### 13.1 Anatomy of a payload

A payload is a sequence of encoded values with no separators:

```
[ field 1 ][ field 2 ][ field 3 ] ...
```

There is no header, no version, no length, no checksum and no padding. The
receiver reads values in the order the schema declares them. Every value is
self-delimiting: a number by its width or its LEB128 continuation bits, a
container by its length prefix, a pointer by its flag byte.

### 13.2 Booleans and bytes

`bool` and `byte` occupy exactly one byte.

```
true     ->  [1]
false    ->  [0]
byte 7   ->  [7]
```

A `bool` byte other than 0 or 1 is rejected with `binco.ErrWrongFormat`.

### 13.3 Varint

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

### 13.4 Fixed width

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

Its advantage is speed: no shifts, no branches, and the compiler usually turns
the `encoding/binary` calls into a single load or store.

### 13.5 Strings and byte slices

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

### 13.6 Slices and arrays

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

### 13.7 Maps

A map is a length followed by key/value pairs:

```
map[int]string{1: "hello"}  ->  [1 2 5 104 101 108 108 111]
```

Entries appear in Go's map iteration order, which is random. Two marshals of the
same map may produce different bytes, though both decode to equal maps. If you
need canonical bytes, encode a sorted slice of pairs instead; see
[Collections](#65-collections).

### 13.8 Pointers

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
[Shared pointers and identity](#67-shared-pointers-and-identity).

### 13.9 Time

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

### 13.10 Type identifiers

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

### 13.11 A payload dissected

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

### 13.12 Compatibility rules

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
decode the old one; see [Versioning and migration](#610-versioning-and-migration).

The generated `SchemaFingerprint` constant is the cheap check: store it beside a
payload (or compare it at deployment time) and a reordered field is caught before
the data is misread.


## 14. FAQ

**Where should I start?** The [Start here](#start-here) table maps a task to the
section that answers it. [Quick start](#3-quick-start) encodes and decodes a
value in three steps and ends with a short reading list.

**Buffer API or streaming API?** Use the buffer API unless the payload is a flow
that is too large to hold in memory or arrives on a connection;
[Buffers and streams](#42-buffers-and-streams) compares them. The bytes are
identical, so a buffer payload can be decoded by a streaming codec and vice
versa (through a `bytes.Reader`/`bytes.Buffer`).

**How do I debug a payload that will not decode?** Check the error with
`errors.Is` first - `ErrTruncated` means the payload ended inside a value,
`ErrWrongFormat` means a byte is not a valid flag. Then dissect the bytes with
[A payload dissected](#1311-a-payload-dissected) and the rest of
[The wire format](#13-the-wire-format). `binco.DecodeExact` reports leftover
bytes with `ErrTrailingBytes`.

**Can I modify a decoded string or byte slice?** With `composite`, yes: its
codecs copy. With `zerocopy`, no: the value aliases the input buffer. Copy it
with `strings.Clone`/`bytes.Clone`, or use the `composite` codecs.

**Can I use binco for data that must be readable by a program in another
language?** Yes. The format is described in [The wire format](#13-the-wire-format)
and uses only LEB128 varints, ZigZag, little-endian fixed-width integers, length
prefixes and the pointer flags. You need to replicate the schema, not a runtime.

**Why are there no field tags?** Because the schema is code. Tags would add
bytes to every payload; the receiver already knows the order.

**Can I skip a field I do not understand?** If you know its codec, yes:
`Skip`. If you do not know its length (an unknown DTM, an unknown schema), no;
stop and report it.

**Do I need the generator?** No. The hand-written codec in
[Structs by hand](#63-structs-by-hand) is exactly what the generator emits.

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
which is the entry point for contributors and coding agents. The everyday
commands are `go test ./...`, `make cover-check`, `make doc-check` and
`make portable`; AGENTS.md explains each one. The user-facing documentation
is this file; the wire format in
[The wire format](#13-the-wire-format) is the shared reference.


## License

MIT. See [LICENSE](./LICENSE).
