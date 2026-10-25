package composite

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewSliceSerializer returns a slice serializer that encodes elements with elemSerializer.
func NewSliceSerializer[T any](elemSerializer wirefmt.Serializer[T], opts ...slopts.SetOption[T]) (
	s listCodec[T],
) {
	o := slopts.Options[T]{}
	slopts.Apply(opts, &o)

	return newListCodec(elemSerializer, o)
}

// NewValidSliceSerializer returns a slice serializer that runs its length and element validators while decoding.
func NewValidSliceSerializer[T any](elemSerializer wirefmt.Serializer[T],
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

func newListCodec[T any](elemSerializer wirefmt.Serializer[T], o slopts.Options[T]) (
	s listCodec[T],
) {
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return listCodec[T]{
		elemSerializer: elemSerializer,
		lenSerializer:  lenSerializer,
	}
}

type listCodec[T any] struct {
	lenSerializer  wirefmt.Serializer[int]
	elemSerializer wirefmt.Serializer[T]
}

func (s listCodec[T]) Marshal(v []T, w wirefmt.Writer) (n int, err error) {
	return MarshalSlice(v, s.elemSerializer, s.lenSerializer, w)
}

func (s listCodec[T]) Unmarshal(r wirefmt.Reader) (v []T, n int, err error) {
	return UnmarshalSlice(s.elemSerializer, s.lenSerializer, r)
}

func (s listCodec[T]) Size(v []T) (size int) {
	return SizeSlice(v, s.elemSerializer, s.lenSerializer)
}

func (s listCodec[T]) Skip(r wirefmt.Reader) (
	n int, err error,
) {
	return SkipSlice(s.elemSerializer, s.lenSerializer, r)
}

type checkedList[T any] struct {
	listCodec[T]
	lenVl  binco.Validator[int]
	elemVl binco.Validator[T]
}

func (s checkedList[T]) Unmarshal(r wirefmt.Reader) (v []T, n int, err error) {
	return UnmarshalValidSlice(s.elemSerializer, s.lenSerializer, s.lenVl, s.elemVl, r)
}

// MarshalSlice writes v, prefixed by its length, to w.
func MarshalSlice[T any](v []T, elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int], w wirefmt.Writer) (n int, err error) {
	n, err = lenSerializer.Marshal(len(v), w)
	if err != nil {
		return
	}
	var n1 int
	for _, e := range v {
		n1, err = elemSerializer.Marshal(e, w)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalSlice reads a slice from r.
func UnmarshalSlice[T any](elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int], r wirefmt.Reader) (v []T, n int, err error) {
	length, n, err := lenSerializer.Unmarshal(r)
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
		e, n1, err = elemSerializer.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		v[i] = e
	}
	return
}

// SizeSlice reports the encoded size of v.
func SizeSlice[T any](v []T, elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int]) (size int) {
	size = lenSerializer.Size(len(v))
	for i := range v {
		size += elemSerializer.Size(v[i])
	}
	return
}

// SkipSlice advances over an encoded slice.
func SkipSlice[T any](elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int], r wirefmt.Reader) (n int, err error) {
	length, n, err := lenSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int
	for range length {
		n1, err = elemSerializer.Skip(r)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalValidSlice reads a slice and validates its length and elements.
func UnmarshalValidSlice[T any](elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int], lenVl binco.Validator[int],
	elemVl binco.Validator[T], r wirefmt.Reader) (v []T, n int, err error) {
	length, n, err := lenSerializer.Unmarshal(r)
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
		e, n1, err = elemSerializer.Unmarshal(r)
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
