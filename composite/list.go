package composite

import (
	"github.com/malivvan/binco"
	slopts "github.com/malivvan/binco/config/slice"
	"github.com/malivvan/binco/varnum"
)

// NewSliceSer returns a slice serializer for elemSer.
func NewSliceSer[T any](elemSer binco.Serializer[T], opts ...slopts.SetOption[T]) (
	s listCodec[T],
) {
	o := slopts.Options[T]{}
	slopts.Apply(opts, &o)

	return newListCodec(elemSer, o)
}

// NewValidSliceSer returns a slice serializer that runs its length and element validators while decoding.
func NewValidSliceSer[T any](elemSer binco.Serializer[T],
	opts ...slopts.SetOption[T],
) checkedList[T] {
	o := slopts.Options[T]{}
	slopts.Apply(opts, &o)

	var lenVl binco.Validator[int]

	var elemVl binco.Validator[T]

	if o.LenVl != nil {
		lenVl = o.LenVl
	}
	if o.ElemVl != nil {
		elemVl = o.ElemVl
	}
	return checkedList[T]{
		listCodec: newListCodec(elemSer, o),
		lenVl:     lenVl,
		elemVl:    elemVl,
	}
}

func newListCodec[T any](elemSer binco.Serializer[T], o slopts.Options[T]) (
	s listCodec[T],
) {
	var lenSer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return listCodec[T]{
		ElemSer: elemSer,
		LenSer:  lenSer,
	}
}

type listCodec[T any] struct {
	LenSer  binco.Serializer[int]
	ElemSer binco.Serializer[T]
}

func (s listCodec[T]) Marshal(v []T, bs []byte) (n int) {
	return MarshalSlice(v, s.ElemSer, s.LenSer, bs)
}

func (s listCodec[T]) Unmarshal(bs []byte) (v []T, n int, err error) {
	return UnmarshalSlice(bs, s.ElemSer, s.LenSer)
}

func (s listCodec[T]) Size(v []T) (size int) {
	return SizeSlice(v, s.ElemSer, s.LenSer)
}

func (s listCodec[T]) Skip(bs []byte) (n int, err error) {
	return SkipSlice(bs, s.ElemSer, s.LenSer)
}

type checkedList[T any] struct {
	listCodec[T]
	lenVl  binco.Validator[int]
	elemVl binco.Validator[T]
}

func (s checkedList[T]) Unmarshal(bs []byte) (v []T, n int, err error) {
	return UnmarshalValidSlice(bs, s.ElemSer, s.LenSer, s.lenVl, s.elemVl)
}

// MarshalSlice writes v, prefixed by its length, into bs.
func MarshalSlice[T any](v []T, elemSer binco.Serializer[T],
	lenSer binco.Serializer[int], bs []byte,
) (n int) {
	n = lenSer.Marshal(len(v), bs)
	for _, e := range v {
		n += elemSer.Marshal(e, bs[n:])
	}
	return
}

// UnmarshalSlice decodes a slice from bs.
func UnmarshalSlice[T any](bs []byte, elemSer binco.Serializer[T],
	lenSer binco.Serializer[int],
) (v []T, n int, err error) {
	length, n, err := lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int

	var e T

	v = make([]T, length)
	for i := range length {
		e, n1, err = elemSer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		v[i] = e
	}
	return
}

// SizeSlice reports the encoded size of v.
func SizeSlice[T any](v []T, elemSer binco.Serializer[T],
	lenSer binco.Serializer[int],
) (size int) {
	length := len(v)
	size = lenSer.Size(length)
	for i := range length {
		size += elemSer.Size(v[i])
	}
	return
}

// SkipSlice advances over an encoded slice.
func SkipSlice[T any](bs []byte, elemSer binco.Serializer[T],
	lenSer binco.Serializer[int],
) (n int, err error) {
	length, n, err := lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int
	for range length {
		n1, err = elemSer.Skip(bs[n:])
		n += n1
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalValidSlice decodes a slice and validates its length and elements.
func UnmarshalValidSlice[T any](bs []byte, elemSer binco.Serializer[T],
	lenSer binco.Serializer[int], lenVl binco.Validator[int],
	elemVl binco.Validator[T],
) (v []T, n int, err error) {
	length, n, err := lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	if lenVl != nil {
		if err = lenVl.Validate(length); err != nil {
			return
		}
	}
	var n1 int

	var e T

	v = make([]T, length)
	for i := range length {
		e, n1, err = elemSer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		if elemVl != nil {
			if err = elemVl.Validate(e); err != nil {
				return
			}
		}
		v[i] = e
	}
	return
}
