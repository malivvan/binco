# AGENTS.md

Working notes for anyone — human or coding agent — changing `binco`.

`README.md` is the user manual: how to encode data, what the bytes mean, and how
to use each package. This file is the other half: how the repository is built,
what it promises, where a change belongs, and how to verify it.

The project is preparing its **1.0.0** release: 1.0.0 is the first stable
release, and the wire format described in README §13 is frozen from it. Until the
tag is pushed no API is frozen, so renames and simpler designs are welcome when
they make the library easier to use - but a change is only done when its
documentation, tests and generated files are updated with it.

## Start here

| If you want to… | Read |
|-----------------|------|
| understand the layout | §3 Repository map |
| understand how a codec works | §4 Architecture |
| add or change a codec | §8 Adding to the library |
| touch the generator or templates | §7 The code generator |
| change the bytes on the wire | §6 Wire format and compatibility |
| run the checks before finishing | §2 Commands, §9 Testing |
| prepare a release | §12 CI and release |

## 1. Ground rules

1. **No third-party dependencies.** `go.mod` requires nothing; there is no
   `go.sum` and no `vendor/` directory. The few helpers the generator used to take
   from `golang.org/x/{mod,tools}` live in `internal/`. The module must build and
   test with `GOPROXY=off`.
2. **Two codec trees, one format.** Every codec exists for a `[]byte` and for an
   `io.Reader`/`io.Writer`. Both must emit and accept the same bytes; a change to
   one is a change to the other.
3. **Every target Go supports must build.** The wire format is little-endian on
   every host and no codec may depend on the host's byte order or word size.
   `make portable` compiles every target in `go tool dist list` and compiles the
   test files on a representative cross-section; a change that only builds on
   amd64 is not done. Android and iOS link main packages through cgo only, so
   there the script builds the library packages and says so.
4. **Stream decoding is bounded by the bytes that arrive.** A length read from a
   stream never becomes an upfront allocation, a codec never reads past the value
   it decodes, and a reader that stops making progress fails with
   `io.ErrNoProgress` instead of blocking. `wirefmt/safety_test.go` enforces this
   for every codec family.
5. **Every exported package-level identifier has a doc comment**, and every
   exported behaviour has a section in the README or this file. Codec method sets
   (`Marshal`, `Unmarshal`, `Size`, `Skip`) are covered by their type's comment
   instead of four separate ones. Documentation is part of the change, not a
   follow-up.
6. **Errors instead of panics, except in `Marshal`.** Encoding has no error return,
   so a short output buffer panics with `binco.ErrBufferTooSmall`. Nothing else
   panics on data: malformed input returns an error. The only other panics are
   programming errors in constructors (see §5).
7. **Tests must be deterministic.** No sleeps, no wall-clock assumptions, no
   network, no ordering that depends on Go's map iteration. The generated-code
   test runs the real `go` tool in a temporary module and is skipped by
   `go test -short`.
8. **No generated files are committed.** The generator is exercised by
   `TestGeneratedCodeCompilesAndRuns`, which writes its output into a temporary
   module and compiles it; the README's code-generation guides are the
   reference. Never add generated code to the repository by hand.

## 2. Commands

