# AGENTS.md

Working notes for anyone — human or coding agent — changing `binco`.

`README.md` is the user manual: how to encode data, what the bytes mean, and how
to use each package. This file is the other half: how the repository is built,
what it promises, where a change belongs, and how to verify it.

The project is **pre-release**, so no API is frozen. Renames and simpler designs
are welcome when they make the library easier to use, but a change is only done
when its documentation, tests and generated files are updated with it.

## 1. Ground rules

1. **No third-party dependencies.** `go.mod` requires nothing; there is no
   `go.sum` and no `vendor/` directory. The few helpers the generator used to take
   from `golang.org/x/{mod,tools}` live in `internal/`. The module must build and
   test with `GOPROXY=off`.
2. **Two codec trees, one format.** Every codec exists for a `[]byte` and for an
   `io.Reader`/`io.Writer`. Both must emit and accept the same bytes; a change to
   one is a change to the other.
3. **Every exported identifier has a doc comment**, and every exported behaviour
   has a section in the README or this file. Documentation is part of the change,
   not a follow-up.
4. **Errors instead of panics, except in `Marshal`.** Encoding has no error return,
   so a short output buffer panics with `binco.ErrBufferTooSmall`. Nothing else
   panics on data: malformed input returns an error. The only other panics are
   programming errors in constructors (see §5).
5. **Tests must be deterministic.** No sleeps, no wall-clock assumptions, no
   network, no ordering that depends on Go's map iteration. The generated-code
   test runs the real `go` tool in a temporary module and is skipped by
   `go test -short`.
6. **Generated files are committed.** `demos/gen/pkg/binco.gen.go` and
   `demos/gen_options/pkg/binco.gen.go` are regenerated with `go generate ./...`;
   CI fails when the committed output differs from a fresh run.
7. **Do not edit a generated file by hand.** Change the generator or the source
   type, then regenerate.

## 2. Commands

```bash
# Build and test everything, including the demos.
go build ./...
go test ./...

# What CI runs.
gofmt -l .                       # must print nothing
go vet ./...
go test -race ./...
GOARCH=386 go test ./...          # the 32-bit wire-width and overflow paths
go generate ./... && git diff --exit-code   # generated code is current

# Faster iteration.
go test ./composite/ -run TestSorted -v
go test -short ./...             # skips the generated-code compilation test

# Fuzz targets live behind the "fuzz" build tag.
go test -tags=fuzz -run='^$' ./...                                  # compile them
go test -tags=fuzz -run='^$' -fuzz=FuzzString -fuzztime=30s ./composite
go test -tags=fuzz -run='^$' -fuzz=FuzzUnmarshal -fuzztime=30s ./varnum

# Offline check: the module must need no network.
GOPROXY=off GOFLAGS=-mod=mod go build ./... && GOPROXY=off GOFLAGS=-mod=mod go test ./...
```

`golangci-lint run --build-tags=fuzz` is what the lint job uses. `-short` is the
fast mode; it only skips the test that shells out to `go`.

## 3. Repository map

```
.
├── binco.go            package doc: the guarantees every codec keeps
├── codec.go            Serializer/Marshaller, Encode/Decode/DecodeExact, Append*
├── errors.go           every sentinel error, MaxDepth, MaxPointerDepth
├── length.go           overflow-safe length checks, CheckWrite, Depth
├── fixed.go            FixedWidth interface and HostLittleEndian (bulk slice path)
├── numbers.go          numeric constraints and varint width constants
├── pointer.go          PtrMap/ReversePtrMap (typed pointer ids)
├── validation.go       Validator/ValidatorFn
├── datatype.go         DTM
├── varnum/             LEB128 + ZigZag codecs
├── scalar/             fixed-width little-endian codecs and time
├── composite/          bool, string, []byte, pointer, slice, array-adjacent, map
├── zerocopy/           zero-copy strings/blobs, array codec, fast scalars
├── ptrmap/             shared pointers, cycles, Wrap
├── typeid/             DTM meta-codec and typed serializers
├── config/             option types: array, byte_slice, map, slice, string
├── wirefmt/            the streaming mirror of everything above, package by package
├── codegen/            the generator: API, spec model, builder, templates, scanner
├── internal/           shared implementation details (see §4.6)
├── demos/              runnable main packages, one per feature (see §10)
├── testdata/           fixture types used by the generator tests
└── .github/workflows/  CI
```

