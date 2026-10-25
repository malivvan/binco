package wirefmt

import (
	"io"
)

// Writer is the destination of a streaming codec: it can write a byte, a byte slice and a string.
type Writer interface {
	io.ByteWriter
	io.Writer
	io.StringWriter
}

// Reader is the source of a streaming codec: it can read a byte and a byte slice.
type Reader interface {
	io.ByteReader
	io.Reader
}

// Serializer is the streaming counterpart of binco.Serializer. Every method reports the bytes processed and an error.
type Serializer[T any] interface {
	Marshal(t T, w Writer) (n int, err error)
	Unmarshal(r Reader) (t T, n int, err error)
	Size(t T) (size int)
	Skip(r Reader) (n int, err error)
}

// Marshaller lets a value write itself to a wire.
type Marshaller interface {
	MarshalBinco(w Writer) (n int, err error)
	SizeBinco() (size int)
}

// MarshallerTyped is Marshaller with a type identifier in front of the payload.
type MarshallerTyped interface {
	MarshalTypedBinco(w Writer) (n int, err error)
	SizeTypedBinco() (size int)
}
