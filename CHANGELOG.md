# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-10-25

The first stable release. The wire format, the public API and the generated
code are frozen for the v1 series.

### Added

- Buffer codecs for every built-in Go type: booleans, bytes, signed and
  unsigned integers of every width, floats, strings, byte slices, arrays,
  slices, maps, pointers, time values, defined types and interfaces.
- Variable-length (`varnum`), fixed-width (`scalar`) and zero-copy (`zerocpy`)
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