### The two trees

| Buffer | Stream | Wire bytes |
|--------|--------|------------|
| `binco` | `binco/wirefmt` | identical |
| `binco/varnum` | `binco/wirefmt/varnum` | identical |
| `binco/scalar` | `binco/wirefmt/scalar` | identical |
| `binco/composite` | `binco/wirefmt/composite` | identical |
| `binco/zerocopy` | `binco/wirefmt/zerocopy` | identical |
| `binco/ptrmap` | `binco/wirefmt/ptrmap` | identical |
| `binco/typeid` | `binco/wirefmt/typeid` | identical |
| `binco/config/*` | `binco/wirefmt/config/*` | n/a (options) |

The stream tree is a separate implementation, not a wrapper: a buffer codec knows
the whole input, a stream codec must cope with what has arrived. That costs
duplication, which `wirefmt/golden_test.go` guards: it encodes with one side and
decodes with the other, both ways, and compares the bytes. When you change a
codec family, add its value to that table.

Streaming-only rules (the buffer side has no counterpart):

* A stream that ends before the first byte of a value returns `io.EOF`; a value
  that starts and does not finish returns `io.ErrUnexpectedEOF`.
* `wirefmt.Decode` never buffers ahead, because it cannot know whether the caller
  still needs the bytes of the next message. `wirefmt.NewDecoder` owns one
  `bufio.Reader` and is the fast path for a stream the codec consumes.
* Lengths are never turned into an upfront allocation; blobs grow with the bytes
  that arrive (`internal/wireio`).

## 4. Architecture

### 4.1 The contract

```go
type Serializer[T any] interface {
    Marshal(t T, bs []byte) (n int)                  // panics with ErrBufferTooSmall
    Unmarshal(bs []byte) (t T, n int, err error)     // returns ErrTruncated on short input
    Size(t T) (size int)                             // exactly what Marshal writes
    Skip(bs []byte) (n int, err error)               // exactly what Unmarshal consumes
}
```

Streaming codecs mirror it with `(n int, err error)` on `Marshal` and
`Reader`/`Writer` instead of `[]byte`.

Four invariants hold for every codec, including generated ones:

1. `Size(v) == len(Marshal(v, buf))` for a large enough `buf`.
2. `Unmarshal` consumes exactly `Size` bytes of the value it returns.
3. `Skip` advances exactly as far as `Unmarshal` would, and applies the same range
   checks (a value `Unmarshal` rejects is not silently skipped).
4. Decoding never reads past the value and never panics on malformed input.

`binco.Encode`/`Decode` are the convenience wrappers; `DecodeExact` adds
`ErrTrailingBytes` for one-value-per-message protocols; `AppendEncode` and
`AppendMarshal*` grow a caller-owned buffer and verify the codec's size contract
(`ErrWrongSize`).

### 4.2 Layering

```
demos ──────────────► codegen ──► templates ──► internal/importfix
                         │
                         ├─► spec ◄── spec/builder ◄── scanner, classifier, typename
                         └─► spec/converter
user code ──► ptrmap, typeid ──► composite, zerocopy ──► varnum, scalar ──► binco
                                    │
                              config/* options
```

* `binco` knows nothing about the codec packages. It defines the contract, the
  error values and the safety helpers.
* `varnum`/`scalar`/`composite`/`zerocopy` depend only on `binco`, `config/*` and
  each other where a container needs an element codec.
* `ptrmap` and `typeid` compose codecs; they do not add wire types of their own.
* `wirefmt/*` mirrors the same layering for streams.
* `codegen` consumes the library and never the reverse.

### 4.3 Codec families

