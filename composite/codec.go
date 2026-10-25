package composite

import "github.com/malivvan/binco"

// Bool serializes a bool as a single byte.
var Bool = boolCodec{}

type boolCodec struct{}

func (s boolCodec) Marshal(v bool, bs []byte) (n int) {
	if v {
		bs[0] = 1
	} else {
		bs[0] = 0
	}
	return 1
}

func (s boolCodec) Unmarshal(bs []byte) (v bool, n int, err error) {
	if len(bs) < 1 {
		return false, 0, binco.ErrTooSmallByteSlice
	}
	if bs[0] > 1 {
		return false, 0, binco.ErrWrongFormat
	}
	return bs[0] == 1, 1, nil
}

func (s boolCodec) Size(_ bool) (size int) {
	return 1
}

func (s boolCodec) Skip(bs []byte) (n int, err error) {
	return SkipBool(bs)
}

// SkipBool advances over an encoded bool.
func SkipBool(bs []byte) (n int, err error) {
	if len(bs) < 1 {
		return 0, binco.ErrTooSmallByteSlice
	}
	if bs[0] == 0 || bs[0] == 1 {
		return 1, nil
	}
	return 0, binco.ErrWrongFormat
}

// NewPtrSer returns a pointer serializer that encodes nil as a flag byte.
func NewPtrSer[T any](baseSer binco.Serializer[T]) ptrCodec[T] {
	return ptrCodec[T]{baseSer}
}

type ptrCodec[T any] struct {
	baseSer binco.Serializer[T]
}

func (s ptrCodec[T]) Marshal(v *T, bs []byte) (n int) {
	if v == nil {
		bs[0] = byte(binco.PtrNil)
		n = 1
		return
	}
	bs[0] = byte(binco.PtrNotNil)
	return 1 + s.baseSer.Marshal(*v, bs[1:])
}

func (s ptrCodec[T]) Unmarshal(bs []byte) (v *T, n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTooSmallByteSlice
		return
	}
	if bs[0] == byte(binco.PtrNil) {
		n = 1
		return
	}
	if bs[0] != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	k, n, err := s.baseSer.Unmarshal(bs[1:])
	if err != nil {
		n = 1 + n
		return
	}
	return &k, 1 + n, err
}

func (s ptrCodec[T]) Size(v *T) (size int) {
	if v != nil {
		return 1 + s.baseSer.Size(*v)
	}
	return 1
}

func (s ptrCodec[T]) Skip(bs []byte) (n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTooSmallByteSlice
		return
	}
	if bs[0] == byte(binco.PtrNil) {
		n = 1
		return
	}
	if bs[0] != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	n, err = s.baseSer.Skip(bs[1:])
	return 1 + n, err
}
