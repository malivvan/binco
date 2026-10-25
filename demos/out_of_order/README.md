# Out of order deserialization

```
payload: [84 12 111 117 116 32 111 102 32 111 114 100 101 114 210 4 0 0] (18 bytes)

offsets: ID ends at 1, Name ends at 14

score: 1234
name:  "out of order"
id:    42

only the score: 1234
truncated payload: binco: too small byte slice
```

A record is encoded as `ID`, `Name`, `Score`. `Skip` measures a value without
decoding it, so the offsets of the fields can be computed first:

```go
idEnd, _   := idSer.Skip(bs)          // end of ID
nameEnd, _ := nameSer.Skip(bs[idEnd:]) // end of Name

score, _, _ := scoreSer.Unmarshal(bs[idEnd+nameEnd:]) // last field first
name, _, _  := nameSer.Unmarshal(bs[idEnd:])
id, _, _    := idSer.Unmarshal(bs)
```

This is how a receiver can

* decode a field that was written last without decoding the fields before it,
* read only the field it needs and skip the rest without allocating.

The same works with the streaming API, but only forwards: a stream cannot seek,
so values are skipped to reach the next one. To read fields of one value in
reverse order from a stream, buffer that value first and use the buffer API.

Offsets are only as good as the payload: `Skip` fails on truncated or malformed
input instead of returning a wrong offset, which is what the last line of the
example shows.