| Package | Role | Notes |
|---------|------|-------|
| `varnum` | LEB128, ZigZag, positive lengths | canonical encodings only; `PositiveInt*` range-check |
| `scalar` | fixed width, little-endian | implements `binco.FixedWidth` for the bulk slice path |
| `composite` | length-prefixed containers and pointers | per-element interface dispatch |
| `zerocopy` | aliasing strings/blobs, arrays | scalars converted with `encoding/binary`; arrays validated at construction |
| `ptrmap` | pointer identity, cycles | typed ids, per-value depth limit, `Wrap` clears the maps |
| `typeid` | DTM dispatch | positive ids only; typed serializer embeds the id |

A codec family must agree byte for byte with its siblings:
`varnum`/`scalar`/`zerocopy` on numbers, `composite`/`zerocopy` on strings and
blobs, buffer and stream everywhere.

### 4.4 Options

`config/{array,byte_slice,map,slice,string}` hold the option structs. The plain
`New*` constructors honour every validator, so there is no second constructor
family: validators are options. Constructors return `binco.Serializer[T]` (a
stream constructor returns `wirefmt.Serializer[T]`), which is all a caller needs.

The generated code names the option packages through the aliases `arropts`,
`bslopts`, `mapopts`, `slopts` and `stropts`, and the streaming libraries as
`wcomposite`, `wscalar`, ... in hand-written code; keep those spellings.

### 4.5 Limits and safety

| Mechanism | Where | Behaviour |
|-----------|-------|-----------|
| `binco.CheckLength` / `CheckLengthElem` | every length-prefixed codec | compares against the remaining bytes without adding, so a hostile length is `ErrTruncated`, not a panic |
| `binco.AllocHint` | slice/map/blob decoding | caps the initial allocation by the input, so five bytes cannot ask for gigabytes |
| `binco.MaxZeroSizeElements` | containers of zero-byte elements | `ErrTooLargeLength` above the cap |
| `binco.Depth` / `MaxDepth` | `ptrmap` | nesting per decoded value |
| `binco.MaxPointerDepth` | plain pointer codec | one atomic counter per serializer |
| canonical varints | `varnum` | `ErrNonCanonical` for a non-shortest encoding |
| typed pointer ids | `binco.PtrMap` / `ReversePtrMap` | keyed by `(address, reflect.Type)`, mismatch is `ErrWrongType` |

Call `binco.CheckWrite` before writing, or panic with `binco.ErrBufferTooSmall`
directly where a length is already known; never index the slice and let the
runtime panic.

### 4.6 `internal/`

Nothing under `internal/` is public API, and nothing outside this module may
import it.

| Package | Responsibility | Shared by |
|---------|----------------|-----------|
| `le` | little-endian fixed-width coding in a `[]byte` | `scalar`, `zerocopy` |
| `lew` | the same over an `io.Reader`/`io.Writer` | `wirefmt/scalar`, `wirefmt/zerocopy` |
| `fixedwidth` | decides whether a slice can be copied in bulk | both `composite` trees |
| `wireio` | read/skip exactly N stream bytes, with `ErrUnexpectedEOF` mapping | both streaming text codecs |
| `importfix` | prunes unused imports, dedupes, groups, gofmts | `codegen` |
| `modpath` | module and import path validation | `codegen/typename`, `codegen/config/gen` |

## 5. Errors and panics

| Error | Meaning |
|-------|---------|
| `binco.ErrTruncated` | the input ends inside a value |
| `binco.ErrBufferTooSmall` | `Marshal` would write past the buffer (panic) |
| `binco.ErrWrongSize` | a codec wrote a different number of bytes than `Size` reported |
| `binco.ErrOverflow` | a varint exceeds its type, or a value does not fit the host |
| `binco.ErrNonCanonical` | a non-shortest varint |
| `binco.ErrNegativeLength`, `ErrTooLargeLength`, `ErrWrongLength` | length rules |
| `binco.ErrWrongFormat` | an invalid flag byte or shape |
| `binco.ErrTooDeep` | nesting beyond `MaxDepth`/`MaxPointerDepth` |
| `binco.ErrWrongType` | a pointer id defined by another type |
| `binco.ErrTrailingBytes` | `DecodeExact` found data after the value |
| `binco.ErrWrongArrayType` | `NewArraySerializer` got mismatched type parameters (panic) |
| `binco.ErrDuplicateDTM` | the generator assigned one id twice |
| `codegen.ErrCodeGenFailed`, `codegen.ErrNameCollision` | generator failures; `Generate` wraps the cause, including a name collision, with `%w` |

