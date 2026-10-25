# Streaming typed messages (DTS)

```
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

This is the streaming counterpart of [typed/different](../different/): messages
of different types arrive one after another on a single connection, and each one
is written as `DTM + data`. Nothing else frames the stream - the schema supplies
the lengths.

## The wire format

```
login  →  [1] [len "alice"] ["alice"]
chat   →  [2] [len from] ["alice"] [len text] ["hello everyone"] [int64 unix seconds]
upload →  [3] [len name] ["report.pdf"] [len data] [1 2 3 4]
```

The receiver reads the DTM first and then hands the stream to the matching
serializer. Because every serializer reads exactly the bytes of its own value,
the stream is left positioned at the next DTM.

## Sending

```go
type typedMessage interface {
	writeTo(w wirefmt.Writer) (n int, err error)
}

func (m Login) writeTo(w wirefmt.Writer) (n int, err error) {
	return LoginTyped.Marshal(m, w)
}
```

The sender only needs this small interface, so a heterogeneous slice of messages
can be written in a loop. `LoginTyped` is a `wirefmt/typeid.Serializer[Login]`, created
once:

```go
var LoginTyped = wtypeid.NewSerializer(LoginDTM, loginSerializer{})
```

Each message is flushed after it is written (`bufio.Writer.Flush`), so the
receiver can process messages as they arrive instead of waiting for the end of
the wire.

## Receiving

```go
for {
	dtm, n, err := wtypeid.DTMSerializer.Unmarshal(r)
	if errors.Is(err, io.EOF) {
		return nil // the peer closed the connection between messages
	}
	...

	switch dtm {
	case LoginDTM:
		login, n, err := LoginTyped.UnmarshalData(r)
		...
	case UploadDTM:
		n, err := UploadTyped.SkipData(r) // known, but not needed
		...
	default:
		return fmt.Errorf("unknown message type: %w",
			binco.NewUnexpectedDTMError(dtm))
	}
}
```

Two details are worth copying:

* **Known but unwanted messages are skipped, not decoded.** `SkipData` advances
  over the payload without allocating it (16 bytes here). Use a *validating*
  serializer for untrusted input, because `Skip` does not validate.
* **An unknown DTM stops the reader.** Its length is unknown, so there is no way
  to resynchronize the stream; reporting `binco.UnexpectedDTMError` is the only
  correct reaction. To evolve a protocol, either append fields to a message type
  and bump nothing (unknown *fields* can be skipped if the writer's schema is
  known), or introduce a new type and keep the old readers out of that wire.

## Dispatch by table

The example uses a `switch`. When message types are registered at run time, the
same dispatch can be a map from DTM to a read function:

```go
var handlers = map[binco.DTM]func(wirefmt.Reader) (msg any, n int, err error){
	LoginDTM: func(r wirefmt.Reader) (any, int, error) {
		v, n, err := LoginTyped.UnmarshalData(r)
		return v, n, err
	},
	...
}

n, err := handlers[dtm](r)
```

The trade-off is the same as in [typed/oneof](../oneof/): a table can be
extended without recompiling the dispatcher, while a `switch` is checked by the
compiler.

## Why this works without framing

The MUS-style format has no message delimiters, no length prefix around a whole
message and no alignment. A receiver can therefore start reading in the middle of
a stream as long as it knows the schema of the next value - which is exactly what
the DTM tells it. This is what makes the format suitable for long-lived
connections: one DTM per message, no per-field metadata.
