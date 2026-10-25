package composite

import (
	"math"
	"unsafe"

	"github.com/malivvan/binco"
	slopts "github.com/malivvan/binco/config/slice"
	"github.com/malivvan/binco/internal/fixedwidth"
	"github.com/malivvan/binco/varnum"
)

// NewSliceSerializer returns a slice serializer that encodes elements with elemSerializer.
//
// Length and element validators passed through opts are honoured; there is no
// separate constructor to remember.
func NewSliceSerializer[T any](elemSerializer binco.Serializer[T], opts ...slopts.SetOption[T]) binco.Serializer[[]T] {
	o := slopts.Options[T]{}
	slopts.Apply(opts, &o)

	s := newListCodec(elemSerializer, o)
	if o.LenVl != nil || o.ElemVl != nil {
		return checkedList[T]{listCodec: s, lenVl: o.LenVl, elemVl: o.ElemVl}
	}
	return s
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
	if width, ok := fixedwidth.Of[T](elemSerializer); ok {
		if len(v) == 0 {
			return n
		}
		if len(v) <= math.MaxInt/width {
			total := len(v) * width
			binco.CheckWrite(bs, n, total)
			src := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(v))), total)
			copy(bs[n:], src)
			return n + total
		}
	}
	for _, e := range v {
		binco.CheckWrite(bs, n, 0)
		n += elemSerializer.Marshal(e, bs[n:])
	}
	return
}

// UnmarshalSlice decodes a slice from bs.
//
// The backing array is grown with the elements as they are decoded rather than
// allocated from the decoded length, so a five-byte input cannot ask for a
// multi-gigabyte allocation. A length whose elements consume no bytes at all is
// capped by binco.MaxZeroSizeElements.
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
	if width, ok := fixedwidth.Of[T](elemSerializer); ok {
		return unmarshalFixedSlice[T](bs, n, length, width)
	}
	// Elements take at least one byte each; allocating more than the input can
	// hold would be an amplification attack.
	v = make([]T, 0, binco.AllocHint(length, binco.Remaining(bs, n)))
	var e T

	var n1 int

	for range length {
		e, n1, err = elemSerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		v = append(v, e)
		if n1 == 0 {
			// The element consumed nothing: every remaining element would
			// decode from the same input and return the same value, so fill the
			// rest without looping once per element.
			if length > binco.MaxZeroSizeElements {
				return nil, n, binco.ErrTooLargeLength
			}
			for len(v) < length {
				v = append(v, e)
			}
			return
		}
	}
	return
}

// unmarshalFixedSlice decodes length elements of a fixed width in one copy.
func unmarshalFixedSlice[T any](bs []byte, n, length, width int) (v []T, end int, err error) {
	if err = binco.CheckLengthElem(bs, n, length, width); err != nil {
		return nil, n, err
	}
	v = make([]T, length)
	if length == 0 {
		return v, n, nil
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(v))), length*width)
	copy(dst, bs[n:n+length*width])
	return v, n + length*width, nil
}

// SizeSlice reports the encoded size of v.
func SizeSlice[T any](v []T, elemSerializer binco.Serializer[T],
	lenSerializer binco.Serializer[int],
) (size int) {
	length := len(v)
	size = lenSerializer.Size(length)
	if width, ok := fixedwidth.Of[T](elemSerializer); ok && length <= math.MaxInt/width {
		return size + length*width
	}
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
	if width, ok := fixedwidth.Of[T](elemSerializer); ok {
		if err = binco.CheckLengthElem(bs, n, length, width); err != nil {
			return n, err
		}
		return n + length*width, nil
	}
	var n1 int
	for range length {
		n1, err = elemSerializer.Skip(bs[n:])
		n += n1
		if err != nil {
			return
		}
		if n1 == 0 {
			if length > binco.MaxZeroSizeElements {
				return n, binco.ErrTooLargeLength
			}
			// Nothing more can be consumed, so the remaining elements are
			// immediately skipped.
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
	v = make([]T, 0, binco.AllocHint(length, binco.Remaining(bs, n)))
	var e T

	var n1 int

	for range length {
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
		v = append(v, e)
		if n1 == 0 {
			if length > binco.MaxZeroSizeElements {
				return nil, n, binco.ErrTooLargeLength
			}
			for len(v) < length {
				v = append(v, e)
			}
			return
		}
	}
	return
}