`binco.ErrNilInterface`, `ErrSchemaMismatch`, `ErrUnsupportedArch` and
`ErrUnsupportedIntSize` are reserved: no code path raises them today, but the
names are stable and documented for callers that switch on them.

Rules:

* Wrap with `%w` (or use `errors.Is`-compatible sentinels); never compare error
  strings.
* A decoder returns the exact error it got from a nested codec; do not replace it.
* Streaming codecs pass I/O errors through unchanged, except the
  `io.EOF`/`io.ErrUnexpectedEOF` mapping described in §3.

## 6. Wire format and compatibility

The format is specified in README §5. Repository-side rules:

* **The bytes are the contract.** `golden_test.go` (root) and
  `wirefmt/golden_test.go` pin the primitive encodings and the cross-tree parity.
  Changing them is a deliberate format change: update the vectors, the README and
  the CHANGELOG together.
* **`int`/`uint` are 64-bit on every host.** A 32-bit build reports
  `ErrOverflow` for values it cannot hold; tests must not assume the host width.
* **Times decode to UTC.** Nothing in the codebase may return `time.Local`.
* **Lengths default to `varnum.PositiveInt`.** Changing a length codec changes the
  format.
* **Maps are unordered** unless `NewSortedMapSerializer` is used; do not make a
  test depend on map order.
* **Pointer ids start at zero per value** and are reset by `ptrmap.Wrap`.
* **DTM ids must be pinned** (`genopts.WithDTM`) for anything that outlives a
  build; `Generator.DTMRegistry` returns the mapping for a golden file.
* **Every generated struct carries a `SchemaFingerprint`** constant (SHA-256 of
  field names and types in declaration order). Changing field order changes it;
  that is the early warning the README promises.

## 7. The code generator

### Pipeline

```
Add*/Register*  ──► spec/builder ──► spec.* ──► text/template ──► internal/importfix ──► Generate
      │                  ▲
      │                  └── scanner + typename + classifier + spec/converter
      └── records queued work, foreign types, anonymous serializers, DTM groups
```

| File / package | Responsibility |
|----------------|----------------|
| `generator.go` | public API, queue, DTM resolution, rendering, `Generate` |
| `names.go` | naming (`FooBinco`, `FooTypedBinco`, `FooDTM`, schema constant), type→codec mapping (`Of`), name-collision detection |
| `templates.go` | template helper functions: signatures, last-argument spelling, patterns |
| `modes.go` | Go type → codec package per generator mode |
| `templates/*.tmpl` | the emitted source |
| `spec` | the data the templates render |
| `spec/builder` | reflect type → `spec`, validation of the registration |
| `spec/converter` | complete name ⇄ full name ⇄ relative name; records referenced packages for imports |
| `scanner` | decomposes type-name strings; does not use reflect |
| `typename` | name types and their parsing |
| `classifier` | type predicates (defined, struct, interface, primitive, ...) |
| `config/*` | the public option types |

### Import handling

The import block is built from two sources: the fixed superset in
`templates/imports.tmpl` (the codec packages, `fmt`, `reflect`) and the packages
the converter resolved while building types (`TypeNameConverter.Imports`).
`internal/importfix` then removes unused entries, deduplicates by path (keeping
an explicit alias) and formats the file.

Consequence: a package that appears **only inside a validator or migration
expression** has no type to derive it from and must be named with
`genopts.WithImport`/`WithImportAlias`. The generator does not guess import paths.

### Adding a generator feature

1. Add the option in `codegen/config/...` (with a doc comment and a test).
2. Carry it through `spec` and `spec/builder`; validate expressions with
   `builder.ValidateExpr`/`ValidateIdent` where user text is pasted into source.
