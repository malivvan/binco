package typeid

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Ser pairs a streaming serializer with the type identifier written in front of its payloads.
type Ser[T any] struct {
	dtm binco.DTM
	ser wirefmt.Serializer[T]
}

// NewSer returns a typed serializer for dtm that delegates the data to ser.
func NewSer[T any](dtm binco.DTM, ser wirefmt.Serializer[T]) Ser[T] {
	return Ser[T]{dtm, ser}
}

// DTM returns the identifier the serializer writes.
func (s Ser[T]) DTM() binco.DTM {
	return s.dtm
}

// Marshal writes the identifier followed by the value.
func (s Ser[T]) Marshal(v T, w wirefmt.Writer) (n int, err error) {
	n, err = DTMSer.Marshal(s.dtm, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.ser.Marshal(v, w)
	n += n1
	return
}

// Unmarshal reads the identifier, checks it and decodes the value.
func (s Ser[T]) Unmarshal(r wirefmt.Reader) (v T, n int, err error) {
	dtm, n, err := DTMSer.Unmarshal(r)
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
func (s Ser[T]) Size(v T) (size int) {
	size = DTMSer.Size(s.dtm)
	return size + s.ser.Size(v)
}

// Skip advances over the identifier and the value.
func (s Ser[T]) Skip(r wirefmt.Reader) (n int, err error) {
	dtm, n, err := DTMSer.Unmarshal(r)
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
func (s Ser[T]) UnmarshalData(r wirefmt.Reader) (v T, n int, err error) {
	return s.ser.Unmarshal(r)
}

// SkipData advances over the value, assuming the identifier has already been read.
func (s Ser[T]) SkipData(r wirefmt.Reader) (n int, err error) {
	return s.ser.Skip(r)
}
