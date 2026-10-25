# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Correctness, security, performance, ergonomics and documentation review
follow-up. The wire format is unchanged for canonical payloads produced by the
previous release, except where noted.

### Security

- Decoded lengths are checked without integer overflow in every codec, so a
  `MaxInt64` length returns `ErrTruncated` instead of panicking.
- Decoded lengths no longer drive an allocation larger than the input that could
  hold it, in the buffer and the streaming codecs alike; streaming blobs and
  strings grow with the bytes that arrive.
- `ptrmap` identifiers are typed. A reference is resolved only when the type it
  was defined with matches, and the encoder keys its table by (address, type).
  Type confusion is reported as `ErrWrongType`.
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
- Varints must be canonical; overlong encodings are rejected with
  `ErrNonCanonical`. `PositiveInt*` range-checks on decode instead of wrapping,
  and `Skip` applies the same range checks as `Unmarshal`.
- Constructors honour every validator they are given, including the plain
  `New*` ones.
- Containers whose elements consume no bytes are capped by `MaxZeroSizeElements`.
- The generator validates validator and migration expressions with `go/parser`
  and quotes import paths.

### Added

- `binco.DecodeExact` and `ErrTrailingBytes` for protocols that carry exactly one
  value per message.
- `binco.AppendEncode`, `AppendMarshal` and `AppendMarshalTyped`.
- `binco.FixedWidth`, with bulk encode/decode of fixed-width element slices.
- `wirefmt.Encoder`, `wirefmt.Decoder`, `EncodeTo`, `DecodeFrom` and
  `wirefmt.SliceWriter`.
- `composite.NewSortedMapSerializer` for deterministic map payloads.
- `genopts.WithDTM`, `genopts.WithLogger`, `Generator.DTMRegistry`.
- Generated `SchemaFingerprint` constants.
- Fuzz targets for `composite`, `zerocopy`, `scalar`, `ptrmap`, `typeid` and the
  buffer `varnum` codecs.
- `binco.ErrWrongSize`, reported when a codec writes a different number of bytes
  than its `Size` promised.
- `AGENTS.md`, the contributor and coding-agent guide; `README.md` is now the
  user manual only.

### Changed

- The module has no third-party dependencies. `go.mod` requires nothing, and
  `go.sum` and the `vendor/` tree are gone.
- The generator no longer uses `golang.org/x/tools/imports`. Its import block is
  built from the packages of the types it resolved, pruned of unused entries and
  formatted locally (`internal/importfix`). Path validation moved to
  `internal/modpath` instead of `golang.org/x/mod`. A package referenced only
  inside a validator or migration expression now needs `genopts.WithImport`.
- Little-endian coding, the bulk fixed-width check and the streaming
  fixed-length read now have one implementation each (`internal/le`,
  `internal/lew`, `internal/fixedwidth`, `internal/wireio`) instead of one per
  package.
- `int` and `uint` are encoded as 64 bits on every architecture; a value the host
  cannot hold is `ErrOverflow`.
- Time codecs always decode to UTC; `TimeUnix*` no longer returns a local time.
- The plain `NewStringSerializer`, `NewByteSliceSerializer`,
  `NewSliceSerializer`, `NewMapSerializer` and `NewArraySerializer` return
  `binco.Serializer[T]`.
- Generated DTM constants carry literal values; duplicate, negative or mixed
  pinned/automatic ids fail generation.
- A nil interface value is written as the reserved identifier `0` and decodes
  back to nil instead of panicking.
- Generator functions return errors instead of panicking, and `Generate` wraps
  the real cause with `%w`.
- `Marshal` panics with one sentinel (`ErrBufferTooSmall`) from every codec; a decoder that runs out of input reports `ErrTruncated`.
- `ptrmap.Wrap` clears the maps in place instead of reallocating them.
- Anonymous serializer names use truncated SHA-256 instead of MD5.
- Package `zerocpy` is now `zerocopy` (`wirefmt/zerocpy` is `wirefmt/zerocopy`).
- The `NewValid*` constructors are gone: every constructor honours the validators
  passed as options, so there is one `New*` name per type.
- `binco.ErrTooSmallByteSlice` is split into `binco.ErrTruncated` (input that
  ends inside a value) and `binco.ErrBufferTooSmall` (a short output buffer), so
  the two failures are distinguishable.
- `binco.DecodeExact` returns `(T, error)` like `binco.Decode`, and
  `AppendEncode`/`AppendMarshal`/`AppendMarshalTyped` take the destination first,
  like `append`.
- `binco.UintMaxVarintLen` and `binco.UintMaxLastByte` are gone: `int` and `uint`
  always use the `uint64` limits. `varnum.IntSerializer` is now `intSerializer`.
- The docs are split by audience: `README.md` explains how to use the library,
  `AGENTS.md` explains how it is built.

### Fixed

- Field options were applied to the wrong field when an earlier field was
  ignored.
- A second type whose bare name matched an already generated one silently reused
  its serializer; it is now a `ErrNameCollision` error.
- `ArrayPattern` is anchored, so `int32` and `Foo1Bar` are no longer classified
  as arrays.
- A truncated versioned payload reported fewer consumed bytes than the
  equivalent interface payload.
- `Skip` accepted values that `Unmarshal` rejected as out of range.
- The streaming sorted-map serializer swallowed an element codec error and wrote
  a shifted payload; it now returns the error, or `ErrWrongSize` when the codec
  disagrees with its `Size`.
- The buffer sorted-map serializer silently truncated an entry in a short buffer;
  it now panics with `ErrBufferTooSmall`, and with `ErrWrongSize` when an element
  codec under-writes. The `Append*` helpers check the same contract.
- Duplicated error constructors in `codegen/spec/converter` and unused ones in
  `codegen/typename`/`codegen/spec/builder` were removed.

### Removed

- The unused `Generator.scratch` field.
- `binco.AllocHintElem`, which no codec used.

## [1.0.0] - 2026-10-25

The first stable release. The wire format, the public API and the generated
code are frozen for the v1 series.

### Added

- Buffer codecs for every built-in Go type: booleans, bytes, signed and
  unsigned integers of every width, floats, strings, byte slices, arrays,
  slices, maps, pointers, time values, defined types and interfaces.
- Variable-length (`varnum`), fixed-width (`scalar`) and zero-copy (`zerocopy`)
  codecs with identical `Marshal`, `Unmarshal`, `Size` and `Skip` contracts.
- Streaming API (`wirefmt`) with the same byte format as the buffer API.
- Pointer mapping (`ptrmap`) for shared references, cycles and graphs.
- Type identifiers (`typeid`) for polymorphism and versioned payloads.
- Validators for lengths, keys, values, elements and whole values.
- Code generator (`codegen`) for structs, defined types, interfaces, versioned
  types and streaming codecs, including multi-package generation.
- Zero-copy decoding of strings, byte slices and arrays.
- Runnable examples in `demos/` and a complete reference in `README.md`.
- Fuzz targets for varint decoding and skipping, length-prefixed strings, byte
  slices, slices and maps, and the code-generation scanner. They are excluded
  from `go test` by the `fuzz` build tag and run in CI.
- Continuous integration: formatting, vet, golangci-lint, race-enabled tests,
  short fuzz runs and coverage uploaded to Coveralls on every push.
- A single CI workflow that publishes a GitHub release for every `v*` tag.

[Unreleased]: https://github.com/malivvan/binco/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/malivvan/binco/releases/tag/v1.0.0