3. Teach the relevant template(s), both `Stream` and non-`Stream` branches.
4. Regenerate the demos and add a generator test that asserts the emitted
   fragment and, where the change is substantial, that the output compiles
   (`TestGeneratedCodeCompilesAndRuns`).

### Adding a type kind

`scanner.Kind` → `spec/builder` (parse and validate) → `spec` (fields) →
`SerializerFns.Of` (codec choice) → templates (marshal/unmarshal/size/skip) →
tests for every mode (safe, unsafe, stream) and for the foreign/local cases.

## 8. Adding to the library

### A new codec

1. Buffer flavour: implement `binco.Serializer[T]`, use `binco.CheckLength` for
   anything length-prefixed and `binco.CheckWrite` before writing.
2. Stream flavour in the matching `wirefmt/…` package, with the same bytes and
   the EOF mapping.
3. `Size` must be exact; if the codec writes a fixed-width raw little-endian
   value, implement `FixedWidth()` so slices take the bulk path.
4. Tests: round trip, boundaries, truncation (`ErrTruncated`), short output
   (`ErrBufferTooSmall`), `Skip` agreement, and a golden vector. Add the value to
   `wirefmt/golden_test.go` when it belongs to a family that crosses trees.
5. Docs: package guide entry in the README, and a sentence here if it introduces
   a new mechanism.

### A new validator or option

Options live in `codegen/config` for the generator and in `config/…` for the
runtime. A runtime option is a field on the option struct plus a `With…`
function; the constructor reads it in one place and the streaming twin does the
same. Never silently ignore an option: either honour it or reject it.

### New internal helper

Only put something in `internal/` when at least two packages need it or when it
replaces a dependency. Document what it is *not* responsible for.

## 9. Testing

The suite is table-driven and lives next to the code. What it covers:

* every codec's round trip, boundaries (`0`, width edges, empty containers, nil),
  truncation and `Skip` agreement;
* buffer/stream byte parity (`wirefmt/golden_test.go`);
* pointer identity, cycles and depth limits;
* DTM dispatch, version migration and pinned-id stability;
* validators, allocation caps and zero-size containers;
* the generator: emitted fragments, ignored fields, field-option alignment,
  name collisions, DTM duplicates, schema fingerprints, and a full
  generate → compile → run cycle in a temporary module.

### Fuzz targets (build tag `fuzz`)

| Package | Targets |
|---------|---------|
| `composite` | `FuzzString`, `FuzzByteSlice`, `FuzzSlice`, `FuzzMap`, `FuzzPtr` |
| `zerocopy` | `FuzzString`, `FuzzByteSlice`, `FuzzArray`, `FuzzScalars` |
| `scalar` | `FuzzFixedWidth`, `FuzzTime` |
| `ptrmap` | `FuzzPointer`, `FuzzPointerSlice` |
| `typeid` | `FuzzDTM`, `FuzzTyped` |
| `varnum` | `FuzzUnmarshal` |
| `wirefmt/varnum` | `FuzzUnmarshal`, `FuzzSkip` |
| `wirefmt/composite` | `FuzzString`, `FuzzByteSlice`, `FuzzSlice`, `FuzzMap` |
| `codegen/scanner` | `FuzzParseDefinedType`, `FuzzParseContainerType`, `FuzzParsePrimitive`, `FuzzScan` |

Every target asserts the same properties: arbitrary bytes never panic, the
consumed count stays inside the input, `Skip` matches `Unmarshal`, and a decoded
value round-trips with an exact `Size`. The buffer container targets and the
`zerocopy` targets run under a memory limit so an allocation regression fails
instead of killing the runner. Crashers
are written to `<package>/testdata/fuzz/<target>`; commit them as seeds (one is
already committed under `varnum/testdata/fuzz`).

### Platform coverage

* `-race` on every run;
* `GOARCH=386` to exercise the 32-bit range checks (`ErrOverflow`) and the
  host-independent wire width;
* `golangci-lint` with the `fuzz` tag so the targets are type-checked too.

## 10. Examples

