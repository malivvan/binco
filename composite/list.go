package composite

import (
	"github.com/malivvan/binco"
	slopts "github.com/malivvan/binco/config/slice"
	"github.com/malivvan/binco/varnum"
)

// NewSliceSerializer returns a slice serializer that encodes elements with elemSerializer.
func NewSliceSerializer[T any](elemSerializer binco.Serializer[T], opts ...slopts.SetOption[T]) (
	s listCodec[T],
) {
	o := slopts.Options[T]{}
	slopts.Apply(opts, &o)

	return newListCodec(elemSerializer, o)
}

// NewValidSliceSerializer returns a slice serializer that runs its length and element validators while decoding.
func NewValidSliceSerializer[T any](elemSerializer binco.Serializer[T],
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
		listCodec: newListCodec(elemSerializer, o),
		lenVl:     lenVl,
		elemVl:    elemVl,
	}
}

func newListCodec[T any](elemSerializer binco.Serializer[T], o slopts.Options[T]) (
	s listCodec[T],
) {
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return listCodec[T]{
		ElemSerializer: elemSerializer,
		LenSerializer:  lenSerializer,
	}
}

type listCodec[T any] struct {
	LenSerializer  binco.Serializer[int]
	ElemSerializer binco.Serializer[T]
}

func (s listCodec[T]) Marshal(v []T, bs []byte) (n int) {
	return MarshalSlice(v, s.ElemSerializer, s.LenSerializer, bs)
}

func (s listCodec[T]) Unmarshal(bs []byte) (v []T, n int, err error) {
	return UnmarshalSlice(bs, s.ElemSerializer, s.LenSerializer)
}

func (s listCodec[T]) Size(v []T) (size int) {
	return SizeSlice(v, s.ElemSerializer, s.LenSerializer)
}

func (s listCodec[T]) Skip(bs []byte) (n int, err error) {
	return SkipSlice(bs, s.ElemSerializer, s.LenSerializer)
}

type checkedList[T any] struct {
	listCodec[T]
	lenVl  binco.Validator[int]
	elemVl binco.Validator[T]
}

func (s checkedList[T]) Unmarshal(bs []byte) (v []T, n int, err error) {
	return UnmarshalValidSlice(bs, s.ElemSerializer, s.LenSerializer, s.lenVl, s.elemVl)
}

// MarshalSlice writes v, prefixed by its length, into bs.
func MarshalSlice[T any](v []T, elemSerializer binco.Serializer[T],
	lenSerializer binco.Serializer[int], bs []byte,
) (n int) {
	n = lenSerializer.Marshal(len(v), bs)
	for _, e := range v {
		n += elemSerializer.Marshal(e, bs[n:])
	}
	return
}

// UnmarshalSlice decodes a slice from bs.
func UnmarshalSlice[T any](bs []byte, elemSerializer binco.Serializer[T],
	lenSerializer binco.Serializer[int],
) (v []T, n int, err error) {
	length, n, err := lenSerializer.Unmarshal(bs)
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
		e, n1, err = elemSerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		v[i] = e
	}
	return
}

// SizeSlice reports the encoded size of v.
func SizeSlice[T any](v []T, elemSerializer binco.Serializer[T],
	lenSerializer binco.Serializer[int],
) (size int) {
	length := len(v)
	size = lenSerializer.Size(length)
	for i := range length {
		size += elemSerializer.Size(v[i])
	}
	return
}

// SkipSlice advances over an encoded slice.
func SkipSlice[T any](bs []byte, elemSerializer binco.Serializer[T],
	lenSerializer binco.Serializer[int],
) (n int, err error) {
	length, n, err := lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int
	for range length {
		n1, err = elemSerializer.Skip(bs[n:])
		n += n1
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalValidSlice decodes a slice and validates its length and elements.
func UnmarshalValidSlice[T any](bs []byte, elemSerializer binco.Serializer[T],
	lenSerializer binco.Serializer[int], lenVl binco.Validator[int],
	elemVl binco.Validator[T],
) (v []T, n int, err error) {
	length, n, err := lenSerializer.Unmarshal(bs)
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
		e, n1, err = elemSerializer.Unmarshal(bs[n:])
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
