package typeid

import "github.com/malivvan/binco"

// Serializer pairs a serializer with the type identifier written in front of its payloads.
type Serializer[T any] struct {
	dtm        binco.DTM
	serializer binco.Serializer[T]
}

// NewSerializer returns a typed serializer for dtm that delegates the data to the wrapped serializer.
func NewSerializer[T any](dtm binco.DTM, serializer binco.Serializer[T]) Serializer[T] {
	return Serializer[T]{dtm, serializer}
}

// DTM returns the identifier the serializer writes.
func (d Serializer[T]) DTM() binco.DTM {
	return d.dtm
}

// Marshal writes the identifier followed by the value.
func (d Serializer[T]) Marshal(t T, bs []byte) (n int) {
	n = DTMSerializer.Marshal(d.dtm, bs)
	n += d.serializer.Marshal(t, bs[n:])
	return
}

// Unmarshal reads the identifier, checks it and decodes the value.
func (d Serializer[T]) Unmarshal(bs []byte) (t T, n int, err error) {
	dtm, n, err := DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	if dtm != d.dtm {
		err = binco.NewWrongDTMError(d.dtm, dtm)
		return
	}
	var n1 int
	t, n1, err = d.UnmarshalData(bs[n:])
	n += n1
	return
}

// Size reports the encoded size of the identifier plus the value.
func (d Serializer[T]) Size(t T) (size int) {
	size = DTMSerializer.Size(d.dtm)
	return size + d.serializer.Size(t)
}

// Skip advances over the identifier and the value.
func (d Serializer[T]) Skip(bs []byte) (n int, err error) {
	dtm, n, err := DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	if dtm != d.dtm {
		err = binco.NewWrongDTMError(d.dtm, dtm)
		return
	}
	var n1 int
	n1, err = d.serializer.Skip(bs[n:])
	n += n1
	return
}

// UnmarshalData decodes the value, assuming the identifier has already been read.
func (d Serializer[T]) UnmarshalData(bs []byte) (t T, n int, err error) {
	return d.serializer.Unmarshal(bs)
}

// SkipData advances over the value, assuming the identifier has already been read.
func (d Serializer[T]) SkipData(bs []byte) (n int, err error) {
	return d.serializer.Skip(bs)
}
