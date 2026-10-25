# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Nothing yet.

## [1.0.0] - 2026-10-25

The first stable release. The wire format, the public API and the generated code
are frozen for the v1 series.

### Highlights

- Binary codecs for every built-in Go type: booleans, bytes, signed and unsigned
  integers of every width, floats, strings, byte slices, arrays, slices, maps,
  pointers, `time.Time`, defined types, interfaces and multiple versions of a
  type - in memory and over streams.
- One wire format, two APIs: the buffer API (`binco` and its subpackages) and the
  streaming API (`binco/wirefmt` and its subpackages) emit and accept identical
  bytes.
- Zero-copy decoding for strings, byte slices and arrays (`zerocopy`).
- Pointer mapping for shared references, cycles and graphs (`ptrmap`), with typed
  identifiers so a payload cannot reinterpret one type as another.
- Type identifiers (`typeid`) for polymorphism and versioned payloads.
- A code generator (`codegen`) that writes serializers for structs, defined
  types, interfaces, versioned types and streaming codecs, including
  multi-package generation.
- Validators for lengths, keys, values, elements and whole values, honoured by
  every constructor.
- No third-party dependencies: `go.mod` requires nothing, there is no `go.sum`
  and no `vendor/` directory, and the module builds and tests with `GOPROXY=off`.

### Added

- `binco.DecodeExact` and `ErrTrailingBytes` for protocols with exactly one value
  per message.
- `binco.AppendEncode`, `AppendMarshal` and `AppendMarshalTyped`.
- `binco.FixedWidth`, with bulk encode/decode of fixed-width element slices.
- `wirefmt.Encoder`, `wirefmt.Decoder`, `EncodeTo`, `DecodeFrom` and
  `wirefmt.SliceWriter`.
- `composite.NewSortedMapSerializer` for deterministic map payloads.
- `genopts.WithDTM`, `genopts.WithLogger`, `genopts.WithImport`,
  `genopts.WithImportAlias` and `Generator.DTMRegistry`.
- `SchemaFingerprint` constants in generated code.
- `binco.ErrWrongSize`, reported when a codec writes a different number of bytes
  than its `Size` promised.
- Fuzz targets for `composite`, `zerocopy`, `scalar`, `ptrmap` and `typeid`, the
  buffer `varnum` codecs, the streaming codecs (`FuzzStreamDecode`) and the
  generator scanner.
- `scripts/portable.sh` and the `Portability` CI job: every `go tool dist list`
  target is built, and the test files are compiled on a representative
  cross-section including big-endian, 32-bit, Windows and wasm.
- `scripts/doccheck.py` and the `Docs` CI job: README.md and AGENTS.md are checked
  against the code on every push.
- `AGENTS.md`, the contributor and coding-agent guide; `README.md` is the user
  manual.

### Security and robustness

- Decoded lengths are checked without integer overflow in every codec, so a
  `MaxInt64` length returns `ErrTruncated` instead of panicking.
- Decoded lengths never drive an allocation larger than the input that could hold
  them, in the buffer and streaming codecs alike; streaming blobs and strings grow
  with the bytes that arrive.
- `ptrmap` identifiers are typed. The encoder keys its table by (address, type),
  and a reference resolved as another type is `ErrWrongType`.
- Pointer codecs bound their recursion: `ptrmap` at `MaxDepth` per value, the
  plain pointer codec at `MaxPointerDepth`. Both report `ErrTooDeep` instead of
  exhausting the stack.
- `zerocopy.NewArraySerializer` validates its type parameters and panics with
  `ErrWrongArrayType` on misuse.
- `zerocopy` scalars are converted through `encoding/binary`, which fixes unaligned
  access and removes the big-endian restriction.
- `wirefmt.Decode` no longer buffers a reader it cannot read single bytes from, so
  it cannot swallow the next message of a shared stream. Truncation is reported
  as `io.ErrUnexpectedEOF` and a clean end as `io.EOF`.
- A reader that keeps returning `(0, nil)` fails with `io.ErrNoProgress` instead
  of blocking (`internal/wireio`).
- Varints must be canonical; overlong encodings are rejected with
  `ErrNonCanonical`. `PositiveInt*` range-checks on decode instead of wrapping,
  and `Skip` applies the same range checks as `Unmarshal`.
- Constructors honour every validator they are given.
- Containers whose elements consume no bytes are capped by `MaxZeroSizeElements`.
- The generator validates validator and migration expressions with `go/parser`
  and quotes import paths.

### Changed

- `int` and `uint` are encoded as 64 bits on every architecture; a value the host
  cannot hold is `ErrOverflow`.
- Time codecs always decode to UTC; `TimeUnix*` no longer returns a local time.
- Every constructor honours its validators, so there is one `New*` constructor
  per type, returning `binco.Serializer[T]`.
- Generated DTM constants carry literal values; duplicate, negative or mixed
  pinned/automatic ids fail generation.
- A nil interface value is written as the reserved identifier `0` and decodes back
  to nil instead of panicking.
- Generator functions return errors instead of panicking, and `Generate` wraps the
  real cause with `%w`.
- `Marshal` panics with one sentinel (`ErrBufferTooSmall`) from every codec; a
  decoder that runs out of input reports `ErrTruncated`.
- `ptrmap.Wrap` clears the maps in place, and anonymous serializer names use
  truncated SHA-256.
- `DecodeExact` returns `(T, error)` like `Decode`, and the `Append*` helpers take
  the destination first, like `append`.
- The streaming tree shares its little-endian, fixed-width-read and bulk-copy
  helpers (`internal/le`, `internal/lew`, `internal/fixedwidth`,
  `internal/wireio`) instead of duplicating them per package.
- Every example is a complete, copyable program in the README task guides; the
  separate `demos/` directory was removed, so there is no example code to keep
  in sync and coverage measures every package in the module.

### Migration notes from the pre-release API

- Package `zerocpy` is now `zerocopy` (`wirefmt/zerocpy` is `wirefmt/zerocopy`).
- The `NewValid*` constructors are gone: pass the validators as options to the
  ordinary `New*` constructor.
- `binco.UintMaxVarintLen` and `binco.UintMaxLastByte` are gone: `int` and `uint`
  always use the `uint64` limits.
- `binco.ErrTooSmallByteSlice` is split into `binco.ErrTruncated` (input that ends
  inside a value) and `binco.ErrBufferTooSmall` (a short output buffer).
- `binco.AllocHintElem` and the unused `Generator.scratch` field were removed.

### Fixed

- Field options were applied to the wrong field when an earlier field was ignored.
- A second type whose bare name matched an already generated one silently reused
  its serializer; it is now an `ErrNameCollision` error.
- `ArrayPattern` is anchored, so `int32` and `Foo1Bar` are no longer classified as
  arrays.
- A truncated versioned payload reported fewer consumed bytes than the equivalent
  interface payload.
- `Skip` accepted values that `Unmarshal` rejected as out of range.
- The streaming sorted-map serializer swallowed an element codec error and wrote a
  shifted payload; the buffer sorted-map serializer silently truncated an entry in
  a short buffer. Both now return the error, or `ErrWrongSize` when an element
  codec disagrees with its `Size`.
- Duplicated error constructors in `codegen/spec/converter` and unused ones in
  `codegen/typename`/`codegen/spec/builder` were removed.

[Unreleased]: https://github.com/malivvan/binco/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/malivvan/binco/releases/tag/v1.0.0