```bash
# Build and test everything.
go build ./...
go test ./...

# What CI runs.
gofmt -l .                       # must print nothing
go vet ./...
go test -race ./...
GOARCH=386 go test ./...          # the 32-bit wire-width and overflow paths
make cover-check                 # 100% statement coverage, whole module
make doc-check                   # README/AGENTS links, symbols, numbering
make portable                    # builds every target Go supports (slow)
make portable-quick              # the cross-section only, for iteration

# Every Go command above has a Makefile entry:
make build test race vet fmt lint cover cover-check doc-check portable portable-quick

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

`golangci-lint run --build-tags=fuzz` is what the lint job uses (`make lint`).
`-short` is the fast mode; it only skips the test that shells out to `go`.
`make cover` writes `coverage.out`; it is generated and ignored - never commit
it, and delete it if a stray copy is left behind.

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
├── testdata/           fixture types used by the generator tests
├── scripts/            covercheck.sh (coverage gate), doccheck.py (docs),
│                       portable.sh (every target)
├── Makefile            build/test/cover/lint entry points
├── NOCOVER.md          the few branches coverage cannot reach, with reasons
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
codegen ──► templates ──► internal/importfix
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
| `wireio` | bounded read/skip of N stream bytes: `ErrUnexpectedEOF` on truncation, `io.ErrNoProgress` on a stalled reader; `CopyN` also backs the stream adapter and `lew` | the streaming tree |
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

The format is specified in README §13 (The wire format). Repository-side rules:

* **The bytes are the contract.** `golden_test.go` (root) and
  `wirefmt/golden_test.go` pin the primitive encodings and the cross-tree parity.
  Changing them is a deliberate format change: update the vectors, the README and
  the CHANGELOG together.
* **`int`/`uint` are 64-bit on every host.** A 32-bit build reports
  `ErrOverflow` for values it cannot hold; tests must not assume the host width.
* **The wire bytes are little-endian on every host.** Every multi-byte value goes
  through `encoding/binary`'s explicit `LittleEndian` (or shifts); no codec
  reinterprets host memory. The only raw-memory path is the fixed-width slice
  copy, gated by `internal/fixedwidth.Of` on `binco.HostLittleEndian`, which
  comes from `binary.NativeEndian` — never from a pointer probe.
  `endianness_test.go` pins the golden bytes on any host;
  `endianness_bigendian_test.go` (build tag `ppc64 || s390x || mips || mips64`)
  asserts that the fast path is disabled there.
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
4. Add a generator test that asserts the emitted fragment and, where the change
   is substantial, that the output compiles (`TestGeneratedCodeCompilesAndRuns`).

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
| `wirefmt` | `FuzzStreamDecode` (one value per streaming family: no panic, no over-read, `Skip` agreement) |
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

### Portability

* `scripts/portable.sh` cross-compiles every target in `go tool dist list`
  (`go build ./...`), then compiles the test files on a representative
  cross-section with `go test -exec=/bin/true`: both byte orders (ppc64, s390x,
  mips, mips64), 32-bit hosts, Windows, the BSDs and both wasm environments.
  `make portable` runs it; the `Portability` CI job runs it on every push.
* `endianness_test.go` and `endianness_bigendian_test.go` pin the byte order and
  the fast-path guard (see §6). `TestSliceBulkPathMatchesTheElementLoop` checks
  that the bulk copy and the per-element loop produce identical bytes on any
  host.
* A test that only works on one platform belongs behind a build tag and must
  still compile in the matrix.

### Stream safety

`wirefmt/safety_test.go` is the contract for decoding untrusted streams. For one
representative value of every codec family it checks that:

* a nine-byte payload declaring a `MaxInt64` length fails immediately and
  allocates less than 8 MiB;
* a reader that keeps returning `(0, nil)` fails with `io.ErrNoProgress` instead
  of blocking (`internal/wireio.CopyN` bounds the empty reads, and
  `internal/lew`'s fixed-width reads go through it);
* an I/O error in the middle of a value reaches the caller unchanged;
* every truncation of every value fails cleanly and never reads past the input;
* `Skip` reports the same outcome as `Unmarshal` for every prefix;
* decoding leaves the next byte intact, through the one-byte adapter and through
  `bufio` alike.

`FuzzStreamDecode` (build tag `fuzz`) fuzzes the same properties. Add a table
row when you add a codec family.

### Documentation

`scripts/doccheck.py` (also `make doc-check`; the `Docs` CI job) verifies that the
two documents match the code:

* every internal link in README.md resolves to a heading;
* every `package.Symbol` mentioned in README.md or AGENTS.md is exported by that
  package (looked up with `go doc`);
* README.md numbers its sections and subsections sequentially, and the table of
  contents lists exactly those headings in order;
* every `README §N` reference in AGENTS.md names an existing section;
* none of the retired names listed in `RETIRED` (in the script) survives.

A new documentation rule belongs in the script, not in a reviewer's memory.

### Coverage

Every package in the module must stay at **100% statement coverage**; the
Makefile measures `go list ./...`.

```bash
make cover-check   # go test -coverprofile over the non-demo packages, then
                   # scripts/covercheck.sh
