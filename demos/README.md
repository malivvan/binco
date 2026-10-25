# binco demos

Runnable examples for every part of `binco`. Each directory is a `main` package
inside the module, so an example is started with `go run` and always compiles
against the current library:

```bash
go run ./demos/types/struct
go run ./demos/typed/stream
```

The output below each example is included in the group READMEs so you can check
that what you see matches what is documented.

## Index

| Example | What it shows |
|---------|---------------|
| [`types/array`](types/array) | arrays; the length in the payload must match the array type |
| [`types/slice`](types/slice) | slices; the length serializer option; nil vs. empty |
| [`types/map`](types/map) | maps; entry order is not deterministic, decoding is |
| [`types/struct`](types/struct) | a hand-written struct serializer, `Skip`, `binco.Marshal` |
| [`validation/string`](validation/string) | validating a length while unmarshalling, `errors.Is` |
| [`validation/slice`](validation/slice) | length and element validators |
| [`validation/map`](validation/map) | length, key and value validators |
| [`out_of_order`](out_of_order) | using `Skip` to decode fields in reverse order |
| [`typed/different`](typed/different) | DTMs: one stream, unrelated types |
| [`typed/oneof`](typed/oneof) | interfaces: marshalling via `binco.MarshallerTyped`, DTM dispatch |
| [`typed/versioning`](typed/versioning) | several versions of a type and migration of old payloads |
| [`typed/stream`](typed/stream) | **streaming typed messages (DTS)**: DTM dispatch over a connection |
| [`pm/simple`](pm/simple) | pointer mapping: one pointer used by several fields |
| [`pm/linked_list`](pm/linked_list) | a doubly linked list, i.e. a cyclic structure |
| [`pm/graph`](pm/graph) | a cyclic graph: vertex → edge → vertex |
| [`native`](native) | zero-copy strings and byte slices, and their pitfalls |
| [`marshal_func`](marshal_func) | `binco.Marshal` / `binco.MarshalTyped` shorthands |
| [`gen`](gen) | code generation for your own types (`go generate`) |
| [`gen_options`](gen_options) | per-field generator options: custom name, raw numbers, validators |

## Reading order

If you are new to `binco`:

1. [`types/struct`](types/struct) - what a serializer is and what the wire
   format looks like for a struct.
2. [`types/slice`](types/slice) and [`types/map`](types/map) - containers.
3. [`typed/different`](typed/different) - how a DTM lets one stream carry
   different types.
4. [`typed/stream`](typed/stream) - the same idea over a network connection.
5. [`pm/graph`](pm/graph) - what pointer mapping adds for cyclic data.
6. [`codegen`](gen) - how to stop writing serializers by hand.

## Regenerating the generated code

Two examples contain generated code, which is committed so that `go run` works
without any extra step. After changing their types or generator options,
regenerate it:

```bash
go generate ./demos/gen ./demos/gen_options
```

`go generate` runs `go run ./gen` inside each example directory (see the
`//go:generate` directive in its `main.go`), which rewrites
`pkg/binco.gen.go`.

## Notes

* The examples print intermediate values (`size`, `encoded`, `decoded`) so the
  format is visible. In production code you would normally just marshal and
  unmarshal.
* Map iteration order is not deterministic, so examples that print a map render
  its keys in sorted order.
* The `zerocopy` example deliberately demonstrates a misuse (keeping a decoded
  string while the input buffer is reused) next to the correct alternatives.
