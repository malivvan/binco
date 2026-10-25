package zerocpy

import (
	"reflect"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	arropts "github.com/malivvan/binco/config/array"
	"github.com/malivvan/binco/varnum"
)

// NewArraySerializer returns a serializer for the array type T, whose elements are encoded with the element serializer.
func NewArraySerializer[T, V any](elemSerializer binco.Serializer[V],
	opts ...arropts.SetOption[V],
) (s arrayView[T, V]) {
	var o = arropts.Options[V]{}

	var t = reflect.TypeFor[T]()

	var length = t.Len()

	arropts.Apply(opts, &o)
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt

	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return arrayView[T, V]{
		length:         length,
		elemSerializer: elemSerializer,
		lenSerializer:  lenSerializer,
		lenVl:          lengthLimit(length),
	}
}

// NewValidArraySerializer is NewArraySerializer with element validation.
func NewValidArraySerializer[T, V any](elemSerializer binco.Serializer[V],
	opts ...arropts.SetOption[V],
) arrayView[T, V] {
	var o = arropts.Options[V]{}

	var t = reflect.TypeFor[T]()

	var length = t.Len()

	arropts.Apply(opts, &o)
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt

	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return arrayView[T, V]{
		length:         length,
		elemSerializer: elemSerializer,
		lenSerializer:  lenSerializer,
		lenVl:          lengthLimit(length),
		elemVl:         o.ElemVl,
	}
}

type arrayView[T, V any] struct {
	length         int
	elemSerializer binco.Serializer[V]
	lenSerializer  binco.Serializer[int]
	lenVl          binco.Validator[int]
	elemVl         binco.Validator[V]
}

func (s arrayView[T, V]) Marshal(v T, bs []byte) (n int) {
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	return composite.MarshalSlice(sl, s.elemSerializer, s.lenSerializer, bs)
}

func (s arrayView[T, V]) Unmarshal(bs []byte) (v T, n int, err error) {
	sl, n, err := composite.UnmarshalValidSlice(bs, s.elemSerializer, s.lenSerializer, s.lenVl, s.elemVl)
	if err != nil {
		return
	}

	if len(sl) != s.length {
		err = binco.ErrWrongLength
		return
	}
	if s.length == 0 {
		return
	}
	v = *(*T)(unsafe_mod.Pointer(unsafe_mod.SliceData(sl)))
	return
}

func (s arrayView[T, V]) Size(v T) (size int) {
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	return composite.SizeSlice(sl, s.elemSerializer, s.lenSerializer)
}

func (s arrayView[T, V]) Skip(bs []byte) (n int, err error) {
	return composite.SkipSlice(bs, s.elemSerializer, s.lenSerializer)
}

func lengthLimit(length int) binco.ValidatorFn[int] {
	return func(t int) (err error) {
		if t > length {
			err = binco.ErrTooLargeLength
		}
		return
	}
}