```

`scripts/covercheck.sh` counts the statements listed in the `nocover` block of
[`NOCOVER.md`](./NOCOVER.md) as covered and fails on anything else. An entry is
only acceptable for a branch that cannot run at all — the file explains each one
and rejects entries that no longer name an uncovered block. Prefer deleting an
unreachable branch or making the dependency injectable over adding an entry.

### Platform coverage

* `-race` on every run;
* `GOARCH=386` to exercise the 32-bit range checks (`ErrOverflow`) and the
  host-independent wire width;
* `golangci-lint` with the `fuzz` tag so the targets are type-checked too.

## 10. Examples

Every task guide in README §6 is a complete, copyable program: the full source
(types, codecs and a `main` that prints its values and the encoded bytes) sits
next to the explanation, with its expected output below it. The generator guides
in §7 include the type, the generator program, the generated file and the code
that uses it.

There is no examples directory to keep in sync: an example that is not in the
README does not exist, so it cannot drift.

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
| `Test` | gofmt check, `go vet`, `go test -race -covermode=atomic -coverprofile=coverage.out ./...`, `scripts/covercheck.sh`, `go build ./...`, Coveralls upload |
| `Test (GOARCH=386)` | `GOARCH=386 go test ./...` on 32-bit |
| `Portability` | `scripts/portable.sh`: build every Go target, test-compile the cross-section |
| `Docs` | `python3 scripts/doccheck.py`: links, documented symbols, numbering, retired names |
| `Lint` | `golangci-lint run --build-tags=fuzz` |
| `Fuzz <target>` | one matrix entry per target: 15 s on a push, 5 min on the nightly schedule |
| `Publish release` | on a `v*` tag: creates the GitHub release |

### Definition of done

A change is ready when all of these hold:

1. `gofmt -l .` prints nothing and `go vet ./...` passes.
2. `go test ./...`, `go test -race ./...` and `GOARCH=386 go test ./...` pass.
3. `make cover-check` reports 100% (or `NOCOVER.md` explains the gap).
4. `make portable-quick` passes; before a release, run `make portable`.
5. `golangci-lint run --build-tags=fuzz` reports no issues.
6. `make doc-check` passes, and README and this file describe the new behaviour.
7. `CHANGELOG.md` has an entry for the change.

Release checklist:

1. Update `CHANGELOG.md`: move `Unreleased` into the new version and date it.
2. `gofmt -l .`, `go vet ./...`, `go test ./...`, `go test -race ./...`,
   `GOARCH=386 go test ./...`, `make cover-check`.
3. Tag `vX.Y.Z` and push the tag; CI publishes the release.

## 13. Traps to know about

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
* **Host byte order.** Never reinterpret a `[]byte` as a multi-byte type: encode
  with `encoding/binary` and leave raw-memory copies to the guarded bulk path.
* **Stalled readers.** `io.ReadFull`/`io.CopyN` spin forever on a reader that
  returns `(0, nil)`. The streaming tree uses `internal/wireio` for anything that
  reads a known number of bytes.
* **Fuzz seeds.** A failing input is written next to the package; commit it, do
  not delete it.
* **Generated artifacts.** `make cover-check` writes `coverage.out`; `go test
  -c` writes `*.test`; `go tool cover -html` writes `coverage.html`. All three
  are ignored and must not be committed. Nothing generated belongs in git.
* **Documentation drift.** README and this file are part of the contract:
  changing a command, an error, a default or a limit means updating both.
* **`/tmp` size.** Some sandboxes have a small `/tmp`; if `go test -race` fails
  with a disk-quota error, set `GOTMPDIR` to a directory on a larger disk.
