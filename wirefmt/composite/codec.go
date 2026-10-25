package composite

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Bool serializes a bool as a single byte.
var Bool = boolCodec{}

type boolCodec struct{}

func (boolCodec) Marshal(v bool, w wirefmt.Writer) (n int, err error) {
	if v {
		err = w.WriteByte(1)
	} else {
		err = w.WriteByte(0)
	}
	if err != nil {
		return
	}
	n = 1
	return
}

func (boolCodec) Unmarshal(r wirefmt.Reader) (v bool, n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n = 1
	if b > 1 {
		err = binco.ErrWrongFormat
		return
	}
	v = b == 1
	return
}

func (boolCodec) Size(v bool) (size int) {
	return SizeBool(v)
}

func (boolCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return SkipBool(r)
}

// SizeBool reports the encoded size of a bool.
func SizeBool(v bool) (size int) {
	return 1
}

// SkipBool advances over an encoded bool.
func SkipBool(r wirefmt.Reader) (n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n = 1
	if b > 1 {
		err = binco.ErrWrongFormat
	}
	return
}

// NewPtrSer returns a pointer serializer that encodes nil as a flag byte.
func NewPtrSer[T any](baseSer wirefmt.Serializer[T]) ptrCodec[T] {
	return ptrCodec[T]{baseSer}
}

type ptrCodec[T any] struct {
	baseSer wirefmt.Serializer[T]
}

func (s ptrCodec[T]) Marshal(v *T, w wirefmt.Writer) (n int, err error) {
	if v == nil {
		err = w.WriteByte(byte(binco.PtrNil))
		if err != nil {
			return
		}
		n++
		return
	}
	if err = w.WriteByte(byte(binco.PtrNotNil)); err != nil {
		return
	}
	n, err = s.baseSer.Marshal(*v, w)
	n += 1
	return
}

func (s ptrCodec[T]) Unmarshal(r wirefmt.Reader) (v *T, n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n++
	if b == byte(binco.PtrNil) {
		return
	}
	if b != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	var n1 int
	k, n1, err := s.baseSer.Unmarshal(r)
	n += n1
	if err != nil {
		return
	}
	v = &k
	return
}

func (s ptrCodec[T]) Size(v *T) (size int) {
	if v != nil {
		return 1 + s.baseSer.Size(*v)
	}
	return 1
}

func (s ptrCodec[T]) Skip(r wirefmt.Reader) (n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n++
	if b == byte(binco.PtrNil) {
		return
	}
	if b != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	n, err = s.baseSer.Skip(r)
	n += 1
	return
}