`demos/` contains a runnable `main` package for every feature; the compiler checks
them on every build, so a stale example is a broken build.

| Example | Shows |
|---------|-------|
| `types/array` | arrays, wrong-length rejection |
| `types/slice` | slices, length codecs, nil vs empty |
| `types/map` | maps, ordering caveat |
| `types/struct` | a hand-written struct codec |
| `validation/string`, `validation/slice`, `validation/map` | the three validator families |
| `out_of_order` | `Skip` offsets, reverse decoding |
| `typed/different`, `typed/oneof`, `typed/versioning`, `typed/stream` | DTMs, interfaces, versions, streaming messages |
| `pm/simple`, `pm/linked_list`, `pm/graph` | shared pointers, cycles, graphs |
| `native` | zero-copy pitfalls and fixes |
| `marshal_func` | `binco.Marshal`/`MarshalTyped` |
| `gen`, `gen_options` | code generation (these two are regenerated) |

```bash
go run ./demos/types/struct
go generate ./demos/gen ./demos/gen_options   # regenerate committed code
```

New demos should be small, self-contained `main` packages that print what they
show, and should be added to this table.

## 11. Style

* **Doc comments** start with the identifier's name and say what the thing is
  *for*, not how it works.
* **Errors** are values from `errors.go` or `fmt.Errorf("...: %w", err)`; the
  message is lower case and has no punctuation.
* **No hidden state.** A codec is a value; the only stateful types are
  `PtrMap`, `ReversePtrMap` (per-value, documented) and the generator (per run).
* **Aliases** in imports: `genopts` (generator), `tpopts`, `stopts`, `fldopts`,
  `veropts`, `intropts`, `arropts`, `bslopts`, `mapopts`, `slopts`, `stropts`,
  and `w…` for the streaming packages in hand-written code.
* **File names** describe a concern, not a Go type: `blobs.go`, `list.go`,
  `dict.go`, `time.go`, `pointer.go`, `memory.go`. Each package has a file named
  after it with the package doc.
* **Keep the README honest.** If a sentence becomes wrong, fix it in the same
  change; docs drift is a bug.

## 12. CI and release

| Job | What it does |
|-----|--------------|
| `Test` | gofmt check, `go vet`, `go test -race -cover`, `go generate` + `git diff --exit-code`, `go build ./...`, Coveralls upload |
| `Test (GOARCH=386)` | `go test ./...` on 32-bit |
| `Lint` | `golangci-lint run --build-tags=fuzz` |
| `Fuzz <target>` | one matrix entry per target: 15 s on a push, 5 min on the nightly schedule |
| `Publish release` | on a `v*` tag: creates the GitHub release |

Release checklist:

1. `go generate ./...` and commit any regenerated file.
2. Update `CHANGELOG.md` (user-visible changes, in the `Unreleased` section).
3. `gofmt -l .`, `go vet ./...`, `go test ./...`, `go test -race ./...`,
   `GOARCH=386 go test ./...`.
4. Tag `vX.Y.Z` and push the tag; CI publishes the release.

## 13. Traps to know about

* **Generated drift.** Changing a template without running `go generate ./...`
  fails CI. Run it before you finish.
* **Import guessing does not exist.** A validator that references another package
  needs `WithImport`; see §7.
* **DTM renumbering.** Adding a version or implementation renumbers unpinned ids.
  Pin them and compare `DTMRegistry` if payloads survive deployments.
* **Field options are positional.** `stopts.WithField` is called once per struct
  field, including ignored ones; an ignored field does not shift the options.
* **`zerocopy` aliasing.** A decoded string or `[]byte` points into the input
  buffer; never reuse or modify it while the value is alive.
* **32-bit builds.** `int` is still 64-bit on the wire; tests must use explicit
  `int64`/`uint64` values where the host cannot hold them.
* **Fuzz seeds.** A failing input is written next to the package; commit it, do
  not delete it.
* **`/tmp` size.** Some sandboxes have a small `/tmp`; if `go test -race` fails
  with a disk-quota error, set `GOTMPDIR` to a directory on a larger disk.
