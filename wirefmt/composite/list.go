package composite

import (
	"bytes"
	"io"
	"math"
	"unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/fixedwidth"
	"github.com/malivvan/binco/wirefmt"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewSliceSerializer returns a slice serializer that encodes elements with elemSerializer.
//
// Length and element validators passed through opts are honoured; there is no
// separate constructor to remember.
func NewSliceSerializer[T any](elemSerializer wirefmt.Serializer[T], opts ...slopts.SetOption[T]) wirefmt.Serializer[[]T] {
	o := slopts.Options[T]{}
	slopts.Apply(opts, &o)

	s := newListCodec(elemSerializer, o)
	if o.LenVl != nil || o.ElemVl != nil {
		return checkedList[T]{listCodec: s, lenVl: o.LenVl, elemVl: o.ElemVl}
	}
	return s
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

func (s listCodec[T]) Skip(r wirefmt.Reader) (n int, err error) {
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

// rawBytes returns the memory of v as bytes, which is the wire form of a fixed
// width little-endian element.
func rawBytes[T any](v []T) []byte {
	if len(v) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(v))), len(v)*int(unsafe.Sizeof(*new(T))))
}

// MarshalSlice writes v, prefixed by its length, to w.
func MarshalSlice[T any](v []T, elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int], w wirefmt.Writer) (n int, err error) {
	n, err = lenSerializer.Marshal(len(v), w)
	if err != nil {
		return
	}
	if _, ok := fixedwidth.Of[T](elemSerializer); ok {
		if len(v) == 0 {
			return n, nil
		}
		var n1 int
		n1, err = w.Write(rawBytes(v))
		n += n1
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
//
// The slice grows as elements arrive; the declared length is only a hint, so a
// stream that announces a huge slice does not pre-allocate it. A length whose
// elements consume no bytes at all is capped by binco.MaxZeroSizeElements.
func UnmarshalSlice[T any](elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int], r wirefmt.Reader) (v []T, n int, err error) {
	length, n, err := lenSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		return nil, n, binco.ErrNegativeLength
	}
	if width, ok := fixedwidth.Of[T](elemSerializer); ok {
		return unmarshalFixedSlice[T](r, n, length, width)
	}
	v = make([]T, 0, hintLength(length))
	var n1 int

	var e T

	for range length {
		e, n1, err = elemSerializer.Unmarshal(r)
		n += n1
		if err != nil {
			return
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

// unmarshalFixedSlice reads length elements of a fixed width.
//
// The bytes are collected as they arrive, so a declared length that the stream
// never delivers costs only what was actually sent.
func unmarshalFixedSlice[T any](r wirefmt.Reader, n, length, width int) (
	v []T, end int, err error) {
	if length < 0 {
		return nil, n, binco.ErrNegativeLength
	}
	if width > 0 && length > math.MaxInt/width {
		return nil, n, binco.ErrTooLargeLength
	}
	total := length * width
	if total == 0 {
		return []T{}, n, nil
	}
	var buf bytes.Buffer
	var n1 int64
	n1, err = io.CopyN(&buf, r, int64(total))
	n += int(n1)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, n, err
	}
	v = make([]T, length)
	copy(rawBytes(v), buf.Bytes())
	return v, n, nil
}

// hintLength caps the initial capacity of a decoded container.
//
// The elements are appended as they are decoded, so the hint only saves a few
// growth steps; it must not turn a declared length into an allocation.
func hintLength(length int) int {
	const maxHint = 1024
	if length < 0 {
		return 0
	}
	if length > maxHint {
		return maxHint
	}
	return length
}

// SizeSlice reports the encoded size of v.
func SizeSlice[T any](v []T, elemSerializer wirefmt.Serializer[T],
	lenSerializer wirefmt.Serializer[int]) (size int) {
	size = lenSerializer.Size(len(v))
	if width, ok := fixedwidth.Of[T](elemSerializer); ok && len(v) <= math.MaxInt/width {
		return size + len(v)*width
	}
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
		return n, binco.ErrNegativeLength
	}
	if width, ok := fixedwidth.Of[T](elemSerializer); ok {
		if width > 0 && length > math.MaxInt/width {
			return n, binco.ErrTooLargeLength
		}
		var n1 int64
		n1, err = io.CopyN(io.Discard, r, int64(length*width))
		n += int(n1)
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return n, err
	}
	var n1 int
	for range length {
		n1, err = elemSerializer.Skip(r)
		n += n1
		if err != nil {
			return
		}
		if n1 == 0 {
			if length > binco.MaxZeroSizeElements {
				return n, binco.ErrTooLargeLength
			}
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
		return nil, n, binco.ErrNegativeLength
	}
	if lenVl != nil {
		if err = lenVl.Validate(length); err != nil {
			return
		}
	}
	if elemVl == nil {
		if width, ok := fixedwidth.Of[T](elemSerializer); ok {
			return unmarshalFixedSlice[T](r, n, length, width)
		}
	}
	v = make([]T, 0, hintLength(length))
	var n1 int

	var e T

	for range length {
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
