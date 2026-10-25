# Typed serialization

Data type metadata (DTM) is a positive integer that is written in front of a
payload and identifies its type:

```
DTM + data
```

With a DTM a single buffer or stream can carry values of different types, and a
receiver knows which serializer to use before it reads the data.

- [different](different/): unrelated types behind DTMs.
- [oneof](oneof/): a Go interface, serialized as DTM + the concrete value.
- [versioning](versioning/): several versions of one type, with migration.
- [stream](stream/): **receiving arbitrary typed messages from a stream (DTS)**.

## different

```
Foo: [1 10] (DTM 1)
Bar: [2 11 104 101 108 108 111 32 119 111 114 108 100] (DTM 2)

received Foo {Num:5}
received Bar {Str:hello world}
Bar reader got a Foo payload: expected 2, actual 1
DTM prefix: 1 (1 byte), data: [10]
```

`typeid.NewSer(dtm, ser)` returns a serializer that reads and writes `DTM + data`.
It implements `binco.Serializer`, so it composes like any other serializer.
`Unmarshal` checks the DTM and returns `binco.WrongDTMError` when the payload
belongs to another type; `UnmarshalData` reads only the data when the caller has
already consumed the DTM.

## oneof

```
marshaled main.Copy {Start:10 End:20} → [1 20 40] (3 bytes)
  unmarshaled main.Copy (read 3 bytes): copy from 10 to 20
marshaled main.Insert {Str:hello world} → [2 11 104 ... 100] (13 bytes)
  unmarshaled main.Insert (read 13 bytes): insert "hello world"

skipped 3 of 3 bytes
unknown instruction: unexpected instruction: binco: unexpected DTM 99
```

`InstructionSer` marshals through `binco.MarshallerTyped`, so the implementations
own their encoding and adding one requires no change to the interface
serializer. Unmarshalling needs the DTM → serializer table, because the concrete
type is unknown until the DTM has been read.

## versioning

```
V1 payload: [1 20] (2 bytes)
V2 payload: [2 7 99 117 114 114 101 110 116] (9 bytes)

decoded 2 bytes: "10"
decoded 9 bytes: "current"

marshaled "rewritten" → [2 9 114 101 119 114 105 116 116 101 110]
binco.Marshal:      [2 9 114 101 119 114 105 116 116 101 110]
binco.MarshalTyped: [1 20]
```

Every version has its own DTM. The current serializer writes the current version
and reads all of them, converting old payloads with a migration function, so the
rest of the program only deals with today's type. An unknown version is reported
(`binco.UnexpectedDTMError`) instead of being decoded as the current one.

## stream

See [stream/README.md](stream/README.md). It is the example to read if you want
to receive a sequence of differently typed messages over a network connection.

## Rules of thumb

* A DTM must be positive and must never be reused for a different type.
* Remove versions only when no stored payload can contain them any more.
* The generated code numbers DTMs automatically and continuously
  (`RegisterInterface`, `RegisterVersioned`), so they stay unique within one
  generated file.
