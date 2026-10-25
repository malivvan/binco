# The marshalling shorthands

```
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

Passing a serializer to every `Marshal` call is explicit, but sometimes a value
should just serialize itself. Two interfaces cover that:

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

```go
bs := binco.Marshal(foo)       // allocates SizeBinco() bytes, then MarshalBinco
bs = binco.MarshalTyped(foo)   // the same, with the DTM prefix
```

`Foo` in this example implements both; `Bar` implements only `Marshaller`, so
`binco.Marshal(bar)` works while `binco.MarshalTyped(bar)` would not compile.

Decoding has no shorthand, on purpose: the payload does not carry its type, so
the caller must name the serializer (`binco.Decode(FooSerializer, bs)`) or read the DTM
first (`typeid.DTMSerializer.Unmarshal` plus dispatch, see
[typed/different](../typed/different)).

The methods are named `*Binco` rather than `Marshal`/`Size` so that a type can
implement several serialization conventions at once (`encoding/json`,
`encoding/binary`, ...) without method collisions.
