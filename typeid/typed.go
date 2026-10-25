package typeid

import "github.com/malivvan/binco"

// Ser pairs a serializer with the type identifier written in front of its payloads.
type Ser[T any] struct {
	dtm binco.DTM
	ser binco.Serializer[T]
}

// NewSer returns a typed serializer for dtm that delegates the data to ser.
func NewSer[T any](dtm binco.DTM, ser binco.Serializer[T]) Ser[T] {
	return Ser[T]{dtm, ser}
}

// DTM returns the identifier the serializer writes.
func (d Ser[T]) DTM() binco.DTM {
	return d.dtm
}

// Marshal writes the identifier followed by the value.
func (d Ser[T]) Marshal(t T, bs []byte) (n int) {
	n = DTMSer.Marshal(d.dtm, bs)
	n += d.ser.Marshal(t, bs[n:])
	return
}

// Unmarshal reads the identifier, checks it and decodes the value.
func (d Ser[T]) Unmarshal(bs []byte) (t T, n int, err error) {
	dtm, n, err := DTMSer.Unmarshal(bs)
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
func (d Ser[T]) Size(t T) (size int) {
	size = DTMSer.Size(d.dtm)
	return size + d.ser.Size(t)
}

// Skip advances over the identifier and the value.
func (d Ser[T]) Skip(bs []byte) (n int, err error) {
	dtm, n, err := DTMSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if dtm != d.dtm {
		err = binco.NewWrongDTMError(d.dtm, dtm)
		return
	}
	var n1 int
	n1, err = d.ser.Skip(bs[n:])
	n += n1
	return
}

// UnmarshalData decodes the value, assuming the identifier has already been read.
func (d Ser[T]) UnmarshalData(bs []byte) (t T, n int, err error) {
	return d.ser.Unmarshal(bs)
}

// SkipData advances over the value, assuming the identifier has already been read.
func (d Ser[T]) SkipData(bs []byte) (n int, err error) {
	return d.ser.Skip(bs)
}
