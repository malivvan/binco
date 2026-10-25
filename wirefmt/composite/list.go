package composite

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewSliceSer returns a slice serializer for elemSer.
func NewSliceSer[T any](elemSer wirefmt.Serializer[T], opts ...slopts.SetOption[T]) (
	s listCodec[T],
) {
	o := slopts.Options[T]{}
	slopts.Apply(opts, &o)

	return newListCodec(elemSer, o)
}

// NewValidSliceSer returns a slice serializer that runs its length and element validators while decoding.
func NewValidSliceSer[T any](elemSer wirefmt.Serializer[T],
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

func newListCodec[T any](elemSer wirefmt.Serializer[T], o slopts.Options[T]) (
	s listCodec[T],
) {
	var lenSer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return listCodec[T]{
		elemSer: elemSer,
		lenSer:  lenSer,
	}
}

type listCodec[T any] struct {
	lenSer  wirefmt.Serializer[int]
	elemSer wirefmt.Serializer[T]
}

func (s listCodec[T]) Marshal(v []T, w wirefmt.Writer) (n int, err error) {
	return MarshalSlice(v, s.elemSer, s.lenSer, w)
}

func (s listCodec[T]) Unmarshal(r wirefmt.Reader) (v []T, n int, err error) {
	return UnmarshalSlice(s.elemSer, s.lenSer, r)
}

func (s listCodec[T]) Size(v []T) (size int) {
	return SizeSlice(v, s.elemSer, s.lenSer)
}

func (s listCodec[T]) Skip(r wirefmt.Reader) (
	n int, err error,
) {
	return SkipSlice(s.elemSer, s.lenSer, r)
}

type checkedList[T any] struct {
	listCodec[T]
	lenVl  binco.Validator[int]
	elemVl binco.Validator[T]
}

func (s checkedList[T]) Unmarshal(r wirefmt.Reader) (v []T, n int, err error) {
	return UnmarshalValidSlice(s.elemSer, s.lenSer, s.lenVl, s.elemVl, r)
}

// MarshalSlice writes v, prefixed by its length, to w.
func MarshalSlice[T any](v []T, elemSer wirefmt.Serializer[T],
	lenSer wirefmt.Serializer[int], w wirefmt.Writer) (n int, err error) {
	n, err = lenSer.Marshal(len(v), w)
	if err != nil {
		return
	}
	var n1 int
	for _, e := range v {
		n1, err = elemSer.Marshal(e, w)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalSlice reads a slice from r.
func UnmarshalSlice[T any](elemSer wirefmt.Serializer[T],
	lenSer wirefmt.Serializer[int], r wirefmt.Reader) (v []T, n int, err error) {
	length, n, err := lenSer.Unmarshal(r)
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
		e, n1, err = elemSer.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		v[i] = e
	}
	return
}

// SizeSlice reports the encoded size of v.
func SizeSlice[T any](v []T, elemSer wirefmt.Serializer[T],
	lenSer wirefmt.Serializer[int]) (size int) {
	size = lenSer.Size(len(v))
	for i := range v {
		size += elemSer.Size(v[i])
	}
	return
}

// SkipSlice advances over an encoded slice.
func SkipSlice[T any](elemSer wirefmt.Serializer[T],
	lenSer wirefmt.Serializer[int], r wirefmt.Reader) (n int, err error) {
	length, n, err := lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int
	for range length {
		n1, err = elemSer.Skip(r)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalValidSlice reads a slice and validates its length and elements.
func UnmarshalValidSlice[T any](elemSer wirefmt.Serializer[T],
	lenSer wirefmt.Serializer[int], lenVl binco.Validator[int],
	elemVl binco.Validator[T], r wirefmt.Reader) (v []T, n int, err error) {
	length, n, err := lenSer.Unmarshal(r)
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
		e, n1, err = elemSer.Unmarshal(r)
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
