package typeid

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Serializer pairs a streaming serializer with the type identifier written in front of its payloads.
type Serializer[T any] struct {
	dtm        binco.DTM
	serializer wirefmt.Serializer[T]
}

// NewSerializer returns a typed serializer for dtm that delegates the data to serializer.
func NewSerializer[T any](dtm binco.DTM, serializer wirefmt.Serializer[T]) Serializer[T] {
	return Serializer[T]{dtm, serializer}
}

// DTM returns the identifier the serializer writes.
func (s Serializer[T]) DTM() binco.DTM {
	return s.dtm
}

// Marshal writes the identifier followed by the value.
func (s Serializer[T]) Marshal(v T, w wirefmt.Writer) (n int, err error) {
	n, err = DTMSerializer.Marshal(s.dtm, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.serializer.Marshal(v, w)
	n += n1
	return
}

// Unmarshal reads the identifier, checks it and decodes the value.
func (s Serializer[T]) Unmarshal(r wirefmt.Reader) (v T, n int, err error) {
	dtm, n, err := DTMSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if dtm != s.dtm {
		err = binco.NewWrongDTMError(s.dtm, dtm)
		return
	}
	var n1 int
	v, n1, err = s.UnmarshalData(r)
	n += n1
	return
}

// Size reports the encoded size of the identifier plus the value.
func (s Serializer[T]) Size(v T) (size int) {
	size = DTMSerializer.Size(s.dtm)
	return size + s.serializer.Size(v)
}

// Skip advances over the identifier and the value.
func (s Serializer[T]) Skip(r wirefmt.Reader) (n int, err error) {
	dtm, n, err := DTMSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if dtm != s.dtm {
		err = binco.NewWrongDTMError(s.dtm, dtm)
		return
	}
	n1, err := s.SkipData(r)
	n += n1
	return
}

// UnmarshalData decodes the value, assuming the identifier has already been read.
func (s Serializer[T]) UnmarshalData(r wirefmt.Reader) (v T, n int, err error) {
	return s.serializer.Unmarshal(r)
}

// SkipData advances over the value, assuming the identifier has already been read.
func (s Serializer[T]) SkipData(r wirefmt.Reader) (n int, err error) {
	return s.serializer.Skip(r)
}
